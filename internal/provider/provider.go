package provider

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const TerraformProviderProductUserAgent = "terraform-provider-terracurl"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HTTPClient

func init() {
	Client = &http.Client{}
}

type TLSClient struct {
	client HTTPClient
}

func (tc *TLSClient) Do(req *http.Request) (*http.Response, error) {
	return tc.client.Do(req)
}

func NewTLSClient(certFile, keyFile, caCert, caDir string, insecureSkipVerify bool, useDefaultClient bool) (HTTPClient, error) {
	if useDefaultClient {
		// Directly return a wrapper around http.DefaultClient for testing
		return &TLSClient{client: http.DefaultClient}, nil
	}
	var cert tls.Certificate
	if certFile != "" {
		c, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		cert = c
	}

	var rootCAs *x509.CertPool
	if caCert != "" {
		rootCAs = x509.NewCertPool()
		caCertBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, err
		}
		if !rootCAs.AppendCertsFromPEM(caCertBytes) {
			return nil, errors.New("failed to append CA certificate")
		}
	} else if caDir != "" {
		rootCAs = x509.NewCertPool()
		files, err := ioutil.ReadDir(caDir)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !strings.HasSuffix(file.Name(), ".pem") {
				continue
			}

			caCert, err := ioutil.ReadFile(filepath.Join(caDir, file.Name()))
			if err != nil {
				return nil, err
			}

			if !rootCAs.AppendCertsFromPEM(caCert) {
				return nil, errors.New("failed to append CA certificate")
			}
		}
	}

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return http.ProxyFromEnvironment(req)
		},
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            rootCAs,
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	return &TLSClient{&http.Client{Transport: tr}}, nil
}

func setClient(certFile, keyFile, caCert, caDir string, insecureSkipVerify bool) error {
	// Determine whether to use http.DefaultClient based on an environment variable or a test flag.
	useDefaultClient := os.Getenv("USE_DEFAULT_CLIENT_FOR_TESTS") == "true"

	tlsClient, err := NewTLSClient(certFile, keyFile, caCert, caDir, insecureSkipVerify, useDefaultClient)
	if err != nil {
		return err
	}

	Client = tlsClient
	return nil
}

func Provider() *schema.Provider {
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"terracurl_request": dataSourceCurlRequest(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"terracurl_request": resourceCurl(),
		},
	}

	return provider
}

//type apiClient struct {
//}
//
//func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
//	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
//
//		return &apiClient{}, nil
//	}
//}
