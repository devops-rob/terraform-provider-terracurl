package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ action.Action = &CurlAction{}

type CurlAction struct{}

func NewCurlAction() action.Action {
	return &CurlAction{}
}

type CurlActionModel struct {
	URL               types.String `tfsdk:"url"`
	Method            types.String `tfsdk:"method"`
	RequestBody       types.String `tfsdk:"request_body"`
	Headers           types.Map    `tfsdk:"headers"`
	RequestParameters types.Map    `tfsdk:"request_parameters"`
	CertFile          types.String `tfsdk:"cert_file"`
	KeyFile           types.String `tfsdk:"key_file"`
	CaCertFile        types.String `tfsdk:"ca_cert_file"`
	CaCertDirectory   types.String `tfsdk:"ca_cert_directory"`
	SkipTlsVerify     types.Bool   `tfsdk:"skip_tls_verify"`
	RetryInterval     types.Int64  `tfsdk:"retry_interval"`
	MaxRetry          types.Int64  `tfsdk:"max_retry"`
	Timeout           types.Int64  `tfsdk:"timeout"`
	ResponseCodes     types.List   `tfsdk:"response_codes"`
}

func (c *CurlAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TerraCurl request action",
		Attributes: map[string]schema.Attribute{
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
				MarkdownDescription: "Time in seconds between each retry attempt",
			},
			"max_retry": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tries until it is marked as failed",
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Time in seconds before each request times out. Defaults to 10",
			},
			"response_codes": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "A list of the expected response status codes that are considered successful.",
				ElementType:         types.StringType,
			},
		},
	}
}

func (c *CurlAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (c *CurlAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data CurlActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	client, diags := data.prepareClient()
	resp.Diagnostics.Append(diags...)

	request, diags := data.prepareRequest()
	resp.Diagnostics.Append(diags...)

	tflog.Debug(ctx, fmt.Sprintf("Invoke Action Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, request.Body))

	var responseCodes []string
	for _, v := range data.ResponseCodes.Elements() {
		if strVal, ok := v.(types.String); ok {
			responseCodes = append(responseCodes, strVal.ValueString())
		}
	}

	requestTimeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		requestTimeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	retryInterval := 5 * time.Second
	if !data.RetryInterval.IsNull() {
		retryInterval = time.Duration(data.RetryInterval.ValueInt64()) * time.Second
	}

	doRequest := func() (diag.Diagnostics, error) {
		var requestDiagnostics diag.Diagnostics

		ctxWithTimeout, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		response, err := client.Do(request.WithContext(ctxWithTimeout))
		if err != nil {
			requestDiagnostics.AddError("Request Failed", err.Error())
			return requestDiagnostics, err
		}

		defer func() {
			_ = response.Body.Close()
		}()

		if !responseCodeChecker(responseCodes, strconv.Itoa(response.StatusCode)) {
			resp.Diagnostics.AddError("Unexpected Response Code", fmt.Sprintf("Received status code: %d", response.StatusCode))
			return requestDiagnostics, errors.New("unexpected status code")
		}

		return requestDiagnostics, nil
	}

	retryOptions := []backoff.RetryOption{
		backoff.WithBackOff(backoff.NewConstantBackOff(retryInterval)),
		backoff.WithMaxTries(uint(data.MaxRetry.ValueInt64())),
	}

	errDiags, err := backoff.Retry(ctx, doRequest, retryOptions...)
	if err != nil {
		resp.Diagnostics.Append(errDiags...)
	}
}

func (m *CurlActionModel) prepareClient() (*http.Client, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	useTLS := !m.CertFile.IsNull() || !m.KeyFile.IsNull() || !m.CaCertFile.IsNull() || !m.CaCertDirectory.IsNull()

	if !useTLS {
		return &http.Client{
			Timeout: 30 * time.Second,
		}, diagnostics
	}

	tlsConfig := &TlsConfig{
		CertFile:        m.CertFile.ValueString(),
		KeyFile:         m.KeyFile.ValueString(),
		CaCertFile:      m.CaCertFile.ValueString(),
		CaCertDirectory: m.CaCertDirectory.ValueString(),
		SkipTlsVerify:   m.SkipTlsVerify.ValueBool(),
	}

	if tlsConfig.CertFile != "" && tlsConfig.KeyFile == "" {
		diagnostics.AddError("Validation Error", "`key_file` must be set if `cert_file` is set.")
		return nil, diagnostics
	}

	client, err := createTlsClient(tlsConfig)
	if err != nil {
		diagnostics.AddError("TLS Client Creation Failed", err.Error())
		return nil, diagnostics
	}

	return client, diagnostics
}

func (m *CurlActionModel) prepareRequest() (*http.Request, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	reqBody := []byte(m.RequestBody.ValueString())
	request, err := http.NewRequest(m.Method.ValueString(), m.URL.ValueString(), bytes.NewBuffer(reqBody))
	if err != nil {
		diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return nil, diagnostics
	}

	if !m.Headers.IsNull() && !m.Headers.IsUnknown() {
		for k, v := range m.Headers.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	if !m.RequestParameters.IsNull() && !m.RequestParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range m.RequestParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}

	return request, diagnostics
}
