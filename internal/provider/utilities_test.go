package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
		codes    []string
		input    string
		expected bool
	}{
		{"Value Present", []string{"200", "404", "500"}, "404", true},
		{"Value Absent", []string{"200", "500"}, "404", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := responseCodeChecker(tt.codes, tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateTlsClient(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		cfg := defaultTlsConfig()
		client, err := createTlsClient(cfg)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if client == nil {
			t.Error("Expected non-nil HTTP client")
		}
	})

	t.Run("Invalid Cert File", func(t *testing.T) {
		cfg := &TlsConfig{CertFile: "nonexistent.pem", KeyFile: "nonexistent-key.pem"}
		client, err := createTlsClient(cfg)
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
	client, err := createTlsClient(cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("Expected body 'success', got '%s'", body)
	}
}
