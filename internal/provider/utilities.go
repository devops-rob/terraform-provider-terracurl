package provider

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/net/http/httpproxy"
)

func sanitizeResponse(response string, fieldsToIgnore []string) (string, error) {
	// Return early if the response is empty.
	if response == "" {
		return "", nil
	}

	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonObj); err != nil {
		// Return the original response if it's not a JSON object.
		fmt.Println("Warning: Response is not a JSON object:", err)
		return response, nil
	}

	// Remove ignored fields.
	for _, field := range fieldsToIgnore {
		delete(jsonObj, field)
	}

	// Convert back to JSON.
	filteredBytes, err := json.Marshal(jsonObj)
	if err != nil {
		return "", fmt.Errorf("failed to serialize filtered JSON: %v", err)
	}

	return string(filteredBytes), nil
}

// setResponseValue writes body to either response or sensitiveResponse based on
// the sensitive flag. The unused attribute is always set to an empty string so
// Terraform does not report it as unknown.
func setResponseValue(sensitive bool, response, sensitiveResponse *types.String, body string) {
	if sensitive {
		*response = types.StringValue("")
		*sensitiveResponse = types.StringValue(body)
	} else {
		*response = types.StringValue(body)
		*sensitiveResponse = types.StringValue("")
	}
}

func responseSensitiveEnabled(value types.Bool) bool {
	if value.IsNull() || value.IsUnknown() {
		return false
	}
	return value.ValueBool()
}

func priorResponseValue(responseSensitive types.Bool, response, sensitiveResponse types.String) string {
	if responseSensitiveEnabled(responseSensitive) {
		return sensitiveResponse.ValueString()
	}
	return response.ValueString()
}

func setResourceResponseValues(data *CurlResourceModel, body string) {
	setResponseValue(
		responseSensitiveEnabled(data.ResponseSensitive),
		&data.Response,
		&data.SensitiveResponse,
		body,
	)
}

func setEphemeralOpenResponse(data *CurlEphemeralModel, body string) {
	setResponseValue(
		responseSensitiveEnabled(data.ResponseSensitive),
		&data.Response,
		&data.SensitiveResponse,
		body,
	)
}

func setEphemeralRenewResponse(data *CurlEphemeralModel, body string) {
	setResponseValue(
		responseSensitiveEnabled(data.ResponseSensitive),
		&data.RenewResponse,
		&data.SensitiveRenewResponse,
		body,
	)
}

func setEphemeralCloseResponse(data *CurlEphemeralModel, body string) {
	setResponseValue(
		responseSensitiveEnabled(data.ResponseSensitive),
		&data.CloseResponse,
		&data.SensitiveCloseResponse,
		body,
	)
}

func setDataSourceResponseValues(data *CurlDataSourceModel, body string) {
	setResponseValue(
		responseSensitiveEnabled(data.ResponseSensitive),
		&data.Response,
		&data.SensitiveResponse,
		body,
	)
}

func responseCodeChecker(expectedStatusCodes types.List, receivedStatusCode int) bool {
	var responseStatusCodes []string
	for _, v := range expectedStatusCodes.Elements() {
		if strVal, ok := v.(types.String); ok {
			responseStatusCodes = append(responseStatusCodes, strVal.ValueString())
		}
	}

	receivedStatusCodeAsInt := strconv.Itoa(receivedStatusCode)

	for _, v := range responseStatusCodes {
		if v == receivedStatusCodeAsInt {
			return true
		}
	}

	return false
}

type TlsConfig struct {
	CertFile        string
	KeyFile         string
	CaCertFile      string
	CaCertDirectory string
	SkipTlsVerify   bool
}

// defaultTlsConfig returns a default TlsConfig instance.
func defaultTlsConfig() *TlsConfig {
	return &TlsConfig{}
}

// ProviderMeta carries provider-level configuration passed to resources and data sources.
type ProviderMeta struct {
	proxyFunc func(*url.URL) (*url.URL, error)
}

// DefaultProviderMeta returns provider metadata that uses only environment-based proxy settings.
func DefaultProviderMeta() *ProviderMeta {
	return &ProviderMeta{
		proxyFunc: httpproxy.FromEnvironment().ProxyFunc(),
	}
}

// NewProviderMeta builds provider metadata, merging optional provider proxy settings with
// environment variables. Explicitly set provider attributes override environment values.
func NewProviderMeta(httpProxy, httpsProxy, noProxy types.String) *ProviderMeta {
	cfg := httpproxy.FromEnvironment()
	if !httpProxy.IsNull() {
		cfg.HTTPProxy = httpProxy.ValueString()
	}
	if !httpsProxy.IsNull() {
		cfg.HTTPSProxy = httpsProxy.ValueString()
	}
	if !noProxy.IsNull() {
		cfg.NoProxy = noProxy.ValueString()
	}
	return &ProviderMeta{proxyFunc: cfg.ProxyFunc()}
}

func (m *ProviderMeta) requestProxy(req *http.Request) (*url.URL, error) {
	if m == nil || m.proxyFunc == nil {
		return httpproxy.FromEnvironment().ProxyFunc()(req.URL)
	}
	return m.proxyFunc(req.URL)
}

func (m *ProviderMeta) transportProxy() func(*http.Request) (*url.URL, error) {
	return m.requestProxy
}

func cloneTransportWithProxy(proxy func(*http.Request) (*url.URL, error)) http.RoundTripper {
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		transport := base.Clone()
		transport.Proxy = proxy
		return transport
	}

	return &http.Transport{
		Proxy: proxy,
	}
}

// NewHTTPClient returns an HTTP client for the given TLS configuration.
// Pass nil tlsCfg for a non-TLS client. Both paths honor configured proxy settings.
func (m *ProviderMeta) NewHTTPClient(tlsCfg *TlsConfig) (*http.Client, error) {
	if tlsCfg != nil {
		return createTlsClient(tlsCfg, m.transportProxy())
	}

	var transport http.RoundTripper
	if _, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = cloneTransportWithProxy(m.transportProxy())
	} else {
		// Preserve test transports such as httpmock's MockTransport.
		transport = http.DefaultTransport
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

func createTlsClient(cfg *TlsConfig, proxy func(*http.Request) (*url.URL, error)) (*http.Client, error) {
	var certificates []tls.Certificate
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate or key: %v", err)
		}
		certificates = append(certificates, cert)
	}

	// Load CA certificates.
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if cfg.CaCertFile != "" {
		caCert, err := os.ReadFile(cfg.CaCertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert file: %v", err)
		}
		if !rootCAs.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
	}

	// Build TLS configuration.
	tlsConfig := &tls.Config{
		Certificates:       certificates,
		RootCAs:            rootCAs,
		InsecureSkipVerify: cfg.SkipTlsVerify,
	}

	transport, ok := cloneTransportWithProxy(proxy).(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("failed to create TLS transport")
	}
	transport.TLSClientConfig = tlsConfig

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

func convertMap(tfMap types.Map) map[string]string {
	if tfMap.IsNull() || tfMap.IsUnknown() {
		return nil
	}

	goMap := make(map[string]string)
	for k, v := range tfMap.Elements() {
		if strVal, ok := v.(types.String); ok {
			goMap[k] = strVal.ValueString()
		}
	}
	return goMap
}

func convertStringSliceToTFValues(input []string) []attr.Value {
	values := make([]attr.Value, len(input))
	for i, str := range input {
		values[i] = types.StringValue(str)
	}
	return values
}

func convertStringMapToTFValues(input map[string]string) map[string]attr.Value {
	result := make(map[string]attr.Value)
	for k, v := range input {
		result[k] = types.StringValue(v)
	}
	return result
}

func hasValue(s types.String) bool {
	return !s.IsNull() && s.ValueString() != ""
}

const hostHeaderMarkdownSuffix = " Host (case-insensitive) overrides the HTTP Host header sent on the wire, independent of the URL hostname."

func applyRequestHeaders(req *http.Request, headers types.Map) {
	if headers.IsNull() || headers.IsUnknown() {
		return
	}

	for k, v := range headers.Elements() {
		if strVal, ok := v.(types.String); ok {
			if strings.EqualFold(k, "host") {
				req.Host = strVal.ValueString()
				continue
			}
			req.Header.Set(k, strVal.ValueString())
		}
	}
}
