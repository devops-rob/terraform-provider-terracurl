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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccCurlActionBasic tests that an action can be configured to call an HTTP server
func TestAccCurlActionBasic(t *testing.T) {
	var requestCount int32

	// Create test HTTP server using httptest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		if strings.HasSuffix(r.URL.Path, "/404") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := invokeRequestAction(context.Background(), t, server.URL, []string{"200"}); err != nil {
						t.Fatalf("Failed to invoke request action: %v", err)
					}
				},
				Config: testAccCurlActionBasic(server.URL),
				Check: func(_ *terraform.State) error {
					if atomic.LoadInt32(&requestCount) != 1 {
						return fmt.Errorf("expected request count to be 1, got %d", atomic.LoadInt32(&requestCount))
					}
					return nil
				},
			},
			{
				PreConfig: func() {
					if err := invokeRequestAction(context.Background(), t, server.URL+"/404", []string{"200", "404"}); err != nil {
						t.Fatalf("Failed to invoke request action: %v", err)
					}
				},
				Config: testAccCurlActionBasic(server.URL),
				Check: func(_ *terraform.State) error {
					if atomic.LoadInt32(&requestCount) != 2 {
						return fmt.Errorf("expected request count to be 2, got %d", atomic.LoadInt32(&requestCount))
					}
					return nil
				},
			},
		},
	})
}

func testAccCurlActionBasic(url string) string {
	// This isn't strictly needed currently since we construct the config directly, but we need
	// _something_ so may as well use this for now.
	return fmt.Sprintf(`
action "terracurl_request" "test" {
  config {
    url    = "%s"
    method = "GET"
    response_codes = ["200"]
  }
}
`, url)
}

// Step 1: Get the terracurl provider as a ProviderServerWithActions
func providerWithActions(ctx context.Context, t *testing.T) tfprotov6.ProviderServerWithActions { //nolint:staticcheck // SA1019: Working in alpha situation
	t.Helper()

	factories := testAccProtoV6ProviderFactories
	providerFactory, exists := factories["terracurl"]
	if !exists {
		t.Fatal("Terracurl provider factory not found in ProtoV6ProviderFactories")
	}

	providerServer, err := providerFactory()
	if err != nil {
		t.Fatalf("Failed to create provider server: %v", err)
	}

	providerWithActions, ok := providerServer.(tfprotov6.ProviderServerWithActions) //nolint:staticcheck // SA1019: Working in alpha situation
	if !ok {
		t.Fatal("Provider does not implement ProviderServerWithActions")
	}

	schemaResp, err := providerWithActions.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("Failed to get provider schema: %v", err)
	}

	if len(schemaResp.ActionSchemas) == 0 {
		t.Fatal("Expected to find action schemas but didn't find any!")
	}

	configValue, err := tfprotov6.NewDynamicValue(tftypes.Object{}, tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}))
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	configureResp, err := providerWithActions.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{
		TerraformVersion: "1.0.0",
		Config:           &configValue,
	})
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	if len(configureResp.Diagnostics) > 0 {
		var diagMessages []string
		for _, diag := range configureResp.Diagnostics {
			diagMessages = append(diagMessages, fmt.Sprintf("Severity: %s, Summary: %s, Detail: %s", diag.Severity, diag.Summary, diag.Detail))
		}
		t.Fatalf("Provider configuration failed: %v", diagMessages)
	}

	return providerWithActions
}

func buildRequestActionConfig(url string, responseCodes []string) (tftypes.Type, map[string]tftypes.Value) {
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

	config := map[string]tftypes.Value{
		"ca_cert_directory":  tftypes.NewValue(tftypes.String, nil),
		"ca_cert_file":       tftypes.NewValue(tftypes.String, nil),
		"cert_file":          tftypes.NewValue(tftypes.String, nil),
		"headers":            tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{}),
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

// Step 3: Programmatic action invocation
func invokeRequestAction(ctx context.Context, t *testing.T, url string, responseCodes []string) error {
	t.Helper()

	p := providerWithActions(ctx, t)
	configType, configMap := buildRequestActionConfig(url, responseCodes)
	actionTypeName := "terracurl_request"

	testConfig, err := tfprotov6.NewDynamicValue(
		configType,
		tftypes.NewValue(configType, configMap),
	)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	invokeResp, err := p.InvokeAction(ctx, &tfprotov6.InvokeActionRequest{
		ActionType: actionTypeName,
		Config:     &testConfig,
	})
	if err != nil {
		return fmt.Errorf("invoke failed: %w", err)
	}

	// Process events and check for completion
	for event := range invokeResp.Events {
		switch eventType := event.Type.(type) {
		case tfprotov6.ProgressInvokeActionEventType:
			t.Logf("Progress: %s", eventType.Message)
		case tfprotov6.CompletedInvokeActionEventType:
			return nil
		default:
			// Handle any other event types or errors
			t.Logf("Received event type: %T", eventType)
		}
	}

	return nil
}
