package provider

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"
	"os"
	"time"
)

func sanitizeResponse(response string, fieldsToIgnore []string) (string, error) {
	// Return early if the response is empty.
	if response == "" {
		return "", nil
	}

	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonObj); err != nil {
		// Return the original response if it's not a JSON object.
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

func responseCodeChecker(s []string, str string) bool {
	for _, v := range s {
		if v == str {
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

func createTlsClient(cfg *TlsConfig) (*http.Client, error) {
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

	// Create the TLS-enabled HTTP client.
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 30 * time.Second,
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
