package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewProviderMetaProxyResolution(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://env-proxy.example:8080")
	t.Setenv("HTTPS_PROXY", "http://env-https-proxy.example:8443")
	t.Setenv("NO_PROXY", "")

	t.Run("uses environment when provider values unset", func(t *testing.T) {
		meta := NewProviderMeta(types.StringNull(), types.StringNull(), types.StringNull())
		proxyURL, err := meta.requestProxy(&http.Request{URL: mustParseURL(t, "http://example.com/path")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proxyURL.String() != "http://env-proxy.example:8080" {
			t.Fatalf("expected env HTTP proxy, got %q", proxyURL)
		}
	})

	t.Run("provider values override environment", func(t *testing.T) {
		meta := NewProviderMeta(
			types.StringValue("http://provider-proxy.example:3128"),
			types.StringNull(),
			types.StringNull(),
		)
		proxyURL, err := meta.requestProxy(&http.Request{URL: mustParseURL(t, "http://example.com/path")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proxyURL.String() != "http://provider-proxy.example:3128" {
			t.Fatalf("expected provider HTTP proxy, got %q", proxyURL)
		}
	})

	t.Run("explicit empty provider value disables proxy", func(t *testing.T) {
		meta := NewProviderMeta(types.StringValue(""), types.StringNull(), types.StringNull())
		proxyURL, err := meta.requestProxy(&http.Request{URL: mustParseURL(t, "http://example.com/path")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proxyURL != nil {
			t.Fatalf("expected no proxy, got %q", proxyURL)
		}
	})

	t.Run("uses HTTPS proxy for HTTPS requests", func(t *testing.T) {
		meta := NewProviderMeta(types.StringNull(), types.StringNull(), types.StringNull())
		proxyURL, err := meta.requestProxy(&http.Request{URL: mustParseURL(t, "https://example.com/path")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proxyURL.String() != "http://env-https-proxy.example:8443" {
			t.Fatalf("expected env HTTPS proxy, got %q", proxyURL)
		}
	})

	t.Run("respects NO_PROXY", func(t *testing.T) {
		meta := NewProviderMeta(types.StringNull(), types.StringNull(), types.StringValue("localhost,127.0.0.1"))
		proxyURL, err := meta.requestProxy(&http.Request{URL: mustParseURL(t, "http://127.0.0.1/path")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proxyURL != nil {
			t.Fatalf("expected direct connection for NO_PROXY host, got %q", proxyURL)
		}
	})
}

func TestProviderMetaNewHTTPClientTransportProxy(t *testing.T) {
	meta := NewProviderMeta(
		types.StringValue("http://proxy.example:8080"),
		types.StringNull(),
		types.StringNull(),
	)

	t.Run("non-TLS client configures proxy", func(t *testing.T) {
		client, err := meta.NewHTTPClient(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Skip("non-Transport client in use (likely httpmock); skipping proxy transport assertion")
		}
		if transport.Proxy == nil {
			t.Fatal("expected transport proxy to be configured")
		}
	})

	t.Run("TLS client configures proxy", func(t *testing.T) {
		client, err := meta.NewHTTPClient(&TlsConfig{SkipTlsVerify: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected *http.Transport, got %T", client.Transport)
		}
		if transport.Proxy == nil {
			t.Fatal("expected transport proxy to be configured")
		}
	})
}

func TestHTTPClientRoutesHTTPThroughProxy(t *testing.T) {
	var proxyUsed bool
	var requestedHost string
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyUsed = true
		requestedHost = r.URL.Host
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("via-proxy"))
	}))
	defer proxyServer.Close()

	meta := NewProviderMeta(
		types.StringValue(proxyServer.URL),
		types.StringNull(),
		types.StringValue(""),
	)
	client, err := meta.NewHTTPClient(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.Get("http://example.com/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if !proxyUsed {
		t.Fatal("expected request to route through proxy")
	}
	if requestedHost != "example.com" {
		t.Fatalf("expected proxied host example.com, got %q", requestedHost)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "via-proxy" {
		t.Fatalf("expected proxied response, got %q", body)
	}
}

func TestHTTPClientRoutesHTTPSThroughProxy(t *testing.T) {
	connectReceived := make(chan string, 1)
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "expected CONNECT", http.StatusBadRequest)
			return
		}
		connectReceived <- r.Host
		http.Error(w, "proxy reached", http.StatusBadGateway)
	}))
	defer proxyServer.Close()

	meta := NewProviderMeta(
		types.StringNull(),
		types.StringValue(proxyServer.URL),
		types.StringValue(""),
	)
	client, err := meta.NewHTTPClient(&TlsConfig{SkipTlsVerify: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := client.Get("https://example.com/test")
		errCh <- err
	}()

	var connectHost string
	select {
	case connectHost = <-connectReceived:
	case err := <-errCh:
		t.Fatalf("expected CONNECT through proxy before client returned, got client error: %v", err)
	}

	if connectHost != "example.com:443" {
		t.Fatalf("expected CONNECT to example.com:443, got %q", connectHost)
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", rawURL, err)
	}
	return parsed
}
