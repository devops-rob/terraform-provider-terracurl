package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	resourceSchemaV0 = 0
	resourceSchemaV1 = 1
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CurlResource{}
var _ resource.ResourceWithImportState = &CurlResource{}
var _ resource.ResourceWithModifyPlan = &CurlResource{}

func NewCurlResource() resource.Resource {
	return &CurlResource{}
}

// CurlResource defines the resource implementation.
type CurlResource struct {
	meta *ProviderMeta
}

// CurlResourceModel describes the resource data model.
type CurlResourceModel struct {
	Id                          types.String          `tfsdk:"id"`
	Name                        types.String          `tfsdk:"name"`
	Url                         types.String          `tfsdk:"url"`
	Method                      types.String          `tfsdk:"method"`
	RequestBody                 types.String          `tfsdk:"request_body"`
	RequestBodyWo               types.String          `tfsdk:"request_body_wo"`
	RequestBodyWoVersion        types.Int64           `tfsdk:"request_body_wo_version"`
	RequestBodyFile             types.String          `tfsdk:"request_body_file"`
	RequestMultipart            *MultipartConfigModel `tfsdk:"request_multipart"`
	Headers                     types.Map             `tfsdk:"headers"`
	HeadersWo                   types.Map             `tfsdk:"headers_wo"`
	HeadersWoVersion            types.Int64           `tfsdk:"headers_wo_version"`
	DigestAuth                  *DigestAuthModel      `tfsdk:"digest_auth"`
	RequestParameters           types.Map             `tfsdk:"request_parameters"`
	RequestUrlString            types.String          `tfsdk:"request_url_string"`
	CertFile                    types.String          `tfsdk:"cert_file"`
	KeyFile                     types.String          `tfsdk:"key_file"`
	CaCertFile                  types.String          `tfsdk:"ca_cert_file"`
	CaCertDirectory             types.String          `tfsdk:"ca_cert_directory"`
	SkipTlsVerify               types.Bool            `tfsdk:"skip_tls_verify"`
	RetryInterval               types.Int64           `tfsdk:"retry_interval"`
	MaxRetry                    types.Int64           `tfsdk:"max_retry"`
	Timeout                     types.Int64           `tfsdk:"timeout"`
	Response                    types.String          `tfsdk:"response"`
	SensitiveResponse           types.String          `tfsdk:"sensitive_response"`
	ResponseSensitive           types.Bool            `tfsdk:"response_sensitive"`
	ResponseCodes               types.List            `tfsdk:"response_codes"`
	StatusCode                  types.String          `tfsdk:"status_code"`
	SkipDestroy                 types.Bool            `tfsdk:"skip_destroy"`
	DestroyUrl                  types.String          `tfsdk:"destroy_url"`
	DestroyMethod               types.String          `tfsdk:"destroy_method"`
	DestroyRequestBody          types.String          `tfsdk:"destroy_request_body"`
	DestroyRequestBodyWo        types.String          `tfsdk:"destroy_request_body_wo"`
	DestroyRequestBodyWoVersion types.Int64           `tfsdk:"destroy_request_body_wo_version"`
	DestroyRequestBodyFile      types.String          `tfsdk:"destroy_request_body_file"`
	DestroyRequestMultipart     *MultipartConfigModel `tfsdk:"destroy_request_multipart"`
	DestroyHeaders              types.Map             `tfsdk:"destroy_headers"`
	DestroyHeadersWo            types.Map             `tfsdk:"destroy_headers_wo"`
	DestroyHeadersWoVersion     types.Int64           `tfsdk:"destroy_headers_wo_version"`
	DestroyDigestAuth           *DigestAuthModel      `tfsdk:"destroy_digest_auth"`
	DestroyRequestParameters    types.Map             `tfsdk:"destroy_request_parameters"`
	DestroyRequestUrlString     types.String          `tfsdk:"destroy_request_url_string"`
	DestroyCertFile             types.String          `tfsdk:"destroy_cert_file"`
	DestroyKeyFile              types.String          `tfsdk:"destroy_key_file"`
	DestroyCaCertFile           types.String          `tfsdk:"destroy_ca_cert_file"`
	DestroyCaCertDirectory      types.String          `tfsdk:"destroy_ca_cert_directory"`
	DestroySkipTlsVerify        types.Bool            `tfsdk:"destroy_skip_tls_verify"`
	DestroyRetryInterval        types.Int64           `tfsdk:"destroy_retry_interval"`
	DestroyMaxRetry             types.Int64           `tfsdk:"destroy_max_retry"`
	DestroyTimeout              types.Int64           `tfsdk:"destroy_timeout"`
	DestroyResponseCodes        types.List            `tfsdk:"destroy_response_codes"`
	SkipRead                    types.Bool            `tfsdk:"skip_read"`
	ReadAfterWrite              types.Bool            `tfsdk:"read_after_write"`
	ReadUrl                     types.String          `tfsdk:"read_url"`
	ReadMethod                  types.String          `tfsdk:"read_method"`
	ReadHeaders                 types.Map             `tfsdk:"read_headers"`
	ReadHeadersWo               types.Map             `tfsdk:"read_headers_wo"`
	ReadHeadersWoVersion        types.Int64           `tfsdk:"read_headers_wo_version"`
	ReadDigestAuth              *DigestAuthModel      `tfsdk:"read_digest_auth"`
	ReadParameters              types.Map             `tfsdk:"read_parameters"`
	ReadRequestBody             types.String          `tfsdk:"read_request_body"`
	ReadRequestBodyWo           types.String          `tfsdk:"read_request_body_wo"`
	ReadRequestBodyWoVersion    types.Int64           `tfsdk:"read_request_body_wo_version"`
	ReadRequestBodyFile         types.String          `tfsdk:"read_request_body_file"`
	ReadRequestMultipart        *MultipartConfigModel `tfsdk:"read_request_multipart"`
	ReadCertFile                types.String          `tfsdk:"read_cert_file"`
	ReadKeyFile                 types.String          `tfsdk:"read_key_file"`
	ReadCaCertFile              types.String          `tfsdk:"read_ca_cert_file"`
	ReadCaCertDirectory         types.String          `tfsdk:"read_ca_cert_directory"`
	ReadSkipTlsVerify           types.Bool            `tfsdk:"read_skip_tls_verify"`
	ReadResponseCodes           types.List            `tfsdk:"read_response_codes"`
	DriftMarker                 types.String          `tfsdk:"drift_marker"`
	IgnoreResponseFields        types.List            `tfsdk:"ignore_response_fields"`
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
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("request_body_wo")),
				},
			},
			"request_body_wo":         writeOnlyBodySchema("Write-only request body for the create call. Not stored in Terraform state. Requires Terraform 1.11 or later."),
			"request_body_wo_version": writeOnlyVersionSchema("Increment to trigger applying an updated `request_body_wo` value."),
			"request_body_file":       resourceRequestBodyFileSchema(requestBodyFileDescription + " Used for the create call."),
			"request_multipart":       resourceMultipartSchema(requestMultipartDescription + " Used for the create call."),
			"headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the API call." + hostHeaderMarkdownSuffix,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Map{
					mapvalidator.PreferWriteOnlyAttribute(path.MatchRoot("headers_wo")),
				},
			},
			"headers_wo":         writeOnlyHeadersSchema("Write-only headers for the create call. Not stored in Terraform state. Requires Terraform 1.11 or later."),
			"headers_wo_version": writeOnlyVersionSchema("Increment to trigger applying updated `headers_wo` values."),
			"digest_auth":        resourceDigestAuthSchema("HTTP Digest authentication credentials for the create request. Overrides provider `default_digest_auth` when configured."),
			"request_parameters": schema.MapAttribute{
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
				Default:             booldefault.StaticBool(false),
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
				MarkdownDescription: "Destroy API endpoint to call. Supports `{response.<path>}` placeholders resolved from the stored create response at destroy time. See the [Destroy Response Templating guide](../guides/destroy_templating).",
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
				MarkdownDescription: "A request body to attach to the destroy API call. Supports `{response.<path>}` placeholders resolved from the stored create response at destroy time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("destroy_request_body_wo")),
				},
			},
			"destroy_request_body_wo":         writeOnlyBodySchema("Write-only request body for the destroy call. Not stored in Terraform state."),
			"destroy_request_body_wo_version": writeOnlyVersionSchema("Increment to trigger applying an updated `destroy_request_body_wo` value."),
			"destroy_request_body_file":       resourceRequestBodyFileSchema(requestBodyFileDescription + " Used for the destroy call."),
			"destroy_request_multipart":       resourceMultipartSchema(requestMultipartDescription + " Used for the destroy call."),
			"destroy_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of headers to attach to the destroy API call." + hostHeaderMarkdownSuffix + " Values support `{response.<path>}` placeholders resolved from the stored create response at destroy time.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Map{
					mapvalidator.PreferWriteOnlyAttribute(path.MatchRoot("destroy_headers_wo")),
				},
			},
			"destroy_headers_wo":         writeOnlyHeadersSchema("Write-only headers for the destroy call. Snapshotted in provider private state for use during destroy."),
			"destroy_headers_wo_version": writeOnlyVersionSchema("Increment to trigger refreshing snapshotted `destroy_headers_wo` values."),
			"destroy_digest_auth":        resourceDigestAuthSchema("HTTP Digest authentication credentials for the destroy request. Overrides provider `default_digest_auth` when configured."),
			"destroy_request_parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Map of parameters to attach to the destroy API call. Values support `{response.<path>}` placeholders resolved from the stored create response at destroy time.",
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
			"read_after_write": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "When true, immediately after a successful create or update Terraform will " +
					"call the read endpoint (configured via `read_url`, `read_method`, and `read_response_codes`) " +
					"and store the read response in state instead of the create/update response body. " +
					"This is the recommended approach when the API returns a different response shape on " +
					"write versus read (e.g. `{\"Success\":true}` on create but the full resource object on read), " +
					"so that subsequent drift detection compares like-for-like. " +
					"`read_url`, `read_method`, and `read_response_codes` must be configured when this is true. " +
					"Defaults to false.",
				Default: booldefault.StaticBool(false),
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
				MarkdownDescription: "Map of headers for the read request." + hostHeaderMarkdownSuffix,
				Validators: []validator.Map{
					mapvalidator.PreferWriteOnlyAttribute(path.MatchRoot("read_headers_wo")),
				},
			},
			"read_headers_wo":         writeOnlyHeadersSchema("Write-only headers for the read call. Snapshotted in provider private state for drift detection."),
			"read_headers_wo_version": writeOnlyVersionSchema("Increment to trigger refreshing snapshotted `read_headers_wo` values."),
			"read_digest_auth":        resourceDigestAuthSchema("HTTP Digest authentication credentials for the read request. Overrides provider `default_digest_auth` when configured."),

			"read_request_body": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional request body to use for the read request.",
				Validators: []validator.String{
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("read_request_body_wo")),
				},
			},
			"read_request_body_wo":         writeOnlyBodySchema("Write-only request body for the read call. Snapshotted in provider private state for drift detection."),
			"read_request_body_wo_version": writeOnlyVersionSchema("Increment to trigger refreshing snapshotted `read_request_body_wo` values."),
			"read_request_body_file":       resourceRequestBodyFileSchema(requestBodyFileDescription + " Used for the read call."),
			"read_request_multipart":       resourceMultipartSchema(requestMultipartDescription + " Used for the read call."),

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
				MarkdownDescription: "Informational marker updated when remote drift is detected during read. Replacement is planned via ModifyPlan when drift is detected.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
	if req.ProviderData == nil {
		return
	}

	meta, ok := req.ProviderData.(*ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.meta = meta
}

func (r *CurlResource) providerMeta() *ProviderMeta {
	if r.meta != nil {
		return r.meta
	}
	return DefaultProviderMeta()
}

func (r *CurlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CurlResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(mergeWriteOnlyFromConfig(ctx, req.Config, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configModel CurlResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configModel)...)
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

	if !data.ReadAfterWrite.IsNull() && data.ReadAfterWrite.ValueBool() {
		if data.ReadUrl.IsNull() || data.ReadMethod.IsNull() || data.ReadResponseCodes.IsNull() {
			resp.Diagnostics.AddError(
				"Invalid Configuration",
				"If `read_after_write` is set to `true`, `read_url`, `read_method`, and `read_response_codes` must be provided.",
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

	var client *http.Client
	var err error
	var tlsConfig *TlsConfig

	if needsTlsClient(data.CertFile, data.KeyFile, data.CaCertFile, data.CaCertDirectory, data.SkipTlsVerify) {
		tlsConfig = tlsConfigFromAttrs(data.CertFile, data.KeyFile, data.CaCertFile, data.CaCertDirectory, data.SkipTlsVerify)

		if tlsConfig.CertFile != "" && tlsConfig.KeyFile == "" {
			resp.Diagnostics.AddError("Validation Error", "`key_file` must be set if `cert_file` is set.")
			return
		}
	}

	client, err = r.providerMeta().NewHTTPClient(tlsConfig, r.providerMeta().ResolveDigestAuth(data.DigestAuth))
	if err != nil {
		resp.Diagnostics.AddError("HTTP Client Creation Failed", err.Error())
		return
	}

	payload, payloadDiags := resolveRequestPayload(data.RequestBody, data.RequestBodyWo, data.RequestBodyFile, data.RequestMultipart)
	resp.Diagnostics.Append(payloadDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var reqBodyReader io.Reader
	if len(payload.Body) > 0 {
		reqBodyReader = bytes.NewBuffer(payload.Body)
	}
	request, err := http.NewRequest(data.Method.ValueString(), data.Url.ValueString(), reqBodyReader)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Request Creation Failed", err.Error())
		return
	}

	// Add headers
	applyRequestHeadersWithWriteOnly(request, data.Headers, data.HeadersWo, r.providerMeta())
	applyResolvedPayloadToRequest(request, payload)

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

	tflog.Debug(ctx, fmt.Sprintf("Resource create API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, payloadLogLabel(payload, data.RequestBody)))
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

	var ignoredFields []string
	for _, v := range data.IgnoreResponseFields.Elements() {
		if strVal, ok := v.(types.String); ok {
			ignoredFields = append(ignoredFields, strVal.ValueString())
		}
	}

	sanitizedResponse, err := sanitizeResponse(bodyString, ignoredFields)
	if err != nil {
		resp.Diagnostics.AddError("Sanitize Error", fmt.Sprintf("Failed to sanitize response: %s", err))
		return
	}

	data.DriftMarker = types.StringValue("initial")
	data.DestroyRequestUrlString = types.StringValue(data.DestroyUrl.ValueString())
	data.RequestUrlString = types.StringValue(request.URL.String())
	setResourceResponseValues(&data, sanitizedResponse)
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	// read_after_write: immediately after a successful create, call the read endpoint and
	// store its response in state instead of the create response. This is the right approach
	// when the API returns a different body shape on create vs read (e.g. {"Success":true}
	// on create but the full resource object on read), so that subsequent drift detection
	// compares like-for-like.
	if data.ReadAfterWrite.ValueBool() {
		resp.Diagnostics.Append(r.executeReadAfterWrite(ctx, &data, "create")...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.SkipDestroy.ValueBool() {
		resp.Diagnostics.Append(validateDestroyTemplates(data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(snapshotWriteOnlyToPrivate(ctx, configModel, resp.Private)...)
	if resp.Diagnostics.HasError() {
		return
	}
	nullWriteOnlyAttributes(&data)
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

type readDriftResult struct {
	Drifted           bool
	SanitizedResponse string
	StatusCode        int
}

func ignoredResponseFields(data CurlResourceModel) []string {
	var ignoredFields []string
	for _, v := range data.IgnoreResponseFields.Elements() {
		if strVal, ok := v.(types.String); ok {
			ignoredFields = append(ignoredFields, strVal.ValueString())
		}
	}
	return ignoredFields
}

// executeReadAfterWrite calls the configured read endpoint and stores its sanitized response
// in state, replacing the write (create/update) response. It is called after a successful
// create or update when read_after_write = true. The operation label ("create"/"update") is
// used only for log and warning messages.
func (r *CurlResource) executeReadAfterWrite(ctx context.Context, data *CurlResourceModel, operation string) diag.Diagnostics {
	var diags diag.Diagnostics

	if data.ReadUrl.IsNull() || data.ReadMethod.IsNull() || data.ReadResponseCodes.IsNull() {
		diags.AddError(
			"read_after_write Configuration Error",
			fmt.Sprintf("`read_url`, `read_method`, and `read_response_codes` must be configured when `read_after_write` is true (triggered after %s).", operation),
		)
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("read_after_write: executing read after %s to store canonical read response in state", operation))

	readStatusCode, readBody, readDiags := r.executeReadRequest(ctx, *data)
	diags.Append(readDiags...)
	if diags.HasError() {
		return diags
	}

	readBodyString := readBody
	if readBodyString == "" {
		readBodyString = "{}"
	}

	ignoredFields := ignoredResponseFields(*data)
	sanitizedReadResponse, err := sanitizeResponse(readBodyString, ignoredFields)
	if err != nil {
		diags.AddWarning(
			fmt.Sprintf("read_after_write: sanitize warning after %s", operation),
			fmt.Sprintf("Failed to sanitize read response; keeping %s response in state: %s", operation, err),
		)
		return diags
	}

	if !responseCodeChecker(data.ReadResponseCodes, readStatusCode) {
		diags.AddWarning(
			fmt.Sprintf("read_after_write: unexpected status code after %s", operation),
			fmt.Sprintf("Read returned status %d (not in read_response_codes); keeping %s response in state.", readStatusCode, operation),
		)
		return diags
	}

	setResourceResponseValues(data, sanitizedReadResponse)
	data.StatusCode = types.StringValue(strconv.Itoa(readStatusCode))
	tflog.Debug(ctx, fmt.Sprintf("read_after_write: stored read response in state after %s", operation))
	return diags
}

func (r *CurlResource) executeReadRequest(ctx context.Context, data CurlResourceModel) (statusCode int, body string, diags diag.Diagnostics) {
	var readTlsConfig *TlsConfig
	if needsTlsClient(data.ReadCertFile, data.ReadKeyFile, data.ReadCaCertFile, data.ReadCaCertDirectory, data.ReadSkipTlsVerify) {
		tflog.Debug(ctx, "Using custom TLS client for Read() operation")
		readTlsConfig = tlsConfigFromAttrs(data.ReadCertFile, data.ReadKeyFile, data.ReadCaCertFile, data.ReadCaCertDirectory, data.ReadSkipTlsVerify)
	} else {
		tflog.Debug(ctx, "Using default HTTP client for Read() operation")
	}

	client, err := r.providerMeta().NewHTTPClient(readTlsConfig, r.providerMeta().ResolveDigestAuth(data.ReadDigestAuth))
	if err != nil {
		diags.AddError("Read Error", fmt.Sprintf("Failed to create HTTP client: %s", err))
		return
	}

	payload, payloadDiags := resolveRequestPayload(data.ReadRequestBody, data.ReadRequestBodyWo, data.ReadRequestBodyFile, data.ReadRequestMultipart)
	diags.Append(payloadDiags...)
	if diags.HasError() {
		return
	}

	var reqBodyReader io.Reader
	if len(payload.Body) > 0 {
		reqBodyReader = bytes.NewBuffer(payload.Body)
	}

	request, err := http.NewRequest(data.ReadMethod.ValueString(), data.ReadUrl.ValueString(), reqBodyReader)
	if err != nil {
		diags.AddError("Read Error", fmt.Sprintf("Failed to create request: %s", err))
		return
	}

	applyRequestHeadersWithWriteOnly(request, data.ReadHeaders, data.ReadHeadersWo, r.providerMeta())
	applyResolvedPayloadToRequest(request, payload)

	if !data.ReadParameters.IsNull() && !data.ReadParameters.IsUnknown() {
		params := request.URL.Query()
		for k, v := range data.ReadParameters.Elements() {
			if strVal, ok := v.(types.String); ok {
				params.Add(k, strVal.ValueString())
			}
		}
		request.URL.RawQuery = params.Encode()
	}

	tflog.Debug(ctx, fmt.Sprintf("Resource read API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, payloadLogLabel(payload, data.ReadRequestBody)))

	httpResp, err := client.Do(request)
	if err != nil {
		diags.AddError("Read Error", fmt.Sprintf("Failed to call API: %s", err))
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpResp.Body)

	bodyBytes, _ := io.ReadAll(httpResp.Body)
	return httpResp.StatusCode, string(bodyBytes), diags
}

func checkReadDrift(data CurlResourceModel, statusCode int, liveResponse string) (drifted bool, sanitizedLive string, diags diag.Diagnostics) {
	ignoredFields := ignoredResponseFields(data)

	var err error
	sanitizedLive, err = sanitizeResponse(liveResponse, ignoredFields)
	if err != nil {
		diags.AddError("Sanitize Error", fmt.Sprintf("Failed to sanitize live response: %s", err))
		return
	}

	priorResponse := priorResponseValue(data.ResponseSensitive, data.Response, data.SensitiveResponse)
	oldSanitized, err := sanitizeResponse(priorResponse, ignoredFields)
	if err != nil {
		diags.AddError("Sanitize Error", fmt.Sprintf("Failed to sanitize prior response: %s", err))
		return
	}

	if !responseCodeChecker(data.ReadResponseCodes, statusCode) {
		return true, sanitizedLive, diags
	}

	if oldSanitized != "null" && oldSanitized != sanitizedLive {
		return true, sanitizedLive, diags
	}

	return false, sanitizedLive, diags
}

func (r *CurlResource) detectReadDrift(ctx context.Context, data CurlResourceModel) (readDriftResult, diag.Diagnostics) {
	var result readDriftResult
	var diags diag.Diagnostics

	if data.SkipRead.ValueBool() {
		return result, diags
	}

	if data.ReadUrl.IsNull() || data.ReadMethod.IsNull() || data.ReadResponseCodes.IsNull() {
		diags.AddError(
			"Read Configuration Error",
			"`read_url`, `read_method`, and `read_response_codes` are required when `skip_read` is false.",
		)
		return result, diags
	}

	statusCode, body, execDiags := r.executeReadRequest(ctx, data)
	diags.Append(execDiags...)
	if diags.HasError() {
		return result, diags
	}

	result.StatusCode = statusCode
	drifted, sanitizedLive, compareDiags := checkReadDrift(data, statusCode, body)
	diags.Append(compareDiags...)
	if diags.HasError() {
		return result, diags
	}

	result.Drifted = drifted
	result.SanitizedResponse = sanitizedLive
	return result, diags
}

func (r *CurlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CurlResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(loadWriteOnlyFromPrivate(ctx, req.Private, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SkipRead.ValueBool() {
		tflog.Debug(ctx, "Skipping Read() as skip_read is true")
		resp.Private = req.Private
		return
	}

	result, diags := r.detectReadDrift(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if result.Drifted {
		tflog.Warn(ctx, "Drift detected: Response has changed, marking for recreation.")
		data.DriftMarker = types.StringValue(time.Now().Format(time.RFC3339Nano))
	} else {
		if data.DriftMarker.IsNull() || data.DriftMarker.IsUnknown() {
			data.DriftMarker = types.StringValue("initial")
		}
		setResourceResponseValues(&data, result.SanitizedResponse)
	}

	nullWriteOnlyAttributes(&data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Private = req.Private
}

func (r *CurlResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var data CurlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(loadWriteOnlyFromPrivate(ctx, req.Private, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SkipRead.ValueBool() {
		return
	}

	result, diags := r.detectReadDrift(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if result.Drifted {
		var plan CurlResourceModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.DriftMarker = types.StringValue(time.Now().Format(time.RFC3339Nano))
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		resp.RequiresReplace.Append(path.Root("drift_marker"))
		resp.Diagnostics.AddWarning(
			"Remote Drift Detected",
			"The read response does not match stored state. Terraform will replace this resource to reconcile configuration.",
		)
	}

	resp.Private = req.Private
}

func (r *CurlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CurlResourceModel
	var state CurlResourceModel
	var config CurlResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Private = req.Private
	if writeOnlyVersionsChanged(state, config) {
		resp.Diagnostics.Append(snapshotWriteOnlyToPrivate(ctx, config, resp.Private)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// read_after_write: immediately after a successful update, call the read endpoint and
	// store its response in state instead of the plan values.
	if plan.ReadAfterWrite.ValueBool() {
		resp.Diagnostics.Append(r.executeReadAfterWrite(ctx, &plan, "update")...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	nullWriteOnlyAttributes(&plan)
	// Update performs no HTTP write call, so computed attributes the plan left unknown must be
	// resolved before writing state or Terraform rejects the result object. read_after_write may
	// also return early (warning paths) without setting response/status_code.
	resolveUnknownComputedFromState(&plan, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// resolveUnknownComputedFromState carries prior state forward for any computed attribute the
// plan marked "known after apply" but that the update path did not populate.
//
// Only attributes that are purely Computed need handling here. Every Optional+Computed
// attribute on this resource declares a Default, so Terraform resolves those during planning
// and they are never unknown by the time Update runs.
func resolveUnknownComputedFromState(plan *CurlResourceModel, state CurlResourceModel) {
	if plan.Id.IsUnknown() {
		plan.Id = state.Id
	}
	if plan.Response.IsUnknown() {
		plan.Response = state.Response
	}
	if plan.SensitiveResponse.IsUnknown() {
		plan.SensitiveResponse = state.SensitiveResponse
	}
	if plan.StatusCode.IsUnknown() {
		plan.StatusCode = state.StatusCode
	}
	if plan.RequestUrlString.IsUnknown() {
		plan.RequestUrlString = state.RequestUrlString
	}
	if plan.DestroyRequestUrlString.IsUnknown() {
		plan.DestroyRequestUrlString = state.DestroyRequestUrlString
	}
	if plan.DriftMarker.IsUnknown() {
		plan.DriftMarker = state.DriftMarker
	}
}

func (r *CurlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CurlResourceModel

	// Read prior state into model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(loadWriteOnlyFromPrivate(ctx, req.Private, &data)...)
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
	var destroyTlsConfig *TlsConfig
	if needsTlsClient(data.DestroyCertFile, data.DestroyKeyFile, data.DestroyCaCertFile, data.DestroyCaCertDirectory, data.DestroySkipTlsVerify) {
		tflog.Debug(ctx, "Using custom TLS client for Destroy() operation")
		destroyTlsConfig = tlsConfigFromAttrs(data.DestroyCertFile, data.DestroyKeyFile, data.DestroyCaCertFile, data.DestroyCaCertDirectory, data.DestroySkipTlsVerify)
	} else {
		tflog.Debug(ctx, "Using default HTTP client for Destroy() operation")
	}

	client, err := r.providerMeta().NewHTTPClient(destroyTlsConfig, r.providerMeta().ResolveDigestAuth(data.DestroyDigestAuth))
	if err != nil {
		resp.Diagnostics.AddError("Destroy Error", fmt.Sprintf("Failed to create HTTP client: %s", err))
		return
	}

	resolved, templateDiags := resolveDestroyTemplates(&data)
	resp.Diagnostics.Append(templateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build Destroy Request
	var reqBody io.Reader
	if len(resolved.Payload.Body) > 0 {
		reqBody = bytes.NewBuffer(resolved.Payload.Body)
	}

	request, err := http.NewRequest(data.DestroyMethod.ValueString(), resolved.URL, reqBody)
	if err != nil {
		resp.Diagnostics.AddError("Destroy Error", fmt.Sprintf("Failed to create request: %s", err))
		return
	}

	// Add Headers
	applyRequestHeadersWithWriteOnly(request, resolved.DestroyHeaders, resolved.DestroyHeadersWo, r.providerMeta())
	applyResolvedPayloadToRequest(request, resolved.Payload)

	// Add Query Parameters
	if !resolved.DestroyParams.IsNull() && !resolved.DestroyParams.IsUnknown() {
		params := request.URL.Query()
		for k, v := range resolved.DestroyParams.Elements() {
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

	tflog.Debug(ctx, fmt.Sprintf("Resource destroy API Call: \nURL: %s\nHeaders: %s\nMethod: %s\nRequest Body: %s\n", request.URL.String(), request.Header, request.Method, payloadLogLabel(resolved.Payload, data.DestroyRequestBody)))

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

		// Validate Response Code
		if responseCodeChecker(data.DestroyResponseCodes, statusCode) {
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
	setResourceResponseValues(&data, string(bodyBytes))
	data.StatusCode = types.StringValue(strconv.Itoa(statusCode))

	// Remove Resource from State
	resp.State.RemoveResource(ctx)
	tflog.Debug(ctx, "Resource removed from state after successful destroy")
}

func (r *CurlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r CurlResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	validators := []resource.ConfigValidator{
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
		resourcevalidator.Conflicting(path.MatchRoot("headers"), path.MatchRoot("headers_wo")),
		resourcevalidator.Conflicting(path.MatchRoot("request_body"), path.MatchRoot("request_body_wo")),
		resourcevalidator.Conflicting(path.MatchRoot("read_headers"), path.MatchRoot("read_headers_wo")),
		resourcevalidator.Conflicting(path.MatchRoot("read_request_body"), path.MatchRoot("read_request_body_wo")),
		resourcevalidator.Conflicting(path.MatchRoot("destroy_headers"), path.MatchRoot("destroy_headers_wo")),
		resourcevalidator.Conflicting(path.MatchRoot("destroy_request_body"), path.MatchRoot("destroy_request_body_wo")),
		resourcevalidator.RequiredTogether(path.MatchRoot("headers_wo"), path.MatchRoot("headers_wo_version")),
		resourcevalidator.RequiredTogether(path.MatchRoot("request_body_wo"), path.MatchRoot("request_body_wo_version")),
		resourcevalidator.RequiredTogether(path.MatchRoot("read_headers_wo"), path.MatchRoot("read_headers_wo_version")),
		resourcevalidator.RequiredTogether(path.MatchRoot("read_request_body_wo"), path.MatchRoot("read_request_body_wo_version")),
		resourcevalidator.RequiredTogether(path.MatchRoot("destroy_headers_wo"), path.MatchRoot("destroy_headers_wo_version")),
		resourcevalidator.RequiredTogether(path.MatchRoot("destroy_request_body_wo"), path.MatchRoot("destroy_request_body_wo_version")),
	}
	validators = append(validators, resourceBodyValidators(
		"request_body",
		"request_body_wo",
		"request_body_file",
		"request_multipart",
	)...)
	validators = append(validators, resourceBodyValidators(
		"read_request_body",
		"read_request_body_wo",
		"read_request_body_file",
		"read_request_multipart",
	)...)
	validators = append(validators, resourceBodyValidators(
		"destroy_request_body",
		"destroy_request_body_wo",
		"destroy_request_body_file",
		"destroy_request_multipart",
	)...)
	return validators
}

func (r *CurlResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		resourceSchemaV0: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				tflog.Debug(ctx, "Beginning state upgrade from v0 to v1")

				// The main issue is that v0 used "destroy_parameters" but v1 uses "destroy_request_parameters"
				// We need to handle this attribute rename during the upgrade

				var oldState CurlResourceModel

				var diags diag.Diagnostics
				useRawState := req.State == nil

				if !useRawState {
					diags = req.State.Get(ctx, &oldState)
					if diags.HasError() {
						useRawState = true
					}
				}

				if useRawState {
					if req.State == nil {
						tflog.Info(ctx, "Prior state is nil, attempting to extract values from raw state")
					} else {
						tflog.Info(ctx, "Direct state extraction failed, attempting to extract values from raw state", map[string]interface{}{
							"errors": diags.Errors(),
						})
					}

					// Try to extract what we can from the raw state
					if req.RawState != nil && len(req.RawState.JSON) > 0 {
						// Parse the raw JSON state
						var rawStateMap map[string]interface{}
						if err := json.Unmarshal(req.RawState.JSON, &rawStateMap); err != nil {
							tflog.Error(ctx, "Failed to unmarshal raw state JSON", map[string]interface{}{
								"error": err.Error(),
							})
							resp.Diagnostics.AddError("State Upgrade Error", fmt.Sprintf("Failed to parse raw state: %s", err))
							return
						}

						tflog.Debug(ctx, "Successfully parsed raw state, extracting values")

						// Extract core values from raw state
						if id, ok := rawStateMap["id"].(string); ok && id != "" {
							oldState.Id = types.StringValue(id)
						} else {
							oldState.Id = types.StringUnknown()
						}

						if name, ok := rawStateMap["name"].(string); ok && name != "" {
							oldState.Name = types.StringValue(name)
						} else {
							oldState.Name = types.StringUnknown()
						}

						if url, ok := rawStateMap["url"].(string); ok && url != "" {
							oldState.Url = types.StringValue(url)
						} else {
							oldState.Url = types.StringUnknown()
						}

						if method, ok := rawStateMap["method"].(string); ok && method != "" {
							oldState.Method = types.StringValue(method)
						} else {
							oldState.Method = types.StringUnknown()
						}

						// Extract optional string fields
						oldState.RequestBody = extractStringFromRaw(rawStateMap, "request_body")
						oldState.CertFile = extractStringFromRaw(rawStateMap, "cert_file")
						oldState.KeyFile = extractStringFromRaw(rawStateMap, "key_file")
						oldState.CaCertFile = extractStringFromRaw(rawStateMap, "ca_cert_file")
						oldState.CaCertDirectory = extractStringFromRaw(rawStateMap, "ca_cert_directory")
						oldState.DestroyUrl = extractStringFromRaw(rawStateMap, "destroy_url")
						oldState.DestroyMethod = extractStringFromRaw(rawStateMap, "destroy_method")
						oldState.DestroyRequestBody = extractStringFromRaw(rawStateMap, "destroy_request_body")

						// Extract boolean fields
						oldState.SkipTlsVerify = extractBoolFromRaw(rawStateMap, "skip_tls_verify")
						oldState.SkipDestroy = extractBoolFromRaw(rawStateMap, "skip_destroy")

						// Extract int64 fields
						oldState.RetryInterval = extractInt64FromRaw(rawStateMap, "retry_interval")
						oldState.MaxRetry = extractInt64FromRaw(rawStateMap, "max_retry")
						oldState.Timeout = extractInt64FromRaw(rawStateMap, "timeout")
						oldState.DestroyRetryInterval = extractInt64FromRaw(rawStateMap, "destroy_retry_interval")
						oldState.DestroyMaxRetry = extractInt64FromRaw(rawStateMap, "destroy_max_retry")
						oldState.DestroyTimeout = extractInt64FromRaw(rawStateMap, "destroy_timeout")

						// Extract map fields
						oldState.Headers = extractMapFromRaw(rawStateMap, "headers")
						oldState.RequestParameters = extractMapFromRaw(rawStateMap, "request_parameters")
						oldState.DestroyHeaders = extractMapFromRaw(rawStateMap, "destroy_headers")

						// Handle the key attribute rename: destroy_parameters -> destroy_request_parameters
						if destroyParams := extractMapFromRaw(rawStateMap, "destroy_parameters"); !destroyParams.IsNull() {
							oldState.DestroyRequestParameters = destroyParams
							tflog.Debug(ctx, "Migrated destroy_parameters to destroy_request_parameters")
						} else {
							oldState.DestroyRequestParameters = extractMapFromRaw(rawStateMap, "destroy_request_parameters")
						}

						// Extract list fields
						oldState.ResponseCodes = extractListFromRaw(rawStateMap, "response_codes")
						oldState.DestroyResponseCodes = extractListFromRaw(rawStateMap, "destroy_response_codes")
						oldState.IgnoreResponseFields = extractListFromRaw(rawStateMap, "ignore_response_fields")

						// Set remaining computed/null fields
						oldState.RequestUrlString = types.StringNull()
						oldState.Response = types.StringNull()
						oldState.StatusCode = types.StringNull()
						oldState.DestroyRequestUrlString = types.StringNull()
						oldState.DriftMarker = types.StringNull()

					} else {
						tflog.Error(ctx, "No raw state available for manual extraction")
						resp.Diagnostics.AddError("State Upgrade Error", "No raw state available for manual extraction")
						return
					}

					// Clear the extraction errors since we handled them manually
					resp.Diagnostics = diag.Diagnostics{}
				}

				// New fields introduced after v0: default response_sensitive
				// to false and leave sensitive_response empty. This preserves
				// the prior non-sensitive response behavior on upgrade.
				if oldState.ResponseSensitive.IsNull() || oldState.ResponseSensitive.IsUnknown() {
					oldState.ResponseSensitive = types.BoolValue(false)
				}
				if oldState.SensitiveResponse.IsNull() || oldState.SensitiveResponse.IsUnknown() {
					oldState.SensitiveResponse = types.StringValue("")
				}

				// The key change in v1: set skip_read to true and clear read-related fields
				oldState.SkipRead = types.BoolValue(true)
				oldState.ReadUrl = types.StringNull()
				oldState.ReadMethod = types.StringNull()
				oldState.ReadHeaders = types.MapNull(types.StringType)
				oldState.ReadParameters = types.MapNull(types.StringType)
				oldState.ReadRequestBody = types.StringNull()
				oldState.ReadCertFile = types.StringNull()
				oldState.ReadKeyFile = types.StringNull()
				oldState.ReadCaCertFile = types.StringNull()
				oldState.ReadCaCertDirectory = types.StringNull()
				oldState.ReadSkipTlsVerify = types.BoolNull()
				oldState.ReadResponseCodes = types.ListNull(types.StringType)

				nullWriteOnlyAttributes(&oldState)

				// Set the upgraded state
				diags = resp.State.Set(ctx, oldState)
				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					tflog.Error(ctx, "Failed to set upgraded state", map[string]interface{}{
						"errors": resp.Diagnostics.Errors(),
					})
				} else {
					tflog.Debug(ctx, "Successfully completed state upgrade from v0 to v1")
				}
			},
		},
	}
}

// Helper functions for extracting values from raw state during upgrade

func extractStringFromRaw(rawState map[string]interface{}, key string) types.String {
	if value, ok := rawState[key].(string); ok && value != "" {
		return types.StringValue(value)
	}
	return types.StringNull()
}

func extractBoolFromRaw(rawState map[string]interface{}, key string) types.Bool {
	if value, ok := rawState[key].(bool); ok {
		return types.BoolValue(value)
	}
	return types.BoolNull()
}

func extractInt64FromRaw(rawState map[string]interface{}, key string) types.Int64 {
	if value, ok := rawState[key].(float64); ok {
		return types.Int64Value(int64(value))
	}
	if value, ok := rawState[key].(int64); ok {
		return types.Int64Value(value)
	}
	return types.Int64Null()
}

func extractMapFromRaw(rawState map[string]interface{}, key string) types.Map {
	if value, ok := rawState[key].(map[string]interface{}); ok {
		elements := make(map[string]attr.Value)
		for k, v := range value {
			if str, ok := v.(string); ok {
				elements[k] = types.StringValue(str)
			}
		}
		if len(elements) > 0 {
			mapValue, _ := types.MapValue(types.StringType, elements)
			return mapValue
		}
	}
	return types.MapNull(types.StringType)
}

func extractListFromRaw(rawState map[string]interface{}, key string) types.List {
	if value, ok := rawState[key].([]interface{}); ok {
		elements := make([]attr.Value, len(value))
		for i, v := range value {
			if str, ok := v.(string); ok {
				elements[i] = types.StringValue(str)
			}
		}
		if len(elements) > 0 {
			listValue, _ := types.ListValue(types.StringType, elements)
			return listValue
		}
	}
	return types.ListNull(types.StringType)
}
