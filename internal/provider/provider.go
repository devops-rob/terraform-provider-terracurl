package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const TerraformProviderProductUserAgent = "terraform-provider-terracurl"

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
)

func init() {
	Client = &http.Client{}
}

func Provider() *schema.Provider {
	provider := &schema.Provider{

		DataSourcesMap: map[string]*schema.Resource{
			//"waypoint_project":        dataSourceProject(),
			//"waypoint_runner_profile": dataSourceRunnerProfile(),
		},
		ResourcesMap: map[string]*schema.Resource{
			//"waypoint_project":        resourceProject(),
			//"waypoint_runner_profile": resourceRunnerProfile(),
			terracurl_request: resourceCurl(),
		},
	}

	//provider.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	//	config := Config{
	//		Token:        d.Get("token").(string),
	//		WaypointAddr: d.Get("waypoint_addr").(string),
	//	}
	//	return config.Client()
	//}

	return provider
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"scaffolding_data_source": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"curl_request": resourceCurl(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	// Add whatever fields, client or connection info, etc. here
	// you would need to setup to communicate with the upstream
	// API.
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// Setup a User-Agent for your API client (replace the provider name for yours):
		// userAgent := p.UserAgent("terraform-provider-scaffolding", version)
		// TODO: myClient.UserAgent = userAgent

		return &apiClient{}, nil
	}
}
