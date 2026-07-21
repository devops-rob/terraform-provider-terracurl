package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestApplyRequestHeadersHostOverride(t *testing.T) {
	tests := []struct {
		name       string
		headerKey  string
		headerVal  string
		wantHost   string
		wantHeader string
	}{
		{name: "Host", headerKey: "Host", headerVal: "www.example.com", wantHost: "www.example.com"},
		{name: "host", headerKey: "host", headerVal: "www.example.com", wantHost: "www.example.com"},
		{name: "HOST", headerKey: "HOST", headerVal: "www.example.com", wantHost: "www.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/path", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			headers, diags := types.MapValueFrom(t.Context(), types.StringType, map[string]string{
				tt.headerKey: tt.headerVal,
			})
			if diags.HasError() {
				t.Fatalf("failed to build headers map: %v", diags)
			}

			applyRequestHeaders(req, headers)

			if req.Host != tt.wantHost {
				t.Fatalf("expected request.Host %q, got %q", tt.wantHost, req.Host)
			}
			if got := req.Header.Get("Host"); got != "" {
				t.Fatalf("expected empty Host header map entry, got %q", got)
			}
		})
	}
}

func TestApplyRequestHeadersNonHost(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	headers, diags := types.MapValueFrom(t.Context(), types.StringType, map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	})
	if diags.HasError() {
		t.Fatalf("failed to build headers map: %v", diags)
	}

	applyRequestHeaders(req, headers)

	if req.Header.Get("Authorization") != "Bearer token" {
		t.Fatalf("expected Authorization header to be set")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type header to be set")
	}
	if req.Host != "example.com" {
		t.Fatalf("expected default host from URL, got %q", req.Host)
	}
}

func TestApplyRequestHeadersNullAndUnknown(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	applyRequestHeaders(req, types.MapNull(types.StringType))
	applyRequestHeaders(req, types.MapUnknown(types.StringType))

	if len(req.Header) != 0 {
		t.Fatalf("expected no headers to be applied, got %v", req.Header)
	}
}

func TestApplyRequestHeadersHostWithOtherHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	headers, diags := types.MapValueFrom(t.Context(), types.StringType, map[string]string{
		"Host":          "www.example.com",
		"Authorization": "Bearer token",
	})
	if diags.HasError() {
		t.Fatalf("failed to build headers map: %v", diags)
	}

	applyRequestHeaders(req, headers)

	if req.Host != "www.example.com" {
		t.Fatalf("expected request.Host override, got %q", req.Host)
	}
	if req.Header.Get("Authorization") != "Bearer token" {
		t.Fatalf("expected Authorization header to be set")
	}
}

func TestApplyRequestHeadersSentOnWire(t *testing.T) {
	var receivedHost string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Host
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	headers, diags := types.MapValueFrom(t.Context(), types.StringType, map[string]string{
		"Host": "www.example.com",
	})
	if diags.HasError() {
		t.Fatalf("failed to build headers map: %v", diags)
	}

	applyRequestHeaders(req, headers)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()
	_, _ = io.ReadAll(resp.Body)

	if receivedHost != "www.example.com" {
		t.Fatalf("expected server to receive Host %q, got %q", "www.example.com", receivedHost)
	}
}
