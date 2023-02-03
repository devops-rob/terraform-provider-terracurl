package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io/ioutil"
	"net/http"
)

func resourceCurl() *schema.Resource {
	return &schema.Resource{
		Description: "A flexible Terraform provider for making API calls",

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
			"cert_file": {
				Type:         schema.TypeString, // check this
				Optional:     true,
				Description:  "Path to a file on local disk that contains the PEM-encoded certificate to present to the server",
				ForceNew:     true,
				RequiredWith: []string{"key_file"},
			},
			"key_file": {
				Type:         schema.TypeString, // check this
				Optional:     true,
				Description:  "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued",
				ForceNew:     true,
				RequiredWith: []string{"cert_file"},
			},
			"ca_cert_file": {
				Type:          schema.TypeString, // check this
				Optional:      true,
				Description:   "Path to a file on local disk that will be used to validate the certificate presented by the server",
				ForceNew:      true,
				ConflictsWith: []string{"ca_cert_directory"},
			},
			"ca_cert_directory": {
				Type:          schema.TypeString, // check this
				Optional:      true,
				Description:   "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server",
				ForceNew:      true,
				ConflictsWith: []string{"ca_cert_file"},
			},
			"skip_tls_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set this to true to disable verification of the Vault server's TLS certificate",
				ForceNew:    true,
			},
			"response": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "JSON response received from request",
			},
			"response_codes": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of expected response codes",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"destroy_url": {
				Description: "Api endpoint to call",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destroy_method": {
				Description:  "HTTP method to use in the API call",
				Type:         schema.TypeString,
				RequiredWith: []string{"destroy_url"},
				Optional:     true,
			},
			"destroy_request_body": {
				Description: "A request body to attach to the API call",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destroy_headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "List of headers to attach to the API call",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"destroy_response_codes": {
				Type:         schema.TypeList,
				Optional:     true,
				RequiredWith: []string{"destroy_url"},
				Description:  "A list of expected response codes for destroy operations",
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

type ResponseData struct {
	responseBody string
}

func responseCodeChecker(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var respBody ResponseData

type TlsConfig struct {
	CertFile        string
	KeyFile         string
	CaCertFile      string
	CaCertDirectory string
	SkipTlsVerify   bool
}

func defaultTlsConfig() TlsConfig {
	return TlsConfig{
		CertFile:        "",
		KeyFile:         "",
		CaCertFile:      "",
		CaCertDirectory: "",
		SkipTlsVerify:   true,
	}
}

func resourceCurlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	id := d.Get("name").(string)
	d.SetId(id)

	tlsConfig := defaultTlsConfig()

	req := defaultRequestData()

	req.Url = d.Get("url").(string)
	req.Method = d.Get("method").(string)
	if reqBody, ok := d.Get("request_body").(string); ok {
		var jsonStr = []byte(reqBody)
		req.RequestBody = jsonStr
	}

	if certFile, ok := d.Get("cert_file").(string); ok {
		tlsConfig.CertFile = certFile
	}

	if keyFile, ok := d.Get("key_file").(string); ok {
		tlsConfig.KeyFile = keyFile
	}

	if caCertFile, ok := d.Get("ca_cert_file").(string); ok {
		tlsConfig.CaCertFile = caCertFile
	}

	if caCertDirectory, ok := d.Get("ca_cert_directory").(string); ok {
		tlsConfig.CaCertDirectory = caCertDirectory
	}

	if skipTlsVerify, ok := d.Get("skip_tls_verify").(bool); ok {
		tlsConfig.SkipTlsVerify = skipTlsVerify
	}

	if err := setClient(
		tlsConfig.CertFile,
		tlsConfig.KeyFile,
		tlsConfig.CaCertFile,
		tlsConfig.CaCertDirectory,
		tlsConfig.SkipTlsVerify); err != nil {
		return diag.FromErr(err)
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
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	respCodes := d.Get("response_codes").([]interface{})
	code := fmt.Sprintf("%v", resp.StatusCode)

	stringConversionList := make([]string, len(respCodes))
	for i, v := range respCodes {
		stringConversionList[i] = fmt.Sprint(v)
	}

	if !responseCodeChecker(stringConversionList, code) {
		return diag.Errorf(fmt.Sprintf("%s response received: %s", code, body))
	}

	respBody.responseBody = string(body)

	d.Set("response", string(body))

	tflog.Trace(ctx, "created a resource")

	return diags
}

func resourceCurlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	var diags diag.Diagnostics
	return diags
}

func resourceCurlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	return diags

}

func resourceCurlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	id := d.Get("name").(string)
	d.SetId(id)

	req := defaultRequestData()

	tlsConfig := defaultTlsConfig()

	if reqMethod, ok := d.Get("destroy_method").(string); ok {
		req.Method = reqMethod
	} else {
		reqMethod = "DELETE"
		req.Method = reqMethod
	}

	method := d.Get("destroy_method").(string)
	if len(method) == 0 {
		req.Method = "DELETE"
	}

	if reqBody, ok := d.Get("destroy_request_body").(string); ok {
		var jsonStr = []byte(reqBody)
		req.RequestBody = jsonStr
	}

	if certFile, ok := d.Get("cert_file").(string); ok {
		tlsConfig.CertFile = certFile
	}

	if keyFile, ok := d.Get("key_file").(string); ok {
		tlsConfig.KeyFile = keyFile
	}

	if caCertFile, ok := d.Get("ca_cert_file").(string); ok {
		tlsConfig.CaCertFile = caCertFile
	}

	if caCertDirectory, ok := d.Get("ca_cert_directory").(string); ok {
		tlsConfig.CaCertDirectory = caCertDirectory
	}

	if skipTlsVerify, ok := d.Get("skip_tls_verify").(bool); ok {
		tlsConfig.SkipTlsVerify = skipTlsVerify
	}

	if err := setClient(
		tlsConfig.CertFile,
		tlsConfig.KeyFile,
		tlsConfig.CaCertFile,
		tlsConfig.CaCertDirectory,
		tlsConfig.SkipTlsVerify); err != nil {
		return diag.FromErr(err)
	}

	if reqUrl, ok := d.Get("destroy_url").(string); ok {
		req.Url = reqUrl
	}

	url := d.Get("destroy_url").(string)
	if len(url) == 0 {
		req.Url = "http://example.com"
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
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	destroyRespCodes := d.Get("destroy_response_codes").([]interface{})
	if len(destroyRespCodes) == 0 {
		destroyRespCodes = make([]interface{}, 1)
		destroyRespCodes[0] = "405"

		//if !responseCodeChecker(defaultList, "405") {
		//	return diag.Errorf(fmt.Sprintf("%s response received: %s", code, body))
		//}

	}

	code := fmt.Sprintf("%v", resp.StatusCode)

	if len(destroyRespCodes) > 0 {
		stringConversionList := make([]string, len(destroyRespCodes))
		for i, v := range destroyRespCodes {
			stringConversionList[i] = fmt.Sprint(v)
		}

		if !responseCodeChecker(stringConversionList, code) {
			return diag.Errorf(fmt.Sprintf("%s response received: %s", code, body))
		}
	}

	return diags
}
