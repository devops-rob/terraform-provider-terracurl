package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/http/httpproxy"
)

func TestSanitizeResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		fieldsToIgnore []string
		expected       string
		expectErr      bool
	}{
		{
			name:           "Valid JSON with fields to ignore",
			response:       `{"username": "test", "password": "secret"}`,
			fieldsToIgnore: []string{"password"},
			expected:       `{"username":"test"}`,
		},
		{
			name:           "Invalid JSON should return original response",
			response:       `invalid json`,
			fieldsToIgnore: []string{"password"},
			expected:       `invalid json`,
		},
		{
			name:           "Empty JSON should return empty string",
			response:       "",
			fieldsToIgnore: []string{"password"},
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizeResponse(tt.response, tt.fieldsToIgnore)
			if (err != nil) != tt.expectErr {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResponseCodeChecker(t *testing.T) {
	tests := []struct {
		name     string
		codes    []attr.Value
		input    int
		expected bool
	}{
		{"Value Present", []attr.Value{types.StringValue("200"), types.StringValue("404"), types.StringValue("500")}, 404, true},
		{"Value Absent", []attr.Value{types.StringValue("200"), types.StringValue("500")}, 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listValue types.List
			listValue, _ = types.ListValue(types.StringType, tt.codes)
			result := responseCodeChecker(listValue, tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateTlsClient(t *testing.T) {
	proxy := httpproxy.FromEnvironment().ProxyFunc()
	proxyForRequest := func(req *http.Request) (*url.URL, error) {
		return proxy(req.URL)
	}

	t.Run("Default Config", func(t *testing.T) {
		cfg := defaultTlsConfig()
		client, err := createTlsClient(cfg, proxyForRequest)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if client == nil {
			t.Error("Expected non-nil HTTP client")
		}
	})

	t.Run("Invalid Cert File", func(t *testing.T) {
		cfg := &TlsConfig{CertFile: "nonexistent.pem", KeyFile: "nonexistent-key.pem"}
		client, err := createTlsClient(cfg, proxyForRequest)
		if err == nil {
			t.Error("Expected error for invalid cert file, got nil")
		}
		if client != nil {
			t.Error("Expected nil client on failure")
		}
	})
}

func TestTlsClientRequests(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	cfg := &TlsConfig{SkipTlsVerify: true}
	proxy := httpproxy.FromEnvironment().ProxyFunc()
	client, err := createTlsClient(cfg, func(req *http.Request) (*url.URL, error) {
		return proxy(req.URL)
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("Expected body 'success', got '%s'", body)
	}
}
