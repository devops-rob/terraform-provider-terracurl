package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/actionvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ action.Action = &CurlAction{}
var _ action.ActionWithConfigure = &CurlAction{}

type CurlAction struct {
	meta *ProviderMeta
}

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

func (c *CurlAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
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
				MarkdownDescription: "Map of headers to attach to the API call." + hostHeaderMarkdownSuffix,
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

func (c *CurlAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	meta, ok := req.ProviderData.(*ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	c.meta = meta
}

func (c *CurlAction) providerMeta() *ProviderMeta {
	if c.meta != nil {
		return c.meta
	}
	return DefaultProviderMeta()
}

func (c *CurlAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data CurlActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	useTLS := !data.CertFile.IsNull() || !data.KeyFile.IsNull() || !data.CaCertFile.IsNull() || !data.CaCertDirectory.IsNull()

	var tlsConfig *TlsConfig
	if useTLS {
		tlsConfig = &TlsConfig{
			CertFile:        data.CertFile.ValueString(),
			KeyFile:         data.KeyFile.ValueString(),
			CaCertFile:      data.CaCertFile.ValueString(),
			CaCertDirectory: data.CaCertDirectory.ValueString(),
			SkipTlsVerify:   data.SkipTlsVerify.ValueBool(),
		}

		if tlsConfig.CertFile != "" && tlsConfig.KeyFile == "" {
			resp.Diagnostics.AddError("Validation Error", "`key_file` must be set if `cert_file` is set.")
			return
		}
	}

	client, err := c.providerMeta().NewHTTPClient(tlsConfig)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Client Creation Failed", err.Error())
		return
	}

	reqBody := []byte(data.RequestBody.ValueString())
	request, err := http.NewRequest(data.Method.ValueString(), data.URL.ValueString(), bytes.NewBuffer(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return
	}

	applyRequestHeaders(request, data.Headers)

	if !data.RequestParameters.IsNull() && !data.RequestParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.RequestParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}

	tflog.Debug(ctx, fmt.Sprintf("Invoke Action Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, request.Body))

	timeout := 10 * time.Second
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	retryCount := 0
	for {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)

		response, err := client.Do(request.WithContext(ctxWithTimeout))
		cancel()
		if err != nil {
			if retryCount < int(data.MaxRetry.ValueInt64()) {
				retryCount++
				time.Sleep(time.Duration(data.RetryInterval.ValueInt64()) * time.Second)
				continue
			}
			resp.Diagnostics.AddError("Request Failed", err.Error())
			return
		}

		statusCode := response.StatusCode
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()

		if responseCodeChecker(data.ResponseCodes, statusCode) {
			return
		}

		if retryCount < int(data.MaxRetry.ValueInt64()) {
			retryCount++
			time.Sleep(time.Duration(data.RetryInterval.ValueInt64()) * time.Second)
		} else {
			resp.Diagnostics.AddError("Unexpected Response Code", fmt.Sprintf("Received status code: %d", statusCode))
			return
		}
	}
}

func (c *CurlAction) ConfigValidators(_ context.Context) []action.ConfigValidator {
	return []action.ConfigValidator{
		actionvalidator.RequiredTogether(
			path.MatchRoot("cert_file"),
			path.MatchRoot("key_file"),
		),
		actionvalidator.Conflicting(
			path.MatchRoot("ca_cert_file"),
			path.MatchRoot("ca_cert_directory"),
		),
		actionvalidator.RequiredTogether(
			path.MatchRoot("max_retry"),
			path.MatchRoot("retry_interval"),
		),
	}
}
