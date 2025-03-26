// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0.

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure TerraCurlProvider satisfies various provider interfaces.
var _ provider.Provider = &TerraCurlProvider{}
var _ provider.ProviderWithFunctions = &TerraCurlProvider{}

// TerraCurlProvider defines the provider implementation.
type TerraCurlProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TerraCurlProviderModel describes the provider data model.
type TerraCurlProviderModel struct {
}

func (p *TerraCurlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "terracurl"
	resp.Version = p.version
}

func (p *TerraCurlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The TerraCurl provider allows you to make custom HTTP requests in Terraform.",
	}
}

func (p *TerraCurlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TerraCurlProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.

	// Example client configuration for data sources and resources.
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
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
