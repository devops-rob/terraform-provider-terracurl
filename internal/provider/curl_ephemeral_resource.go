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
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var _ provider.ProviderWithEphemeralResources = (*TerraCurlProvider)(nil)

type EphemeralCurlResource struct {
	client *http.Client
}

func (e *EphemeralCurlResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

// EphemeralResources With the provider.ProviderWithEphemeralResources implementation
func (p *TerraCurlProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewCurlEphemeralResource,
	}
}

// NewCurlEphemeralResource With the ephemeral.EphemeralResource implementation
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
				// TODO - set default to true in close method
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
	var data CurlEphemeralModel

	// Read Terraform config data into the model
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

	// useTLS is used to decide
	useTLS := !data.CertFile.IsNull() || !data.KeyFile.IsNull() || !data.CaCertFile.IsNull() || !data.CaCertDirectory.IsNull()

	var client *http.Client
	var err error

	if useTLS {
		// Build TLS Config
		tlsConfig := &TlsConfig{
			CertFile:        data.CertFile.ValueString(),
			KeyFile:         data.KeyFile.ValueString(),
			CaCertFile:      data.CaCertFile.ValueString(),
			CaCertDirectory: data.CaCertDirectory.ValueString(),
			SkipTlsVerify:   data.SkipTlsVerify.ValueBool(),
		}

		// Validate TLS settings
		if tlsConfig.CertFile != "" && tlsConfig.KeyFile == "" {
			resp.Diagnostics.AddError("Validation Error", "`key_file` must be set if `cert_file` is set.")
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

	// Add headers
	if !data.Headers.IsNull() && !data.Headers.IsUnknown() {
		for k, v := range data.Headers.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	// Add query parameters
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
		body, _ = ioutil.ReadAll(response.Body)
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

	if !data.SkipRenew.ValueBool() {
		resp.RenewAt = time.Now().Add(time.Duration(data.RenewInterval.ValueInt64()) * time.Second)

		privateData, diags := json.Marshal(&data)

		if diags != nil {
			resp.Diagnostics.AddError("Error marshaling data", fmt.Sprintf("%s", diags))
			return
		}
		resp.Private.SetKey(ctx, "response", privateData)
	}

	// Save data into ephemeral result data
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func (e *EphemeralCurlResource) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	privateBytes, diags := req.Private.GetKey(ctx, "response")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var data CurlEphemeralModel

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

	// Setting this as a sane default
	if data.SkipRenew.IsUnknown() || data.SkipRenew.IsNull() {
		tflog.Debug(ctx, "`skip_renew` not set. Using default value of `true`")
		data.SkipRenew = types.BoolValue(true)
	}

	// Unmarshal private data (error handling omitted for brevity).
	if data.SkipRenew.ValueBool() {
		tflog.Debug(ctx, "Skipping Renew() as skip_renew is true")
		return
	}

	err := json.Unmarshal(privateBytes, &data)
	if err != nil {
		return
	}

	if data.RenewUrl.IsNull() || data.RenewMethod.IsNull() || data.RenewResponseCodes.IsNull() {
		resp.Diagnostics.AddError(
			"Read Configuration Error",
			"`renew_url`, `renew_method`, and `renew_response_codes` are required when `skip_renew` is false.",
		)
		return
	}

	// Perform external call to renew "thing" data

	useTLS := !data.CertFile.IsNull() || !data.KeyFile.IsNull() || !data.CaCertFile.IsNull() || !data.CaCertDirectory.IsNull()

	var client *http.Client

	if useTLS {
		// Build TLS Config
		tlsConfig := &TlsConfig{
			CertFile:        data.RenewCertFile.ValueString(),
			KeyFile:         data.RenewKeyFile.ValueString(),
			CaCertFile:      data.RenewCaCertFile.ValueString(),
			CaCertDirectory: data.RenewCaCertDirectory.ValueString(),
			SkipTlsVerify:   data.RenewSkipTlsVerify.ValueBool(),
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
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	reqBody := []byte(data.RenewRequestBody.ValueString())
	request, err := http.NewRequest(data.RenewMethod.ValueString(), data.RenewUrl.ValueString(), bytes.NewBuffer(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return
	}

	// Add headers
	if !data.RenewHeaders.IsNull() && !data.RenewHeaders.IsUnknown() {
		for k, v := range data.RenewHeaders.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	// Add query parameters
	if !data.RenewRequestParameters.IsNull() && !data.RenewRequestParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.RenewRequestParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}
	data.RenewRequestUrlString = types.StringValue(request.URL.String())

	timeout := 10 * time.Second
	if !data.RenewTimeout.IsNull() {
		timeout = time.Duration(data.RenewTimeout.ValueInt64()) * time.Second
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
			if retryCount < int(data.RenewMaxRetry.ValueInt64()) {
				retryCount++
				time.Sleep(time.Duration(data.RenewRetryInterval.ValueInt64()) * time.Second)
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
		body, _ = ioutil.ReadAll(response.Body)
		statusCode = response.StatusCode

		bodyString = string(body)
		if bodyString == "" {
			bodyString = "{}"
		}

		var responseCodes []string
		for _, v := range data.RenewResponseCodes.Elements() {
			if strVal, ok := v.(types.String); ok {
				responseCodes = append(responseCodes, strVal.ValueString())
			}
		}

		if responseCodeChecker(responseCodes, strconv.Itoa(statusCode)) {
			break
		}

		if retryCount < int(data.RenewMaxRetry.ValueInt64()) {
			retryCount++
			time.Sleep(time.Duration(data.RenewRetryInterval.ValueInt64()) * time.Second)
		} else {
			resp.Diagnostics.AddError("Unexpected Response Code", fmt.Sprintf("Received status code: %d", statusCode))
			return
		}
	}

	data.RenewRequestUrlString = types.StringValue(request.URL.String())
	data.RenewResponse = types.StringValue(bodyString)
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	// Renew again
	resp.RenewAt = time.Now().Add(time.Duration(data.RenewInterval.ValueInt64()) * time.Second)

	// If needed, you can also set new `Private` data on the response.
	privateData, err := json.Marshal(&data)
	if err != nil {
		diags.AddError("Error marshaling renewal data", fmt.Sprintf("%s", err))
	}
	resp.Private.SetKey(ctx, "response", privateData)
}

func (e *EphemeralCurlResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	privateBytes, diags := req.Private.GetKey(ctx, "response")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unmarshal private data (error handling omitted for brevity).
	var privateData CurlEphemeralModel
	err := json.Unmarshal(privateBytes, &privateData)
	if err != nil {
		resp.Diagnostics.AddError("Error unmarshaling response", fmt.Sprintf("%s", err))
		return
	}

	// Perform external call to close/clean up data

	if privateData.SkipClose.IsUnknown() || privateData.SkipClose.IsNull() {
		tflog.Debug(ctx, "`skip_close` not set. using default value `true`")
		privateData.SkipClose = types.BoolValue(true)
	}

	if privateData.SkipClose.ValueBool() {
		tflog.Debug(ctx, "`skip_close` set to `true`. Skipping close call.")
		return
	}

	var client *http.Client
	useCloseTls := !privateData.CloseCertFile.IsNull() || !privateData.CloseKeyFile.IsNull() || !privateData.CloseCaCertFile.IsNull()

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

	request, err := http.NewRequest(privateData.CloseMethod.ValueString(), privateData.CloseUrl.ValueString(), reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Close Error", fmt.Sprintf("Failed to create request: %s", err))
		return
	}

	// Add Headers
	if !privateData.CloseHeaders.IsNull() && !privateData.CloseHeaders.IsUnknown() {
		for k, v := range privateData.CloseHeaders.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
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

		bodyBytes, _ = ioutil.ReadAll(response.Body)
		statusCode = response.StatusCode
		err = response.Body.Close()
		if err != nil {
			return
		}

		var expectedCodes []string
		for _, v := range privateData.CloseResponseCodes.Elements() {
			if strVal, ok := v.(types.String); ok {
				expectedCodes = append(expectedCodes, strVal.ValueString())
			}
		}

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

	//data.DestroyRequestUrlString = types.StringValue(data.DestroyUrl.ValueString())
	//data.RequestUrlString = types.StringValue(request.URL.String())
	//data.Response = types.StringValue(string(bodyBytes))
	//data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

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
