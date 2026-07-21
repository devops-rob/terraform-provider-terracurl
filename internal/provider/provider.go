// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TerraCurlProvider satisfies various provider interfaces.
var _ provider.Provider = &TerraCurlProvider{}
var _ provider.ProviderWithFunctions = &TerraCurlProvider{}
var _ provider.ProviderWithActions = &TerraCurlProvider{}

// TerraCurlProvider defines the provider implementation.
type TerraCurlProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TerraCurlProviderModel describes the provider data model.
type TerraCurlProviderModel struct {
	HttpProxy  types.String `tfsdk:"http_proxy"`
	HttpsProxy types.String `tfsdk:"https_proxy"`
	NoProxy    types.String `tfsdk:"no_proxy"`
}

func (p *TerraCurlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "terracurl"
	resp.Version = p.version
}

func (p *TerraCurlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The TerraCurl provider allows you to make custom HTTP requests in Terraform.",
		Attributes: map[string]schema.Attribute{
			"http_proxy": schema.StringAttribute{
				MarkdownDescription: "Proxy URL for HTTP requests. Overrides the `HTTP_PROXY` environment variable when set.",
				Optional:            true,
			},
			"https_proxy": schema.StringAttribute{
				MarkdownDescription: "Proxy URL for HTTPS requests. Overrides the `HTTPS_PROXY` environment variable when set.",
				Optional:            true,
			},
			"no_proxy": schema.StringAttribute{
				MarkdownDescription: "Comma-separated list of hosts that should bypass the proxy. Overrides the `NO_PROXY` environment variable when set.",
				Optional:            true,
			},
		},
	}
}

func (p *TerraCurlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TerraCurlProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	meta := NewProviderMeta(data.HttpProxy, data.HttpsProxy, data.NoProxy)
	resp.DataSourceData = meta
	resp.ResourceData = meta
	resp.EphemeralResourceData = meta
	resp.ActionData = meta
}

func (p *TerraCurlProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		NewCurlAction,
	}
}

func (p *TerraCurlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCurlResource,
	}
}

func (p *TerraCurlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCurlDataSource,
	}
}

func (p *TerraCurlProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TerraCurlProvider{
			version: version,
		}
	}
}
