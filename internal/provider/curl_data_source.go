package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &CurlDataSource{}

type ThingDataSource struct{}

type CurlDataSource struct {
	meta *ProviderMeta
}

func NewCurlDataSource() datasource.DataSource {
	return &CurlDataSource{}
}

type CurlDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
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
	SensitiveResponse types.String `tfsdk:"sensitive_response"`
	ResponseSensitive types.Bool   `tfsdk:"response_sensitive"`
	ResponseCodes     types.List   `tfsdk:"response_codes"`
	StatusCode        types.String `tfsdk:"status_code"`
}

func (d *CurlDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (d *CurlDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TerraCurl request data source",
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
				MarkdownDescription: "Map of headers to attach to the API call." + hostHeaderMarkdownSuffix,
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
				MarkdownDescription: "JSON response received from request. Empty when `response_sensitive` is `true`; use `sensitive_response` instead.",
			},
			"sensitive_response": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "JSON response received from request, marked as sensitive so it is not displayed in plan output. Populated only when `response_sensitive` is `true`.",
			},
			"response_sensitive": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set to `true` to treat the response as sensitive. When enabled, the response body is written to `sensitive_response` (a sensitive attribute) and `response` is left empty so that secret values are not displayed in plan output. Defaults to `false` to preserve existing behavior.",
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
		},
	}
}

func (d *CurlDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	meta, ok := req.ProviderData.(*ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.meta = meta
}

func (d *CurlDataSource) providerMeta() *ProviderMeta {
	if d.meta != nil {
		return d.meta
	}
	return DefaultProviderMeta()
}

func (d *CurlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CurlDataSourceModel

	// Read Terraform configuration data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	data.ID = types.StringValue(data.Name.ValueString())

	useTLS := !data.CertFile.IsNull() || !data.KeyFile.IsNull() || !data.CaCertFile.IsNull() || !data.CaCertDirectory.IsNull()

	var client *http.Client
	var err error
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

	client, err = d.providerMeta().NewHTTPClient(tlsConfig)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Client Creation Failed", err.Error())
		return
	}

	reqBody := []byte(data.RequestBody.ValueString())
	request, err := http.NewRequest(data.Method.ValueString(), data.Url.ValueString(), bytes.NewBuffer(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return
	}

	// Add headers.
	applyRequestHeaders(request, data.Headers)

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

	tflog.Debug(ctx, fmt.Sprintf("Data Source API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, request.Body))

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

		if responseCodeChecker(data.ResponseCodes, statusCode) {
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
	setDataSourceResponseValues(&data, bodyString)
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d CurlDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("cert_file"),
			path.MatchRoot("key_file"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("ca_cert_file"),
			path.MatchRoot("ca_cert_directory"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("max_retry"),
			path.MatchRoot("retry_interval"),
		),
	}
}
