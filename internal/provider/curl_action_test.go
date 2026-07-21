package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestInvokeCurlActionBasic(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		if strings.HasSuffix(r.URL.Path, "/404") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	if err := invokeRequestAction(context.Background(), t, server.URL, []string{"200"}, nil); err != nil {
		t.Fatalf("failed to invoke request action: %v", err)
	}
	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Fatalf("expected request count 1, got %d", got)
	}

	if err := invokeRequestAction(context.Background(), t, server.URL+"/404", []string{"200", "404"}, nil); err != nil {
		t.Fatalf("failed to invoke request action for alternate status: %v", err)
	}
	if got := atomic.LoadInt32(&requestCount); got != 2 {
		t.Fatalf("expected request count 2, got %d", got)
	}
}

func TestInvokeCurlActionHostHeader(t *testing.T) {
	const wantHost = "www.example.com"
	var gotHost atomic.Value

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost.Store(r.Host)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := invokeRequestAction(context.Background(), t, server.URL, []string{"200"}, map[string]string{
		"Host": wantHost,
	}); err != nil {
		t.Fatalf("failed to invoke request action: %v", err)
	}

	host, ok := gotHost.Load().(string)
	if !ok || host != wantHost {
		t.Fatalf("expected Host %q on the wire, got %q", wantHost, host)
	}
}

func providerWithActions(ctx context.Context, t *testing.T) tfprotov6.ProviderServerWithActions { //nolint:staticcheck // SA1019: actions RPCs still exposed here
	t.Helper()

	providerFactory, exists := testAccProtoV6ProviderFactories["terracurl"]
	if !exists {
		t.Fatal("terracurl provider factory not found")
	}

	providerServer, err := providerFactory()
	if err != nil {
		t.Fatalf("failed to create provider server: %v", err)
	}

	providerWithActions, ok := providerServer.(tfprotov6.ProviderServerWithActions) //nolint:staticcheck // SA1019: actions RPCs still exposed here
	if !ok {
		t.Fatal("provider does not implement ProviderServerWithActions")
	}

	schemaResp, err := providerWithActions.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("failed to get provider schema: %v", err)
	}
	if len(schemaResp.ActionSchemas) == 0 {
		t.Fatal("expected action schemas but found none")
	}

	providerConfigType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"http_proxy":  tftypes.String,
			"https_proxy": tftypes.String,
			"no_proxy":    tftypes.String,
		},
		OptionalAttributes: map[string]struct{}{
			"http_proxy":  {},
			"https_proxy": {},
			"no_proxy":    {},
		},
	}
	providerConfigValue := tftypes.NewValue(providerConfigType, map[string]tftypes.Value{
		"http_proxy":  tftypes.NewValue(tftypes.String, nil),
		"https_proxy": tftypes.NewValue(tftypes.String, nil),
		"no_proxy":    tftypes.NewValue(tftypes.String, nil),
	})
	configValue, err := tfprotov6.NewDynamicValue(providerConfigType, providerConfigValue)
	if err != nil {
		t.Fatalf("failed to build provider config: %v", err)
	}

	configureResp, err := providerWithActions.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{
		TerraformVersion: "1.14.0",
		Config:           &configValue,
	})
	if err != nil {
		t.Fatalf("failed to configure provider: %v", err)
	}
	if len(configureResp.Diagnostics) > 0 {
		t.Fatalf("provider configuration failed: %v", configureResp.Diagnostics)
	}

	return providerWithActions
}

func buildRequestActionConfig(url string, responseCodes []string, headers map[string]string) (tftypes.Type, map[string]tftypes.Value) {
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"ca_cert_directory":  tftypes.String,
			"ca_cert_file":       tftypes.String,
			"cert_file":          tftypes.String,
			"headers":            tftypes.Map{ElementType: tftypes.String},
			"key_file":           tftypes.String,
			"max_retry":          tftypes.Number,
			"method":             tftypes.String,
			"request_body":       tftypes.String,
			"request_parameters": tftypes.Map{ElementType: tftypes.String},
			"response_codes":     tftypes.List{ElementType: tftypes.String},
			"retry_interval":     tftypes.Number,
			"skip_tls_verify":    tftypes.Bool,
			"timeout":            tftypes.Number,
			"url":                tftypes.String,
		},
		OptionalAttributes: map[string]struct{}{
			"ca_cert_directory":  {},
			"ca_cert_file":       {},
			"cert_file":          {},
			"headers":            {},
			"key_file":           {},
			"max_retry":          {},
			"request_body":       {},
			"request_parameters": {},
			"retry_interval":     {},
			"skip_tls_verify":    {},
			"timeout":            {},
		},
	}

	responseCodesList := make([]tftypes.Value, 0, len(responseCodes))
	for _, code := range responseCodes {
		responseCodesList = append(responseCodesList, tftypes.NewValue(tftypes.String, code))
	}

	headerValues := make(map[string]tftypes.Value, len(headers))
	for key, value := range headers {
		headerValues[key] = tftypes.NewValue(tftypes.String, value)
	}

	config := map[string]tftypes.Value{
		"ca_cert_directory":  tftypes.NewValue(tftypes.String, nil),
		"ca_cert_file":       tftypes.NewValue(tftypes.String, nil),
		"cert_file":          tftypes.NewValue(tftypes.String, nil),
		"headers":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, headerValues),
		"key_file":           tftypes.NewValue(tftypes.String, nil),
		"max_retry":          tftypes.NewValue(tftypes.Number, nil),
		"method":             tftypes.NewValue(tftypes.String, "GET"),
		"request_body":       tftypes.NewValue(tftypes.String, nil),
		"request_parameters": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{}),
		"response_codes":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, responseCodesList),
		"retry_interval":     tftypes.NewValue(tftypes.Number, nil),
		"skip_tls_verify":    tftypes.NewValue(tftypes.Bool, nil),
		"timeout":            tftypes.NewValue(tftypes.Number, nil),
		"url":                tftypes.NewValue(tftypes.String, url),
	}

	return configType, config
}

func invokeRequestAction(ctx context.Context, t *testing.T, url string, responseCodes []string, headers map[string]string) error {
	t.Helper()

	p := providerWithActions(ctx, t)
	configType, configMap := buildRequestActionConfig(url, responseCodes, headers)

	testConfig, err := tfprotov6.NewDynamicValue(
		configType,
		tftypes.NewValue(configType, configMap),
	)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	invokeResp, err := p.InvokeAction(ctx, &tfprotov6.InvokeActionRequest{
		ActionType: "terracurl_request",
		Config:     &testConfig,
	})
	if err != nil {
		return fmt.Errorf("invoke failed: %w", err)
	}

	for event := range invokeResp.Events {
		switch eventType := event.Type.(type) {
		case tfprotov6.ProgressInvokeActionEventType:
			t.Logf("progress: %s", eventType.Message)
		case tfprotov6.CompletedInvokeActionEventType:
			if len(eventType.Diagnostics) > 0 {
				return fmt.Errorf("action completed with diagnostics: %v", eventType.Diagnostics)
			}
			return nil
		default:
			t.Logf("received event type: %T", eventType)
		}
	}

	return fmt.Errorf("invoke finished without completion event")
}
