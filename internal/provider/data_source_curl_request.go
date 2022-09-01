package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCurlRequest() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Sample data source in the Terraform provider scaffolding.",

		ReadContext: dataSourceCurlRequestRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Friendly name for this API call",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				Description: "Api endpoint to call",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"method": {
				Description: "HTTP method to use in the API call",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"request_body": {
				Description: "A request body to attach to the API call",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Map of headers to attach to the API call",
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"response": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON response received from request",
			},
		},
	}
}

func dataSourceCurlRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	var diags diag.Diagnostics
	id := d.Get("name").(string)
	d.SetId(id)

	req := defaultRequestData()

	req.Url = d.Get("url").(string)
	req.Method = d.Get("method").(string)
	if reqBody, ok := d.Get("request_body").(string); ok {
		var jsonStr = []byte(reqBody)
		req.RequestBody = jsonStr
	}

	request, err := http.NewRequestWithContext(context.TODO(), req.Method, req.Url, bytes.NewBuffer(req.RequestBody))
	if err != nil {
		diag.FromErr(err)
	}

	if requestHeaders, ok := d.Get("headers").(map[string]interface{}); ok {
		headersMap := make(map[string]string)
		for k, v := range requestHeaders {
			strKey := fmt.Sprintf("%v", k)
			strValue := fmt.Sprintf("%v", v)
			headersMap[strKey] = strValue
		}

		for k, v := range headersMap {
			request.Header.Set(k, v)
		}
	}

	resp, err := Client.Do(request)
	if err != nil {
		return diag.FromErr(err)
	}

	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	respBody.responseBody = string(body)

	if err := d.Set("response", string(body)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
