package provider

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCurl() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Sample resource in the Terraform provider scaffolding.",

		CreateContext: resourceCurlCreate,
		ReadContext:   resourceCurlRead,
		UpdateContext: resourceCurlUpdate,
		DeleteContext: resourceCurlDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Friendly name for this API call",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"url": {
				// This description is used by the documentation generator and the language server.
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
			"destroy_url": {
				// This description is used by the documentation generator and the language server.
				Description: "Api endpoint to call",
				Type:        schema.TypeString,
				Required:    true,
			},
			"destroy_method": {
				Description: "HTTP method to use in the API call",
				Type:        schema.TypeString,
				Required:    true,
			},
			"destroy_request_body": {
				Description: "A request body to attach to the API call",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destroy_headers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of headers to attach to the API call",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type RequestData struct {
	Url         string
	Method      string
	RequestBody []byte
	Headers     map[string]string
}

func defaultRequestData() RequestData {
	return RequestData{
		Url:         "",
		Method:      "",
		RequestBody: nil,
		Headers:     nil,
	}
}

func resourceCurlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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
	body, err := ioutil.ReadAll(resp.Body)

	d.Set("response", body)

	// write logs using the tflog package
	// see https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog
	// for more information
	tflog.Trace(ctx, "created a resource")

	return diags
}

func resourceCurlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceCurlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceCurlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	var diags diag.Diagnostics
	id := d.Get("name").(string)
	d.SetId(id)

	req := defaultRequestData()

	req.Url = d.Get("destroy_url").(string)
	req.Method = d.Get("destroy_method").(string)
	if reqBody, ok := d.Get("destroy_request_body").(string); ok {
		var jsonStr = []byte(reqBody)
		req.RequestBody = jsonStr
	}

	request, err := http.NewRequestWithContext(context.TODO(), req.Method, req.Url, bytes.NewBuffer(req.RequestBody))
	if err != nil {
		diag.FromErr(err)
	}

	if requestHeaders, ok := d.Get("destroy_headers").(map[string]interface{}); ok {
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
	_, err = ioutil.ReadAll(resp.Body)

	return diags
}
