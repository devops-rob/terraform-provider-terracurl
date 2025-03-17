package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/ephemeralvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strconv"
	"time"
)

var _ provider.ProviderWithEphemeralResources = (*TerraCurlProvider)(nil)
var _ ephemeral.EphemeralResourceWithRenew = (*EphemeralCurlResource)(nil)
var _ ephemeral.EphemeralResourceWithClose = (*EphemeralCurlResource)(nil)

type EphemeralCurlResource struct {
	client *http.Client
}

func (e *EphemeralCurlResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

// EphemeralResources With the provider.ProviderWithEphemeralResources implementation.
func (p *TerraCurlProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewCurlEphemeralResource,
	}
}

// NewCurlEphemeralResource With the ephemeral.EphemeralResource implementation.
func NewCurlEphemeralResource() ephemeral.EphemeralResource {
	return &EphemeralCurlResource{}
}

type CurlEphemeralModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Url               types.String `tfsdk:"url"`
	Method            types.String `tfsdk:"method"`
	RequestBody       types.String `tfsdk:"request_body"`
	Headers           types.Map    `tfsdk:"headers"`
	RequestParameters types.Map    `tfsdk:"request_parameters"`
	RequestUrlString  types.String `tfsdk:"request_url_string"`
	CertFile          types.String `tfsdk:"cert_file"`
	KeyFile           types.String `tfsdk:"key_file"`
	CaCertFile        types.String `tfsdk:"ca_cert_file"`
	CaCertDirectory   types.String `tfsdk:"ca_cert_directory"`
	SkipTlsVerify     types.Bool   `tfsdk:"skip_tls_verify"`
	RetryInterval     types.Int64  `tfsdk:"retry_interval"`
	MaxRetry          types.Int64  `tfsdk:"max_retry"`
	Timeout           types.Int64  `tfsdk:"timeout"`
	Response          types.String `tfsdk:"response"`
	ResponseCodes     types.List   `tfsdk:"response_codes"`
	StatusCode        types.String `tfsdk:"status_code"`
	SkipRenew         types.Bool   `tfsdk:"skip_renew"`

	RenewInterval          types.Int64  `tfsdk:"renew_interval"`
	RenewUrl               types.String `tfsdk:"renew_url"`
	RenewMethod            types.String `tfsdk:"renew_method"`
	RenewRequestBody       types.String `tfsdk:"renew_request_body"`
	RenewHeaders           types.Map    `tfsdk:"renew_headers"`
	RenewRequestParameters types.Map    `tfsdk:"renew_request_parameters"`
	RenewRequestUrlString  types.String `tfsdk:"renew_request_url_string"`
	RenewCertFile          types.String `tfsdk:"renew_cert_file"`
	RenewKeyFile           types.String `tfsdk:"renew_key_file"`
	RenewCaCertFile        types.String `tfsdk:"renew_ca_cert_file"`
	RenewCaCertDirectory   types.String `tfsdk:"renew_ca_cert_directory"`
	RenewSkipTlsVerify     types.Bool   `tfsdk:"renew_skip_tls_verify"`
	RenewRetryInterval     types.Int64  `tfsdk:"renew_retry_interval"`
	RenewMaxRetry          types.Int64  `tfsdk:"renew_max_retry"`
	RenewTimeout           types.Int64  `tfsdk:"renew_timeout"`
	RenewResponse          types.String `tfsdk:"renew_response"`
	RenewResponseCodes     types.List   `tfsdk:"renew_response_codes"`

	SkipClose types.Bool `tfsdk:"skip_close"`

	CloseUrl               types.String `tfsdk:"close_url"`
	CloseMethod            types.String `tfsdk:"close_method"`
	CloseRequestBody       types.String `tfsdk:"close_request_body"`
	CloseHeaders           types.Map    `tfsdk:"close_headers"`
	CloseRequestParameters types.Map    `tfsdk:"close_request_parameters"`
	CloseRequestUrlString  types.String `tfsdk:"close_request_url_string"`
	CloseCertFile          types.String `tfsdk:"close_cert_file"`
	CloseKeyFile           types.String `tfsdk:"close_key_file"`
	CloseCaCertFile        types.String `tfsdk:"close_ca_cert_file"`
	CloseCaCertDirectory   types.String `tfsdk:"close_ca_cert_directory"`
	CloseSkipTlsVerify     types.Bool   `tfsdk:"close_skip_tls_verify"`
	CloseRetryInterval     types.Int64  `tfsdk:"close_retry_interval"`
	CloseMaxRetry          types.Int64  `tfsdk:"close_max_retry"`
	CloseTimeout           types.Int64  `tfsdk:"close_timeout"`
	CloseResponse          types.String `tfsdk:"close_response"`
	CloseResponseCodes     types.List   `tfsdk:"close_response_codes"`
}

func (e *EphemeralCurlResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TerraCurl request ephemeral resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Friendly name for this API call",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
			},
			"url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Api endpoint to call",
			},
			"method": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "HTTP method to use in the API call",
			},
			"request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A request body to attach to the API call",
			},
			"headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the API call",
			},
			"request_parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of parameters to attach to the API call",
			},
			"request_url_string": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Request URL includes parameters if request specified",
			},
			"cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded certificate to present to the server",
			},
			"key_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued",
			},
			"ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that will be used to validate the certificate presented by the server",
			},
			"ca_cert_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server",
			},
			"skip_tls_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set this to true to disable verification of the server's TLS certificate",
			},
			"retry_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interval between each attempt",
			},
			"max_retry": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tries until it is marked as failed",
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time in seconds before each request times out. Defaults to 10",
			},
			"response": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON response received from request",
			},
			"response_codes": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "A list of expected response codes",
				ElementType:         types.StringType,
			},
			"status_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Response status code received from request",
			},
			"skip_renew": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set to true to skip renewing ephemeral resources",
			},

			"renew_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interval in seconds to renew this resource.",
			},
			"renew_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Api endpoint to call",
			},
			"renew_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP method to use in the API call",
			},
			"renew_request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A request body to attach to the API call",
			},
			"renew_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the API call",
			},
			"renew_request_parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of parameters to attach to the API call",
			},
			"renew_request_url_string": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Request URL includes parameters if request specified",
			},
			"renew_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded certificate to present to the server",
			},
			"renew_key_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued",
			},
			"renew_ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that will be used to validate the certificate presented by the server",
			},
			"renew_ca_cert_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server",
			},
			"renew_skip_tls_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set this to true to disable verification of the server's TLS certificate",
			},
			"renew_retry_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interval between each attempt",
			},
			"renew_max_retry": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tries until it is marked as failed",
			},
			"renew_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time in seconds before each request times out. Defaults to 10",
			},
			"renew_response": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON response received from request",
			},
			"renew_response_codes": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "A list of expected response codes",
				ElementType:         types.StringType,
			},
			"skip_close": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set to true if there are no api calls to make to clean up the ephemeral resource on the target platform.",
			},

			"close_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Api endpoint to call",
			},
			"close_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP method to use in the API call",
			},
			"close_request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A request body to attach to the API call",
			},
			"close_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the API call",
			},
			"close_request_parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of parameters to attach to the API call",
			},
			"close_request_url_string": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Request URL includes parameters if request specified",
			},
			"close_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded certificate to present to the server",
			},
			"close_key_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued",
			},
			"close_ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that will be used to validate the certificate presented by the server",
			},
			"close_ca_cert_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server",
			},
			"close_skip_tls_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set this to true to disable verification of the server's TLS certificate",
			},
			"close_retry_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interval between each attempt",
			},
			"close_max_retry": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tries until it is marked as failed",
			},
			"close_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time in seconds before each request times out. Defaults to 10",
			},
			"close_response": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "JSON response received from request",
			},
			"close_response_codes": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "A list of expected response codes",
				ElementType:         types.StringType,
			},
		},
	}
}

func (e *EphemeralCurlResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	tflog.Debug(ctx, "Running open()")
	var data CurlEphemeralModel

	// Read Terraform config data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.SkipRenew.IsNull() && !data.SkipRenew.ValueBool() {
		if data.RenewUrl.IsNull() || data.RenewMethod.IsNull() || data.RenewResponseCodes.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"If `skip_renew` is set to `false`, `renew_url`, `renew_method`, and `renew_response_codes` must be provided.",
			)
			return
		}
	}

	if !data.SkipClose.IsNull() && !data.SkipClose.ValueBool() {
		if data.CloseUrl.IsNull() || data.CloseMethod.IsNull() || data.CloseResponseCodes.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"If `skip_close` is set to `false`, `close_url`, `close_method`, and `close_response_codes` must be provided.",
			)
			return
		}
	}

	data.Id = types.StringValue(data.Name.ValueString())

	// useTLS is used to decide.
	useTLS := !data.CertFile.IsNull() || !data.KeyFile.IsNull() || !data.CaCertFile.IsNull() || !data.CaCertDirectory.IsNull()

	var client *http.Client
	var err error

	if useTLS {
		tflog.Debug(ctx, "Creating TLS enabled client")
		// Build TLS Config.
		tlsConfig := &TlsConfig{
			CertFile:        data.CertFile.ValueString(),
			KeyFile:         data.KeyFile.ValueString(),
			CaCertFile:      data.CaCertFile.ValueString(),
			CaCertDirectory: data.CaCertDirectory.ValueString(),
			SkipTlsVerify:   data.SkipTlsVerify.ValueBool(),
		}

		// Validate TLS settings.
		if tlsConfig.CertFile != "" && tlsConfig.KeyFile == "" {
			resp.Diagnostics.AddError("Validation Error", "`key_file` must be set if `cert_file` is set.")
			return
		}

		// Create TLS-enabled client.
		client, err = createTlsClient(tlsConfig)
		if err != nil {
			resp.Diagnostics.AddError("TLS Client Creation Failed", err.Error())
			return
		}

	} else {
		// Use default non-TLS client.
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	reqBody := []byte(data.RequestBody.ValueString())
	request, err := http.NewRequest(data.Method.ValueString(), data.Url.ValueString(), bytes.NewBuffer(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return
	}

	// Add headers.
	if !data.Headers.IsNull() && !data.Headers.IsUnknown() {
		for k, v := range data.Headers.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	// Add query parameters.
	if !data.RequestParameters.IsNull() && !data.RequestParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.RequestParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}
	data.RequestUrlString = types.StringValue(request.URL.String())

	timeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	var body []byte
	var bodyString string
	var statusCode int
	retryCount := 0

	for {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		response, err := client.Do(request.WithContext(ctxWithTimeout))
		if err != nil {
			if retryCount < int(data.MaxRetry.ValueInt64()) {
				retryCount++
				time.Sleep(time.Duration(data.RetryInterval.ValueInt64()) * time.Second)
				continue
			}
			resp.Diagnostics.AddError("Request Failed", err.Error())
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(response.Body)
		body, _ = io.ReadAll(response.Body)
		statusCode = response.StatusCode

		bodyString = string(body)
		if bodyString == "" {
			bodyString = "{}"
		}

		var responseCodes []string
		for _, v := range data.ResponseCodes.Elements() {
			if strVal, ok := v.(types.String); ok {
				responseCodes = append(responseCodes, strVal.ValueString())
			}
		}

		if responseCodeChecker(responseCodes, strconv.Itoa(statusCode)) {
			break
		}

		if retryCount < int(data.MaxRetry.ValueInt64()) {
			retryCount++
			time.Sleep(time.Duration(data.RetryInterval.ValueInt64()) * time.Second)
		} else {
			resp.Diagnostics.AddError("Unexpected Response Code", fmt.Sprintf("Received status code: %d", statusCode))
			return
		}
	}

	data.RequestUrlString = types.StringValue(request.URL.String())
	data.Response = types.StringValue(bodyString)
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	tflog.Debug(ctx, fmt.Sprintf("renew parameters in Open() is set to %v", convertMap(data.RenewRequestParameters)))

	if !data.SkipRenew.ValueBool() {
		renewDuration := time.Duration(data.RenewInterval.ValueInt64()) * time.Second
		resp.RenewAt = time.Now().Add(renewDuration)
		tflog.Debug(ctx, fmt.Sprintf("Setting RenewAt to: %s (in %d seconds)", resp.RenewAt, renewDuration/time.Second))
	}

	privateData := map[string]interface{}{
		"Id":                data.Id.ValueString(),
		"Name":              data.Name.ValueString(),
		"Url":               data.Url.ValueString(),
		"Method":            data.Method.ValueString(),
		"RequestBody":       data.RequestBody.ValueString(),
		"Headers":           convertMap(data.Headers),
		"RequestParameters": convertMap(data.RequestParameters),
		"RequestUrlString":  data.RequestUrlString.ValueString(),
		"CertFile":          data.CertFile.ValueString(),
		"KeyFile":           data.KeyFile.ValueString(),
		"CaCertFile":        data.CaCertFile.ValueString(),
		"CaCertDirectory":   data.CaCertDirectory.ValueString(),
		"SkipTlsVerify":     data.SkipTlsVerify.ValueBool(),
		"RetryInterval":     data.RetryInterval.ValueInt64(),
		"MaxRetry":          data.MaxRetry.ValueInt64(),
		"Timeout":           data.Timeout.ValueInt64(),
		"StatusCode":        data.StatusCode.ValueString(),
		"Response_codes":    data.ResponseCodes.Elements(),
		"Response":          data.Response.ValueString(),

		"SkipRenew":              data.SkipRenew.ValueBool(),
		"RenewUrl":               data.RenewUrl.ValueString(),
		"RenewMethod":            data.RenewMethod.ValueString(),
		"RenewHeaders":           convertMap(data.RenewHeaders),
		"RenewRequestParameters": convertMap(data.RenewRequestParameters),
		"RenewRequestBody":       data.RenewRequestBody.ValueString(),
		"RenewRequestUrlString":  data.RenewRequestUrlString.ValueString(),
		"RenewCertFile":          data.RenewCertFile.ValueString(),
		"RenewKeyFile":           data.RenewKeyFile.ValueString(),
		"RenewCaCertFile":        data.RenewCaCertFile.ValueString(),
		"RenewCaCertDirectory":   data.RenewCaCertDirectory.ValueString(),
		"RenewSkipTlsVerify":     data.RenewSkipTlsVerify.ValueBool(),
		"RenewRetryInterval":     data.RenewRetryInterval.ValueInt64(),
		"RenewMaxRetry":          data.RenewMaxRetry.ValueInt64(),
		"RenewTimeout":           data.RenewTimeout.ValueInt64(),
		"RenewResponse_codes":    data.RenewResponseCodes.Elements(),
		"RenewResponse":          data.RenewResponse.ValueString(),
		"RenewInterval":          data.RenewInterval.ValueInt64(),

		"SkipClose":              data.SkipClose.ValueBool(),
		"CloseUrl":               data.CloseUrl.ValueString(),
		"CloseMethod":            data.CloseMethod.ValueString(),
		"CloseHeaders":           convertMap(data.CloseHeaders),
		"CloseRequestParameters": convertMap(data.CloseRequestParameters),
		"CloseRequestBody":       data.CloseRequestBody.ValueString(),
		"CloseRequestUrlString":  data.CloseRequestUrlString.ValueString(),
		"CloseCertFile":          data.CloseCertFile.ValueString(),
		"CloseKeyFile":           data.CloseKeyFile.ValueString(),
		"CloseCaCertFile":        data.CloseCaCertFile.ValueString(),
		"CloseCaCertDirectory":   data.CloseCaCertDirectory.ValueString(),
		"CloseSkipTlsVerify":     data.CloseSkipTlsVerify.ValueBool(),
		"CloseRetryInterval":     data.CloseRetryInterval.ValueInt64(),
		"CloseMaxRetry":          data.CloseMaxRetry.ValueInt64(),
		"CloseTimeout":           data.CloseTimeout.ValueInt64(),
		"CloseResponseCodes":     data.CloseResponseCodes.Elements(),
		"CloseResponse":          data.CloseResponse.ValueString(),
	}

	var closeResponseCodesList []string
	for _, v := range data.CloseResponseCodes.Elements() {
		if strVal, ok := v.(types.String); ok {
			closeResponseCodesList = append(closeResponseCodesList, strVal.ValueString())
		}
	}

	var renewResponseCodesList []string
	for _, v := range data.RenewResponseCodes.Elements() {
		if strVal, ok := v.(types.String); ok {
			renewResponseCodesList = append(renewResponseCodesList, strVal.ValueString())
		}
	}

	privateData["CloseResponseCodes"] = closeResponseCodesList
	privateData["RenewResponseCodes"] = renewResponseCodesList

	privateBytes, diags := json.Marshal(privateData)

	if diags != nil {
		resp.Diagnostics.AddError("Error marshaling data", fmt.Sprintf("%s", diags))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Storing privateData: %s", string(privateBytes)))

	resp.Private.SetKey(ctx, "response", privateBytes)
	resp.Result.Set(ctx, privateData)

	// Save data into ephemeral result data.
	resp.Diagnostics.Append(resp.Result.Set(ctx, data)...)
}

func (e *EphemeralCurlResource) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	tflog.Debug(ctx, "Running Renew()")
	privateBytes, diags := req.Private.GetKey(ctx, "response")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Error getting private data at renew step")
		return
	}

	var privateMap map[string]interface{}
	err := json.Unmarshal(privateBytes, &privateMap)
	if err != nil {
		resp.Diagnostics.AddError("Error unmarshaling response", fmt.Sprintf("%s", err))
		return
	}

	renewMethod, ok := privateMap["RenewMethod"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "RenewMethod is not a string")
		return
	}

	renewUrl, ok := privateMap["RenewUrl"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "RenewUrl is not a string")
		return
	}

	skipRenew, ok := privateMap["SkipRenew"].(bool)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "SkipRenew is not a boolean")
		return
	}

	var renewResponseCodes []string

	// Check if the key exists AND is not nil before asserting type.
	if rawList, exists := privateMap["RenewResponseCodes"]; exists && rawList != nil {
		switch v := rawList.(type) {
		case []interface{}:
			for _, item := range v {
				if strVal, ok := item.(string); ok {
					renewResponseCodes = append(renewResponseCodes, strVal)
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", "RenewResponseCodes contains non-string elements")
					return
				}
			}
		case []string:
			renewResponseCodes = v
		default:
			resp.Diagnostics.AddError(
				"Type Assertion Error",
				fmt.Sprintf("RenewResponseCodes has unexpected type: %T", rawList),
			)
			return
		}
	} else {
		// If attribute is missing, default to an empty slice without throwing an error.
		renewResponseCodes = []string{}
	}

	// Convert to Terraform ListValue.
	renewResponseCodesList, diags := types.ListValue(
		types.StringType,
		convertStringSliceToTFValues(renewResponseCodes),
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	renewRequestBody, ok := privateMap["RenewRequestBody"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "RenewRequestBody is not a string")
		return
	}

	renewHeaders := make(map[string]string) // Default to empty map

	if rawRenewHeaders, exists := privateMap["RenewHeaders"]; exists && rawRenewHeaders != nil {
		if rawHeaders, ok := rawRenewHeaders.(map[string]interface{}); ok {
			for key, value := range rawHeaders {
				if strValue, ok := value.(string); ok {
					renewHeaders[key] = strValue
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", fmt.Sprintf("RenewHeaders[%s] is not a string", key))
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("Type Assertion Error", "RenewHeaders is not a map[string]interface{}")
			return
		}
	}

	// Convert to Terraform types.Map.
	renewHeadersTF, diags := types.MapValue(types.StringType, convertStringMapToTFValues(renewHeaders))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	renewParameters := make(map[string]string) // Default to empty map
	tflog.Debug(ctx, fmt.Sprintf("Private Map renew parameters: %v", privateMap["RenewRequestParameters"]))
	if rawRenewParameters, exists := privateMap["RenewRequestParameters"]; exists && rawRenewParameters != nil {
		if rawParameters, ok := rawRenewParameters.(map[string]interface{}); ok {
			for key, value := range rawParameters {
				if strValue, ok := value.(string); ok {
					tflog.Debug(ctx, fmt.Sprintf("Adding %s to renew request parameters\n", renewParameters[key]))
					renewParameters[key] = strValue
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", fmt.Sprintf("RenewRequestParameters[%s] is not a string", key))
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("Type Assertion Error", "RenewRequestParameters is not a map[string]interface{}")
			return
		}
	}

	// Convert to Terraform types.Map.
	renewParametersTF, diags := types.MapValue(types.StringType, convertStringMapToTFValues(renewParameters))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	renewRequestUrlString, ok := privateMap["RenewRequestUrlString"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewRequestUrlString is not a string")
		return
	}

	renewCertFile, ok := privateMap["RenewCertFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewCertFile is not a string")
		return
	}

	renewKeyFile, ok := privateMap["RenewKeyFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewKeyFile is not a string")
		return
	}

	renewCaCertFile, ok := privateMap["RenewCaCertFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewCaCertFile is not a string")
		return
	}

	renewCaCertDirectory, ok := privateMap["RenewCaCertDirectory"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewCaCertDirectory is not a string")
		return
	}

	renewSkipTlsVerify, ok := privateMap["RenewSkipTlsVerify"].(bool)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewSkipTlsVerify is not a bool")
		return
	}

	renewRetryIntervalFloat, ok := privateMap["RenewRetryInterval"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "RenewRetryInterval is not a float64 (expected int64)")
		return
	}
	renewRetryInterval := int64(renewRetryIntervalFloat)

	renewMaxRetryFloat, ok := privateMap["RenewMaxRetry"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "RenewMaxRetry is not a float64 (expected int64)")
		return
	}
	renewMaxRetry := int64(renewMaxRetryFloat)

	renewTimeoutFloat, ok := privateMap["RenewTimeout"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "RenewTimeout is not a float64 (expected int64)")
		return
	}
	renewTimeout := int64(renewTimeoutFloat)

	renewResponse, ok := privateMap["RenewResponse"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " RenewResponse is not a string")
		return
	}

	closeMethod, ok := privateMap["CloseMethod"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseMethod is not a string")
		return
	}

	closeUrl, ok := privateMap["CloseUrl"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseUrl is not a string")
		return
	}

	skipClose, ok := privateMap["SkipClose"].(bool)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "SkipClose is not a boolean")
		return
	}

	var closeResponseCodes []string

	// Check if the key exists AND is not nil before asserting type.
	if rawList, exists := privateMap["CloseResponseCodes"]; exists && rawList != nil {
		switch v := rawList.(type) {
		case []interface{}:
			for _, item := range v {
				if strVal, ok := item.(string); ok {
					closeResponseCodes = append(closeResponseCodes, strVal)
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", "CloseResponseCodes contains non-string elements")
					return
				}
			}
		case []string:
			closeResponseCodes = v
		default:
			resp.Diagnostics.AddError(
				"Type Assertion Error",
				fmt.Sprintf("CloseResponseCodes has unexpected type: %T", rawList),
			)
			return
		}
	} else {
		// If attribute is missing, default to an empty slice without throwing an error.
		closeResponseCodes = []string{}
	}

	// Convert to Terraform ListValue.
	closeResponseCodesList, diags := types.ListValue(
		types.StringType,
		convertStringSliceToTFValues(closeResponseCodes),
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	closeRequestBody, ok := privateMap["CloseRequestBody"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseRequestBody is not a string")
		return
	}

	closeHeaders := make(map[string]string)
	if rawCloseHeaders, exists := privateMap["CloseHeaders"]; exists && rawCloseHeaders != nil {
		if rawHeaders, ok := rawCloseHeaders.(map[string]interface{}); ok {
			for key, value := range rawHeaders {
				if strValue, ok := value.(string); ok {
					closeHeaders[key] = strValue
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", fmt.Sprintf("CloseHeaders[%s] is not a string", key))
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("Type Assertion Error", "CloseHeaders is not a map[string]interface{}")
			return
		}
	}

	// Convert to Terraform types.Map.
	closeHeadersTF, diags := types.MapValue(types.StringType, convertStringMapToTFValues(closeHeaders))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	closeParameters := make(map[string]string) // Default to empty map

	tflog.Debug(ctx, fmt.Sprintf("raw close parameters: %v\n", privateMap["CloseRequestParameters"]))
	if rawCloseParameters, exists := privateMap["CloseRequestParameters"]; exists && rawCloseParameters != nil {
		if rawParameters, ok := rawCloseParameters.(map[string]interface{}); ok {
			for key, value := range rawParameters {
				if strValue, ok := value.(string); ok {
					closeParameters[key] = strValue
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", fmt.Sprintf("CloseRequestParameters[%s] is not a string", key))
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("Type Assertion Error", "CloseRequestParameters is not a map[string]interface{}")
			return
		}
	}

	// Convert to Terraform types.Map.
	closeParametersTF, diags := types.MapValue(types.StringType, convertStringMapToTFValues(closeParameters))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	closeRequestUrlString, ok := privateMap["CloseRequestUrlString"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseRequestUrlString is not a string")
		return
	}

	closeCertFile, ok := privateMap["CloseCertFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseCertFile is not a string")
		return
	}

	closeKeyFile, ok := privateMap["CloseKeyFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseKeyFile is not a string")
		return
	}

	closeCaCertFile, ok := privateMap["CloseCaCertFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseCaCertFile is not a string")
		return
	}

	closeCaCertDirectory, ok := privateMap["CloseCaCertDirectory"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseCaCertDirectory is not a string")
		return
	}

	closeSkipTlsVerify, ok := privateMap["CloseSkipTlsVerify"].(bool)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseSkipTlsVerify is not a bool")
		return
	}

	closeRetryIntervalFloat, ok := privateMap["CloseRetryInterval"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseRetryInterval is not a float64 (expected int64)")
		return
	}

	closeRetryInterval := int64(closeRetryIntervalFloat)

	closeMaxRetryFloat, ok := privateMap["CloseMaxRetry"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseMaxRetry is not a float64 (expected int64)")
		return
	}
	closeMaxRetry := int64(closeMaxRetryFloat)

	closeTimeoutFloat, ok := privateMap["CloseTimeout"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseTimeout is not a float64 (expected int64)")
		return
	}
	closeTimeout := int64(closeTimeoutFloat)

	closeResponse, ok := privateMap["CloseResponse"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseResponse is not a string")
		return
	}

	privateData := CurlEphemeralModel{
		RenewMethod:            types.StringValue(renewMethod),
		RenewUrl:               types.StringValue(renewUrl),
		SkipRenew:              types.BoolValue(skipRenew),
		RenewResponseCodes:     renewResponseCodesList,
		RenewHeaders:           renewHeadersTF,
		RenewRequestParameters: renewParametersTF,
		RenewRequestBody:       types.StringValue(renewRequestBody),
		RenewRequestUrlString:  types.StringValue(renewRequestUrlString),
		RenewCertFile:          types.StringValue(renewCertFile),
		RenewKeyFile:           types.StringValue(renewKeyFile),
		RenewCaCertFile:        types.StringValue(renewCaCertFile),
		RenewCaCertDirectory:   types.StringValue(renewCaCertDirectory),
		RenewSkipTlsVerify:     types.BoolValue(renewSkipTlsVerify),
		RenewRetryInterval:     types.Int64Value(renewRetryInterval),
		RenewMaxRetry:          types.Int64Value(renewMaxRetry),
		RenewTimeout:           types.Int64Value(renewTimeout),
		RenewResponse:          types.StringValue(renewResponse),

		CloseMethod:            types.StringValue(closeMethod),
		CloseUrl:               types.StringValue(closeUrl),
		SkipClose:              types.BoolValue(skipClose),
		CloseResponseCodes:     closeResponseCodesList,
		CloseHeaders:           closeHeadersTF,
		CloseRequestParameters: closeParametersTF,
		CloseRequestBody:       types.StringValue(closeRequestBody),
		CloseRequestUrlString:  types.StringValue(closeRequestUrlString),
		CloseCertFile:          types.StringValue(closeCertFile),
		CloseKeyFile:           types.StringValue(closeKeyFile),
		CloseCaCertFile:        types.StringValue(closeCaCertFile),
		CloseCaCertDirectory:   types.StringValue(closeCaCertDirectory),
		CloseSkipTlsVerify:     types.BoolValue(closeSkipTlsVerify),
		CloseRetryInterval:     types.Int64Value(closeRetryInterval),
		CloseMaxRetry:          types.Int64Value(closeMaxRetry),
		CloseTimeout:           types.Int64Value(closeTimeout),
		CloseResponse:          types.StringValue(closeResponse),
	}

	if privateData.SkipRenew.IsUnknown() {
		tflog.Debug(ctx, "`skip_renew` not set. using default value `true`")
		privateData.SkipRenew = types.BoolValue(true)
	}

	if privateData.SkipRenew.ValueBool() {
		tflog.Debug(ctx, "`skip_renew` set to `true`. Skipping renew call.")
		return
	}

	if !privateData.SkipRenew.ValueBool() {
		if privateData.RenewUrl.ValueString() == "" || privateData.RenewMethod.ValueString() == "" || privateData.RenewResponseCodes.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"If `skip_renew` is set to `false`, `renew_url`, `renew_method`, and `renew_response_codes` must be provided.",
			)
			return
		}
	}

	// Perform external call to renew "thing" data

	var client *http.Client
	useTls := !(privateData.RenewCertFile.IsNull() || privateData.RenewCertFile.ValueString() == "") ||
		!(privateData.RenewKeyFile.IsNull() || privateData.RenewKeyFile.ValueString() == "") ||
		!(privateData.RenewCaCertFile.IsNull() || privateData.RenewCaCertFile.ValueString() == "")

	if useTls {
		tflog.Debug(ctx, "using TLS client for renew call")
		// Build TLS Config
		tlsConfig := &TlsConfig{
			CertFile:        privateData.RenewCertFile.ValueString(),
			KeyFile:         privateData.RenewKeyFile.ValueString(),
			CaCertFile:      privateData.RenewCaCertFile.ValueString(),
			CaCertDirectory: privateData.RenewCaCertDirectory.ValueString(),
			SkipTlsVerify:   privateData.RenewSkipTlsVerify.ValueBool(),
		}

		// Validate TLS settings
		if tlsConfig.CertFile != "" && tlsConfig.KeyFile == "" {
			resp.Diagnostics.AddError("Validation Error", "`renew_key_file` must be set if `renew_cert_file` is set.")
			return
		}

		// Create TLS-enabled client
		client, err = createTlsClient(tlsConfig)
		if err != nil {
			resp.Diagnostics.AddError("TLS Client Creation Failed", err.Error())
			return
		}

	} else {
		// Use default non-TLS client
		tflog.Debug(ctx, "using default client for renew call")
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	reqBody := []byte(privateData.RenewRequestBody.ValueString())
	request, err := http.NewRequest(privateData.RenewMethod.ValueString(), privateData.RenewUrl.ValueString(), bytes.NewBuffer(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return
	}

	// Add headers
	if !privateData.RenewHeaders.IsNull() && !privateData.RenewHeaders.IsUnknown() {
		for k, v := range privateData.RenewHeaders.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Parameters: %v\n", privateData.RenewRequestParameters.Elements()))

	// Add query parameters
	if !privateData.RenewRequestParameters.IsNull() && !privateData.RenewRequestParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range privateData.RenewRequestParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
		tflog.Debug(ctx, fmt.Sprintf("Renew URL string: %s\n", request.URL.String()))
	}
	privateData.RenewRequestUrlString = types.StringValue(request.URL.String())

	timeout := 10 * time.Second
	if !privateData.RenewTimeout.IsNull() {
		timeout = time.Duration(privateData.RenewTimeout.ValueInt64()) * time.Second
	}

	var body []byte
	var bodyString string
	var statusCode int
	retryCount := 0

	for {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		response, err := client.Do(request.WithContext(ctxWithTimeout))
		if err != nil {
			if retryCount < int(privateData.RenewMaxRetry.ValueInt64()) {
				retryCount++
				time.Sleep(time.Duration(privateData.RenewRetryInterval.ValueInt64()) * time.Second)
				continue
			}
			resp.Diagnostics.AddError("Request Failed", err.Error())
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(response.Body)
		body, _ = io.ReadAll(response.Body)
		statusCode = response.StatusCode

		bodyString = string(body)
		if bodyString == "" {
			bodyString = "{}"
		}

		var responseCodes []string
		for _, v := range privateData.RenewResponseCodes.Elements() {
			if strVal, ok := v.(types.String); ok {
				responseCodes = append(responseCodes, strVal.ValueString())
			}
		}

		if responseCodeChecker(responseCodes, strconv.Itoa(statusCode)) {
			break
		}

		if retryCount < int(privateData.RenewMaxRetry.ValueInt64()) {
			retryCount++
			time.Sleep(time.Duration(privateData.RenewRetryInterval.ValueInt64()) * time.Second)
		} else {
			resp.Diagnostics.AddError("Unexpected Response Code", fmt.Sprintf("Received status code: %d", statusCode))
			return
		}
	}

	privateData.RenewRequestUrlString = types.StringValue(request.URL.String())
	privateData.RenewResponse = types.StringValue(bodyString)
	privateData.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	// Renew again
	resp.RenewAt = time.Now().Add(time.Duration(privateData.RenewInterval.ValueInt64()) * time.Second)

}

func (e *EphemeralCurlResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	privateBytes, diags := req.Private.GetKey(ctx, "response")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Error getting private data at close step")
		return
	}

	//tflog.Debug(ctx, fmt.Sprintf("Raw privateBytes received: %s", string(privateBytes)))

	var privateMap map[string]interface{}
	err := json.Unmarshal(privateBytes, &privateMap)
	if err != nil {
		resp.Diagnostics.AddError("Error unmarshaling response", fmt.Sprintf("%s", err))
		return
	}

	closeMethod, ok := privateMap["CloseMethod"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseMethod is not a string")
		return
	}

	closeUrl, ok := privateMap["CloseUrl"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseUrl is not a string")
		return
	}

	skipClose, ok := privateMap["SkipClose"].(bool)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "SkipClose is not a boolean")
		return
	}

	var closeResponseCodes []string

	// Check if the key exists AND is not nil before asserting type
	if rawList, exists := privateMap["CloseResponseCodes"]; exists && rawList != nil {
		switch v := rawList.(type) {
		case []interface{}:
			for _, item := range v {
				if strVal, ok := item.(string); ok {
					closeResponseCodes = append(closeResponseCodes, strVal)
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", "CloseResponseCodes contains non-string elements")
					return
				}
			}
		case []string:
			closeResponseCodes = v
		default:
			resp.Diagnostics.AddError(
				"Type Assertion Error",
				fmt.Sprintf("CloseResponseCodes has unexpected type: %T", rawList),
			)
			return
		}
	} else {
		// If attribute is missing, default to an empty slice without throwing an error
		closeResponseCodes = []string{}
	}

	// Convert to Terraform ListValue
	closeResponseCodesList, diags := types.ListValue(
		types.StringType,
		convertStringSliceToTFValues(closeResponseCodes),
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Converted response codes list: %v", privateMap["CloseResponseCodes"]))

	closeRequestBody, ok := privateMap["CloseRequestBody"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseRequestBody is not a string")
		return
	}

	closeHeaders := make(map[string]string) // Default to empty map
	if rawCloseHeaders, exists := privateMap["CloseHeaders"]; exists && rawCloseHeaders != nil {
		if rawHeaders, ok := rawCloseHeaders.(map[string]interface{}); ok {
			for key, value := range rawHeaders {
				if strValue, ok := value.(string); ok {
					closeHeaders[key] = strValue
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", fmt.Sprintf("CloseHeaders[%s] is not a string", key))
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("Type Assertion Error", "CloseHeaders is not a map[string]interface{}")
			return
		}
	}

	// Convert to Terraform types.Map
	closeHeadersTF, diags := types.MapValue(types.StringType, convertStringMapToTFValues(closeHeaders))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	closeParameters := make(map[string]string) // Default to empty map

	if rawCloseParameters, exists := privateMap["CloseRequestParameters"]; exists && rawCloseParameters != nil {
		if rawParameters, ok := rawCloseParameters.(map[string]interface{}); ok {
			for key, value := range rawParameters {
				if strValue, ok := value.(string); ok {
					closeParameters[key] = strValue
				} else {
					resp.Diagnostics.AddError("Type Assertion Error", fmt.Sprintf("CloseRequestParameters[%s] is not a string", key))
					return
				}
			}
		} else {
			resp.Diagnostics.AddError("Type Assertion Error", "CloseRequestParameters is not a map[string]interface{}")
			return
		}
	}

	// Convert to Terraform types.Map
	closeParametersTF, diags := types.MapValue(types.StringType, convertStringMapToTFValues(closeParameters))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	closeRequestUrlString, ok := privateMap["CloseRequestUrlString"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseRequestUrlString is not a string")
		return
	}

	closeCertFile, ok := privateMap["CloseCertFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseCertFile is not a string")
		return
	}

	closeKeyFile, ok := privateMap["CloseKeyFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseKeyFile is not a string")
		return
	}

	closeCaCertFile, ok := privateMap["CloseCaCertFile"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseCaCertFile is not a string")
		return
	}

	closeCaCertDirectory, ok := privateMap["CloseCaCertDirectory"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseCaCertDirectory is not a string")
		return
	}

	closeSkipTlsVerify, ok := privateMap["CloseSkipTlsVerify"].(bool)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseSkipTlsVerify is not a bool")
		return
	}

	closeRetryIntervalFloat, ok := privateMap["CloseRetryInterval"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseRetryInterval is not a float64 (expected int64)")
		return
	}
	closeRetryInterval := int64(closeRetryIntervalFloat)

	closeMaxRetryFloat, ok := privateMap["CloseMaxRetry"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseMaxRetry is not a float64 (expected int64)")
		return
	}
	closeMaxRetry := int64(closeMaxRetryFloat)

	closeTimeoutFloat, ok := privateMap["CloseTimeout"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", "CloseTimeout is not a float64 (expected int64)")
		return
	}
	closeTimeout := int64(closeTimeoutFloat)

	closeResponse, ok := privateMap["CloseResponse"].(string)
	if !ok {
		resp.Diagnostics.AddError("Type Assertion Error", " CloseResponse is not a string")
		return
	}

	privateData := CurlEphemeralModel{
		CloseMethod:            types.StringValue(closeMethod),
		CloseUrl:               types.StringValue(closeUrl),
		SkipClose:              types.BoolValue(skipClose),
		CloseResponseCodes:     closeResponseCodesList,
		CloseHeaders:           closeHeadersTF,
		CloseRequestParameters: closeParametersTF,
		CloseRequestBody:       types.StringValue(closeRequestBody),
		CloseRequestUrlString:  types.StringValue(closeRequestUrlString),
		CloseCertFile:          types.StringValue(closeCertFile),
		CloseKeyFile:           types.StringValue(closeKeyFile),
		CloseCaCertFile:        types.StringValue(closeCaCertFile),
		CloseCaCertDirectory:   types.StringValue(closeCaCertDirectory),
		CloseSkipTlsVerify:     types.BoolValue(closeSkipTlsVerify),
		CloseRetryInterval:     types.Int64Value(closeRetryInterval),
		CloseMaxRetry:          types.Int64Value(closeMaxRetry),
		CloseTimeout:           types.Int64Value(closeTimeout),
		CloseResponse:          types.StringValue(closeResponse),
	}

	//tflog.Debug(ctx, fmt.Sprintf("Unmarshaled privateData: %+v", privateData))
	//tflog.Debug(ctx, fmt.Sprintf("Private bytes: %v", privateBytes))
	//tflog.Debug(ctx, fmt.Sprintf("Private data: %v", privateData.Name))
	//tflog.Debug(ctx, fmt.Sprintf("Private data skip_close value: %v", privateData.SkipClose.ValueBool()))
	//tflog.Debug(ctx, fmt.Sprintf("Method: %s \n Url: %s \n", privateData.CloseMethod.ValueString(), privateData.CloseUrl.ValueString()))

	// Perform external call to close/clean up data

	if privateData.SkipClose.IsUnknown() {
		tflog.Debug(ctx, "`skip_close` not set. using default value `true`")
		privateData.SkipClose = types.BoolValue(true)
	}

	if privateData.SkipClose.ValueBool() {
		tflog.Debug(ctx, "`skip_close` set to `true`. Skipping close call.")
		return
	}

	var client *http.Client
	useCloseTls := !(privateData.CloseCertFile.IsNull() || privateData.CloseCertFile.ValueString() == "") ||
		!(privateData.CloseKeyFile.IsNull() || privateData.CloseKeyFile.ValueString() == "") ||
		!(privateData.CloseCaCertFile.IsNull() || privateData.CloseCaCertFile.ValueString() == "")

	if useCloseTls {
		tflog.Debug(ctx, "Using custom TLS client for Close() operation")

		closeTlsConfig := &TlsConfig{
			CertFile:        privateData.CloseCertFile.ValueString(),
			KeyFile:         privateData.CloseKeyFile.ValueString(),
			CaCertFile:      privateData.CloseCaCertFile.ValueString(),
			CaCertDirectory: privateData.CloseCaCertDirectory.ValueString(),
			SkipTlsVerify:   privateData.CloseSkipTlsVerify.ValueBool(),
		}

		tlsClient, err := createTlsClient(closeTlsConfig)
		if err != nil {
			resp.Diagnostics.AddError("Close Error", fmt.Sprintf("Failed to create TLS client: %s", err))
			return
		}
		client = tlsClient
	} else {
		// Default non-TLS client
		tflog.Debug(ctx, "Using default HTTP client for Close() operation")
		client = &http.Client{Timeout: 30 * time.Second}
	}

	// Build Close Request
	var reqBody io.Reader = nil
	if !privateData.CloseRequestBody.IsNull() && !privateData.CloseRequestBody.IsUnknown() {
		reqBody = bytes.NewBuffer([]byte(privateData.CloseRequestBody.ValueString()))
	}
	tflog.Debug(ctx, fmt.Sprintf("Method: %s \n Url: %s \n", privateData.CloseMethod.ValueString(), privateData.CloseUrl.ValueString()))

	request, err := http.NewRequest(privateData.CloseMethod.ValueString(), privateData.CloseUrl.ValueString(), reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Close Error", fmt.Sprintf("Failed to create request: %s", err))
		return
	}

	if !privateData.CloseHeaders.IsNull() && !privateData.CloseHeaders.IsUnknown() {
		for k, v := range privateData.CloseHeaders.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	} else {
		tflog.Debug(ctx, "No CloseHeaders provided, proceeding without headers")
	}

	// Add Query Parameters
	if !privateData.CloseRequestParameters.IsNull() && !privateData.CloseRequestParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range privateData.CloseRequestParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}

	// Execute Request with Retry Logic
	timeout := time.Duration(privateData.CloseTimeout.ValueInt64()) * time.Second
	retryInterval := time.Duration(privateData.CloseRetryInterval.ValueInt64()) * time.Second
	maxRetry := int(privateData.CloseMaxRetry.ValueInt64())

	tflog.Debug(ctx, "Making Close() request", map[string]interface{}{
		"url":     request.URL.String(),
		"method":  request.Method,
		"headers": request.Header,
	})

	var bodyBytes []byte
	var statusCode int
	var retryCount int

	for {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		response, err := client.Do(request.WithContext(ctxWithTimeout))
		if err != nil {
			if retryCount < maxRetry {
				retryCount++
				tflog.Warn(ctx, fmt.Sprintf("Close request failed, retrying... Attempt %d/%d", retryCount, maxRetry))
				time.Sleep(retryInterval)
				continue
			}
			resp.Diagnostics.AddError("Close Error", fmt.Sprintf("Request failed: %s", err))
			return
		}

		bodyBytes, _ = io.ReadAll(response.Body)
		statusCode = response.StatusCode
		err = response.Body.Close()
		if err != nil {
			return
		}

		var expectedCodes []string
		tflog.Debug(ctx, fmt.Sprintf("private data response code list: %v", privateData.CloseResponseCodes.Elements()))
		for _, v := range privateData.CloseResponseCodes.Elements() {
			if strVal, ok := v.(types.String); ok {
				expectedCodes = append(expectedCodes, strVal.ValueString())
			}
		}

		tflog.Debug(ctx, fmt.Sprintf("response code received: %v", statusCode))
		tflog.Debug(ctx, fmt.Sprintf("expected response code received: %v", expectedCodes))

		// Validate Response Code
		if responseCodeChecker(expectedCodes, strconv.Itoa(statusCode)) {
			tflog.Debug(ctx, "Close request completed successfully")
			break
		} else {
			if retryCount < maxRetry {
				retryCount++
				tflog.Warn(ctx, fmt.Sprintf(
					"Close request returned unexpected status %d. Retrying (%d/%d)",
					statusCode, retryCount, maxRetry,
				))
				time.Sleep(retryInterval)
				continue
			}
			resp.Diagnostics.AddError(
				"Close Error",
				fmt.Sprintf("Unexpected response code from close request: %d. Response: %s", statusCode, string(bodyBytes)),
			)
			return
		}
	}

}

func (e EphemeralCurlResource) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	return []ephemeral.ConfigValidator{
		ephemeralvalidator.RequiredTogether(
			path.MatchRoot("cert_file"),
			path.MatchRoot("key_file"),
		),
		ephemeralvalidator.Conflicting(
			path.MatchRoot("ca_cert_file"),
			path.MatchRoot("ca_cert_directory"),
		),
		ephemeralvalidator.RequiredTogether(
			path.MatchRoot("renew_url"),
			path.MatchRoot("renew_method"),
			path.MatchRoot("renew_response_codes"),
		),
		ephemeralvalidator.RequiredTogether(
			path.MatchRoot("close_url"),
			path.MatchRoot("close_method"),
			path.MatchRoot("close_response_codes"),
		),
		ephemeralvalidator.RequiredTogether(
			path.MatchRoot("renew_cert_file"),
			path.MatchRoot("renew_key_file"),
		),
		ephemeralvalidator.Conflicting(
			path.MatchRoot("renew_ca_cert_file"),
			path.MatchRoot("renew_ca_cert_directory"),
		),
		ephemeralvalidator.RequiredTogether(
			path.MatchRoot("close_cert_file"),
			path.MatchRoot("close_key_file"),
		),
		ephemeralvalidator.Conflicting(
			path.MatchRoot("close_ca_cert_file"),
			path.MatchRoot("close_ca_cert_directory"),
		),
	}
}
