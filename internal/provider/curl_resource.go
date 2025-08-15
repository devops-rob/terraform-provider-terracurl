package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	resourceSchemaV0 = 0
	resourceSchemaV1 = 1
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CurlResource{}
var _ resource.ResourceWithImportState = &CurlResource{}

func NewCurlResource() resource.Resource {
	return &CurlResource{}
}

// CurlResource defines the resource implementation.
type CurlResource struct {
	client *http.Client
}

// CurlResourceModel describes the resource data model.
type CurlResourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Url                     types.String `tfsdk:"url"`
	Method                  types.String `tfsdk:"method"`
	RequestBody             types.String `tfsdk:"request_body"`
	Headers                 types.Map    `tfsdk:"headers"`
	Parameters              types.Map    `tfsdk:"parameters"`
	RequestUrlString        types.String `tfsdk:"request_url_string"`
	CertFile                types.String `tfsdk:"cert_file"`
	KeyFile                 types.String `tfsdk:"key_file"`
	CaCertFile              types.String `tfsdk:"ca_cert_file"`
	CaCertDirectory         types.String `tfsdk:"ca_cert_directory"`
	SkipTlsVerify           types.Bool   `tfsdk:"skip_tls_verify"`
	RetryInterval           types.Int64  `tfsdk:"retry_interval"`
	MaxRetry                types.Int64  `tfsdk:"max_retry"`
	Timeout                 types.Int64  `tfsdk:"timeout"`
	Response                types.String `tfsdk:"response"`
	ResponseCodes           types.List   `tfsdk:"response_codes"`
	StatusCode              types.String `tfsdk:"status_code"`
	SkipDestroy             types.Bool   `tfsdk:"skip_destroy"`
	DestroyUrl              types.String `tfsdk:"destroy_url"`
	DestroyMethod           types.String `tfsdk:"destroy_method"`
	DestroyRequestBody      types.String `tfsdk:"destroy_request_body"`
	DestroyHeaders          types.Map    `tfsdk:"destroy_headers"`
	DestroyParameters       types.Map    `tfsdk:"destroy_parameters"`
	DestroyRequestUrlString types.String `tfsdk:"destroy_request_url_string"`
	DestroyCertFile         types.String `tfsdk:"destroy_cert_file"`
	DestroyKeyFile          types.String `tfsdk:"destroy_key_file"`
	DestroyCaCertFile       types.String `tfsdk:"destroy_ca_cert_file"`
	DestroyCaCertDirectory  types.String `tfsdk:"destroy_ca_cert_directory"`
	DestroySkipTlsVerify    types.Bool   `tfsdk:"destroy_skip_tls_verify"`
	DestroyRetryInterval    types.Int64  `tfsdk:"destroy_retry_interval"`
	DestroyMaxRetry         types.Int64  `tfsdk:"destroy_max_retry"`
	DestroyTimeout          types.Int64  `tfsdk:"destroy_timeout"`
	DestroyResponseCodes    types.List   `tfsdk:"destroy_response_codes"`
	SkipRead                types.Bool   `tfsdk:"skip_read"`
	ReadUrl                 types.String `tfsdk:"read_url"`
	ReadMethod              types.String `tfsdk:"read_method"`
	ReadHeaders             types.Map    `tfsdk:"read_headers"`
	ReadParameters          types.Map    `tfsdk:"read_parameters"`
	ReadRequestBody         types.String `tfsdk:"read_request_body"`
	ReadCertFile            types.String `tfsdk:"read_cert_file"`
	ReadKeyFile             types.String `tfsdk:"read_key_file"`
	ReadCaCertFile          types.String `tfsdk:"read_ca_cert_file"`
	ReadCaCertDirectory     types.String `tfsdk:"read_ca_cert_directory"`
	ReadSkipTlsVerify       types.Bool   `tfsdk:"read_skip_tls_verify"`
	ReadResponseCodes       types.List   `tfsdk:"read_response_codes"`
	DriftMarker             types.String `tfsdk:"drift_marker"`
	IgnoreResponseFields    types.List   `tfsdk:"ignore_response_fields"`
}

func (r *CurlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (r *CurlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TerraCurl request resource",
		Version:             resourceSchemaV1,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Friendly name for this API call",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Api endpoint to call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"method": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "HTTP method to use in the API call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A request body to attach to the API call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the API call",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of parameters to attach to the API call",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"request_url_string": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Request URL includes parameters if request specified",
			},
			"cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded certificate to present to the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that will be used to validate the certificate presented by the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ca_cert_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"skip_tls_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set this to true to disable verification of the server's TLS certificate",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"retry_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interval between each attempt",
				Default:             int64default.StaticInt64(10),
			},
			"max_retry": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tries until it is marked as failed",
			},
			"timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time in seconds before each request times out. Defaults to 10",
				Default:             int64default.StaticInt64(10),
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

			"skip_destroy": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set this to true to skip issuing a request when the resource is being destroyed",
				Default:             booldefault.StaticBool(true),
			},

			"destroy_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Destroy API endpoint to call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Destroy HTTP method to use in the API call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A request body to attach to the destroy API call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the destroy API call",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"destroy_parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of parameters to attach to the destroy API call",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"destroy_request_url_string": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Destroy request URL includes parameters if request specified",
			},
			"destroy_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded certificate to present to the server for the destroy call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_key_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued for the destroy call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a file on local disk that will be used to validate the certificate presented by the server for the destroy call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_ca_cert_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server for the destroy call",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destroy_skip_tls_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Set this to true to disable verification of the server's TLS certificate for the destroy call",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"destroy_retry_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interval between each attempt for the destroy call",
				Default:             int64default.StaticInt64(10),
			},
			"destroy_max_retry": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of tries until it is marked as failed for the destroy call",
			},
			"destroy_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time in seconds before each request times out for the destroy call. Defaults to 10",
				Default:             int64default.StaticInt64(10),
			},
			"destroy_response_codes": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "A list of expected response codes for the destroy call",
				ElementType:         types.StringType,
			},
			"skip_read": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set to true to skip the read operation (no drift detection). Defaults to true.",
				Default:             booldefault.StaticBool(true),
			},
			"read_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "API endpoint for reading resource state. Required if `skip_read` is false.",
			},

			"read_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "HTTP method for reading resource state. Required if `skip_read` is false.",
			},

			"read_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers for the read request.",
			},

			"read_request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional request body to use for the read request.",
			},

			"read_parameters": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Optional request parameters to add to the URL",
			},

			"read_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a PEM-encoded certificate for the read request (TLS).",
			},

			"read_key_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a PEM-encoded private key for the read request (TLS).",
			},

			"read_ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a PEM-encoded CA certificate for the read request (TLS).",
			},

			"read_ca_cert_directory": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a PEM-encoded CA certificate for the read request (TLS).",
			},

			"read_skip_tls_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Skip TLS verification for the read request.",
			},
			"read_response_codes": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "Expected response codes for the read request. Required if `skip_read` is false.",
				ElementType:         types.StringType,
			},
			"drift_marker": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Marker to track state drift and trigger resource replacement",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ignore_response_fields": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "List of JSON fields to ignore during drift detection.",
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *CurlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *CurlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CurlResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.SkipRead.IsNull() && !data.SkipRead.ValueBool() {
		tflog.Debug(ctx, "skip_read validation triggered", map[string]interface{}{
			"skip_read_is_null":   data.SkipRead.IsNull(),
			"skip_read_value":     data.SkipRead.ValueBool(),
			"read_url_is_null":    data.ReadUrl.IsNull(),
			"read_method_is_null": data.ReadMethod.IsNull(),
			"response_codes_null": data.ReadResponseCodes.IsNull(),
		})
		if data.ReadUrl.IsNull() || data.ReadMethod.IsNull() || data.ReadResponseCodes.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"If `skip_read` is set to `false`, `read_url`, `read_method`, and `read_response_codes` must be provided.",
			)
			return
		}
	}

	if !data.SkipDestroy.IsNull() && !data.SkipDestroy.ValueBool() {
		if data.DestroyUrl.IsNull() || data.DestroyMethod.IsNull() || data.DestroyResponseCodes.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"If `skip_destroy` is set to `false`, `destroy_url`, `destroy_method`, and `destroy_response_codes` must be provided.",
			)
			return
		}
	}

	if data.DestroyUrl.IsNull() {
		data.DestroyRequestUrlString = types.StringValue("")
	} else {
		data.DestroyRequestUrlString = types.StringValue(data.DestroyUrl.ValueString())
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
	if !data.Parameters.IsNull() && !data.Parameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.Parameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}
	data.RequestUrlString = types.StringValue(request.URL.String())

	tflog.Debug(ctx, fmt.Sprintf("Resource create API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, data.RequestBody.ValueString()))
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

	data.DriftMarker = types.StringValue("initial")
	data.DestroyRequestUrlString = types.StringValue(data.DestroyUrl.ValueString())
	data.RequestUrlString = types.StringValue(request.URL.String())
	data.Response = types.StringValue(bodyString)
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CurlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CurlResourceModel

	// Load prior state
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip read if configured
	if data.SkipRead.ValueBool() {
		tflog.Debug(ctx, "Skipping Read() as skip_read is true")
		return
	}
	// ======= Validate Required Read Arguments =======
	if data.ReadUrl.IsNull() || data.ReadMethod.IsNull() || data.ReadResponseCodes.IsNull() {
		resp.Diagnostics.AddError(
			"Read Configuration Error",
			"`read_url`, `read_method`, and `read_response_codes` are required when `skip_read` is false.",
		)
		return
	}

	// ======= Build TLS Client if `read_*` TLS Arguments Provided =======
	var client *http.Client
	useReadTls := !data.ReadCertFile.IsNull() || !data.ReadKeyFile.IsNull() || !data.ReadCaCertFile.IsNull()

	if useReadTls {
		tflog.Debug(ctx, "Using custom TLS client for Read() operation")

		readTlsConfig := &TlsConfig{
			CertFile:        data.ReadCertFile.ValueString(),
			KeyFile:         data.ReadKeyFile.ValueString(),
			CaCertFile:      data.ReadCaCertFile.ValueString(),
			CaCertDirectory: data.ReadCaCertDirectory.ValueString(),
			SkipTlsVerify:   data.ReadSkipTlsVerify.ValueBool(),
		}

		tlsClient, err := createTlsClient(readTlsConfig)
		if err != nil {
			resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Failed to create TLS client: %s", err))
			return
		}
		client = tlsClient
	} else {
		// Default non-TLS client
		tflog.Debug(ctx, "Using default HTTP client for Read() operation")
		client = &http.Client{Timeout: 30 * time.Second}
	}

	// ======= Build Read Request =======
	var reqBody io.Reader = nil
	if !data.ReadRequestBody.IsNull() && !data.ReadRequestBody.IsUnknown() {
		reqBody = bytes.NewBuffer([]byte(data.ReadRequestBody.ValueString()))
	}

	request, err := http.NewRequest(data.ReadMethod.ValueString(), data.ReadUrl.ValueString(), reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Failed to create request: %s", err))
		return
	}

	// ======= Add Headers =======
	if !data.ReadHeaders.IsNull() && !data.ReadHeaders.IsUnknown() {
		for k, v := range data.ReadHeaders.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	// ======= Add Query Parameters =======
	if !data.ReadParameters.IsNull() && !data.ReadParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.ReadParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}

	// ======= Execute Request =======
	tflog.Debug(ctx, fmt.Sprintf("Resource read API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, data.RequestBody.ValueString()))

	httpResp, err := client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("Failed to call API: %s", err))
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(httpResp.Body)

	// Read and store the response
	bodyBytes, _ := io.ReadAll(httpResp.Body)
	newResponse := string(bodyBytes)

	// ===== DRIFT DETECTION =====

	var ignoredFields []string
	for _, v := range data.IgnoreResponseFields.Elements() {
		if strVal, ok := v.(types.String); ok {
			ignoredFields = append(ignoredFields, strVal.ValueString())
		}
	}

	sanitizedResponse, err := sanitizeResponse(newResponse, ignoredFields)
	if err != nil {
		resp.Diagnostics.AddError("Sanitize Error", fmt.Sprintf("Failed to sanitize stored response: %s", err))
		return
	}

	// Compare old and new sanitized responses
	oldSanitized, err := sanitizeResponse(data.Response.ValueString(), ignoredFields)
	if err != nil {
		resp.Diagnostics.AddError("Sanitize Error", fmt.Sprintf("Failed to sanitize prior response: %s", err))
		return
	}

	// Drift detection
	if oldSanitized != sanitizedResponse {
		tflog.Warn(ctx, "Drift detected: Response has changed, marking for recreation.")
		data.DriftMarker = types.StringValue(time.Now().Format(time.RFC3339Nano))
	} else {
		// Set an initial drift marker if none exists
		if data.DriftMarker.IsNull() || data.DriftMarker.IsUnknown() {
			data.DriftMarker = types.StringValue("initial")
		}
	}

	// Store the new sanitized response
	data.Response = types.StringValue(sanitizedResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CurlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CurlResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CurlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CurlResourceModel

	// Read prior state into model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip Destroy if `skip_destroy` is true
	if data.SkipDestroy.ValueBool() {
		tflog.Debug(ctx, "Skipping Destroy() because skip_destroy is set to true")
		return
	}

	// Validate Required Destroy Arguments
	if data.DestroyUrl.IsNull() || data.DestroyMethod.IsNull() || data.DestroyResponseCodes.IsNull() {
		resp.Diagnostics.AddError(
			"Destroy Configuration Error",
			"`destroy_url`, `destroy_method`, and `destroy_response_codes` are required when `skip_destroy` is false.",
		)
		return
	}

	// Build TLS Client if `destroy_*` TLS Arguments Provided
	var client *http.Client
	useDestroyTls := !data.DestroyCertFile.IsNull() || !data.DestroyKeyFile.IsNull() || !data.DestroyCaCertFile.IsNull()

	if useDestroyTls {
		tflog.Debug(ctx, "Using custom TLS client for Destroy() operation")

		destroyTlsConfig := &TlsConfig{
			CertFile:        data.DestroyCertFile.ValueString(),
			KeyFile:         data.DestroyKeyFile.ValueString(),
			CaCertFile:      data.DestroyCaCertFile.ValueString(),
			CaCertDirectory: data.DestroyCaCertDirectory.ValueString(),
			SkipTlsVerify:   data.DestroySkipTlsVerify.ValueBool(),
		}

		tlsClient, err := createTlsClient(destroyTlsConfig)
		if err != nil {
			resp.Diagnostics.AddError("Destroy Error", fmt.Sprintf("Failed to create TLS client: %s", err))
			return
		}
		client = tlsClient
	} else {
		// Default non-TLS client
		tflog.Debug(ctx, "Using default HTTP client for Destroy() operation")
		client = &http.Client{Timeout: 30 * time.Second}
	}

	// Build Destroy Request
	var reqBody io.Reader = nil
	if !data.DestroyRequestBody.IsNull() && !data.DestroyRequestBody.IsUnknown() {
		reqBody = bytes.NewBuffer([]byte(data.DestroyRequestBody.ValueString()))
	}

	request, err := http.NewRequest(data.DestroyMethod.ValueString(), data.DestroyUrl.ValueString(), reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Destroy Error", fmt.Sprintf("Failed to create request: %s", err))
		return
	}

	// Add Headers
	if !data.DestroyHeaders.IsNull() && !data.DestroyHeaders.IsUnknown() {
		for k, v := range data.DestroyHeaders.Elements() {
			if strVal, ok := v.(types.String); ok {
				request.Header.Set(k, strVal.ValueString())
			}
		}
	}

	// Add Query Parameters
	if !data.DestroyParameters.IsNull() && !data.DestroyParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.DestroyParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}

	// Execute Request with Retry Logic
	timeout := time.Duration(data.DestroyTimeout.ValueInt64()) * time.Second
	retryInterval := time.Duration(data.DestroyRetryInterval.ValueInt64()) * time.Second
	maxRetry := int(data.DestroyMaxRetry.ValueInt64())

	tflog.Debug(ctx, fmt.Sprintf("Resource destroy API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, data.RequestBody.ValueString()))

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
				tflog.Warn(ctx, fmt.Sprintf("Destroy request failed, retrying... Attempt %d/%d", retryCount, maxRetry))
				time.Sleep(retryInterval)
				continue
			}
			resp.Diagnostics.AddError("Destroy Error", fmt.Sprintf("Request failed: %s", err))
			return
		}

		bodyBytes, _ = io.ReadAll(response.Body)
		statusCode = response.StatusCode
		err = response.Body.Close()
		if err != nil {
			return
		}

		var expectedCodes []string
		for _, v := range data.DestroyResponseCodes.Elements() {
			if strVal, ok := v.(types.String); ok {
				expectedCodes = append(expectedCodes, strVal.ValueString())
			}
		}

		// Validate Response Code
		if responseCodeChecker(expectedCodes, strconv.Itoa(statusCode)) {
			tflog.Debug(ctx, "Destroy request completed successfully")
			break
		} else {
			if retryCount < maxRetry {
				retryCount++
				tflog.Warn(ctx, fmt.Sprintf(
					"Destroy request returned unexpected status %d. Retrying (%d/%d)",
					statusCode, retryCount, maxRetry,
				))
				time.Sleep(retryInterval)
				continue
			}
			resp.Diagnostics.AddError(
				"Destroy Error",
				fmt.Sprintf("Unexpected response code from destroy request: %d. Response: %s", statusCode, string(bodyBytes)),
			)
			return
		}
	}

	data.DestroyRequestUrlString = types.StringValue(data.DestroyUrl.ValueString())
	data.RequestUrlString = types.StringValue(request.URL.String())
	data.Response = types.StringValue(string(bodyBytes))
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	// Remove Resource from State
	resp.State.RemoveResource(ctx)
	tflog.Debug(ctx, "Resource removed from state after successful destroy")
}

func (r *CurlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r CurlResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("cert_file"),
			path.MatchRoot("key_file"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("ca_cert_file"),
			path.MatchRoot("ca_cert_directory"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("destroy_cert_file"),
			path.MatchRoot("destroy_key_file"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("destroy_ca_cert_file"),
			path.MatchRoot("destroy_ca_cert_directory"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("read_url"),
			path.MatchRoot("read_method"),
			path.MatchRoot("read_response_codes"),
		),
	}
}

//func (r *CurlResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
//	return map[int64]resource.StateUpgrader{
//		resourceSchemaV0: {
//			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
//				tflog.Debug(ctx, "Beginning state upgrade from v0 to v1")
//
//				var oldState CurlResourceModel
//				diags := req.State.Get(ctx, &oldState)
//				resp.Diagnostics.Append(diags...)
//				if resp.Diagnostics.HasError() {
//					return
//				}
//
//				// Set skip_read to true and clear read-related fields
//				oldState.SkipRead = types.BoolValue(true)
//				oldState.ReadUrl = types.StringNull()
//				oldState.ReadMethod = types.StringNull()
//				oldState.ReadHeaders = types.MapNull(types.StringType)
//				oldState.ReadParameters = types.MapNull(types.StringType)
//				oldState.ReadRequestBody = types.StringNull()
//				oldState.ReadCertFile = types.StringNull()
//				oldState.ReadKeyFile = types.StringNull()
//				oldState.ReadCaCertFile = types.StringNull()
//				oldState.ReadCaCertDirectory = types.StringNull()
//				oldState.ReadSkipTlsVerify = types.BoolNull()
//				oldState.ReadResponseCodes = types.ListNull(types.StringType)
//
//				// Set the upgraded state
//				diags = resp.State.Set(ctx, oldState)
//				resp.Diagnostics.Append(diags...)
//				tflog.Debug(ctx, "Completed state upgrade from v0 to v1")
//			},
//		},
//	}
//}

func (r *CurlResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				// Check if the state is nil
				if req.State == nil {
					resp.Diagnostics.AddError(
						"Unable to Upgrade Resource State",
						"State is nil, cannot perform upgrade.",
					)
					return
				}

				var oldState CurlResourceModel
				diags := req.State.Get(ctx, &oldState)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}

				// Create new state with the same schema
				newState := CurlResourceModel{
					// Preserve required and important fields
					Id:     oldState.Id,
					Name:   oldState.Name,
					Url:    oldState.Url,
					Method: oldState.Method,

					// Initialize maps with non-nil values or empty maps
					Headers:        oldState.Headers,
					Parameters:     oldState.Parameters,
					DestroyHeaders: oldState.DestroyHeaders,

					// Initialize lists with non-nil values or empty lists
					ResponseCodes:        oldState.ResponseCodes,
					DestroyResponseCodes: oldState.DestroyResponseCodes,
					IgnoreResponseFields: oldState.IgnoreResponseFields,

					// Preserve or set to null other fields
					RequestBody:      oldState.RequestBody,
					CertFile:         types.StringNull(),
					KeyFile:          types.StringNull(),
					CaCertFile:       types.StringNull(),
					CaCertDirectory:  types.StringNull(),
					SkipTlsVerify:    types.BoolNull(),
					Timeout:          types.Int64Null(),
					MaxRetry:         types.Int64Null(),
					RetryInterval:    types.Int64Null(),
					StatusCode:       types.StringNull(),
					Response:         types.StringNull(),
					RequestUrlString: types.StringNull(),
					DriftMarker:      types.StringNull(),

					// Clear read-related fields
					ReadUrl:             types.StringNull(),
					ReadMethod:          types.StringNull(),
					ReadHeaders:         types.MapValueMust(types.StringType, map[string]attr.Value{}),
					ReadParameters:      types.MapValueMust(types.StringType, map[string]attr.Value{}),
					ReadRequestBody:     types.StringNull(),
					ReadResponseCodes:   types.ListValueMust(types.StringType, []attr.Value{}),
					ReadCertFile:        types.StringNull(),
					ReadKeyFile:         types.StringNull(),
					ReadCaCertFile:      types.StringNull(),
					ReadCaCertDirectory: types.StringNull(),
					ReadSkipTlsVerify:   types.BoolNull(),

					// Initialize destroy-related fields
					DestroyUrl:              oldState.DestroyUrl,
					DestroyMethod:           oldState.DestroyMethod,
					DestroyParameters:       oldState.DestroyParameters,
					DestroyRequestBody:      oldState.DestroyRequestBody,
					DestroyTimeout:          types.Int64Null(),
					DestroyMaxRetry:         types.Int64Null(),
					DestroyRetryInterval:    types.Int64Null(),
					DestroyRequestUrlString: types.StringNull(),
					DestroyCertFile:         types.StringNull(),
					DestroyKeyFile:          types.StringNull(),
					DestroyCaCertFile:       types.StringNull(),
					DestroyCaCertDirectory:  types.StringNull(),
					DestroySkipTlsVerify:    types.BoolNull(),
				}

				// Set the new state
				diags = resp.State.Set(ctx, &newState)
				resp.Diagnostics.Append(diags...)
			},
		},
	}
}
