package provider

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sethvargo/go-retry"
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
			"request_parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Map of parameters to attach to the API call",
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"request_url_string": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Request URL includes parameters if request specified",
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
			"retry_interval": {
				Type:        schema.TypeInt,
				Description: "Interval between each attempt",
				ForceNew:    false,
				Optional:    true,
				Default:     10,
			},
			"max_retry": {
				Type:        schema.TypeInt,
				Description: "Maximum number of tries until it is marked as failed",
				ForceNew:    false,
				Optional:    true,
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
			"status_code": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Response status code received from request",
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
			"destroy_parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Map of parameters to attach to the API call",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"destroy_cert_file": {
				Type:         schema.TypeString, // check this
				Optional:     true,
				Description:  "Path to a file on local disk that contains the PEM-encoded certificate to present to the server",
				ForceNew:     true,
				RequiredWith: []string{"key_file"},
			},
			"destroy_key_file": {
				Type:         schema.TypeString, // check this
				Optional:     true,
				Description:  "Path to a file on local disk that contains the PEM-encoded private key for which the authentication certificate was issued",
				ForceNew:     true,
				RequiredWith: []string{"cert_file"},
			},
			"destroy_ca_cert_file": {
				Type:          schema.TypeString, // check this
				Optional:      true,
				Description:   "Path to a file on local disk that will be used to validate the certificate presented by the server",
				ForceNew:      true,
				ConflictsWith: []string{"ca_cert_directory"},
			},
			"destroy_ca_cert_directory": {
				Type:          schema.TypeString, // check this
				Optional:      true,
				Description:   "Path to a directory on local disk that contains one or more certificate files that will be used to validate the certificate presented by the server",
				ForceNew:      true,
				ConflictsWith: []string{"ca_cert_file"},
			},
			"destroy_skip_tls_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set this to true to disable verification of the Vault server's TLS certificate",
				ForceNew:    true,
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
			"destroy_retry_interval": {
				Type:        schema.TypeInt,
				Description: "Interval between each attempt",
				ForceNew:    true,
				Optional:    true,
				Default:     10,
			},
			"destroy_max_retry": {
				Type:        schema.TypeInt,
				Description: "Maximum number of tries until it is marked as failed",
				ForceNew:    true,
				Optional:    true,
			},
		},
	}
}

type RequestData struct {
	Url               string
	Method            string
	RequestBody       []byte
	Headers           map[string]string
	RequestParameters map[string]string
}

func defaultRequestData() RequestData {
	return RequestData{
		Url:               "",
		Method:            "",
		RequestBody:       nil,
		Headers:           nil,
		RequestParameters: nil,
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
		jsonStr := []byte(reqBody)
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

	ok := false

	retryInterval := 10
	if retryInterval, ok = d.Get("retry_interval").(int); !ok {
		tflog.Warn(ctx, "using default value of 1s for retryInterval")
	}

	maxRetry := 0
	if maxRetry, ok = d.Get("max_retry").(int); !ok {
		tflog.Warn(ctx, "using default value of 0 for maxRetry")
	}

	respCodes := d.Get("response_codes").([]interface{})
	stringConversionList := make([]string, len(respCodes))
	for i, v := range respCodes {
		stringConversionList[i] = fmt.Sprint(v)
	}

	var body []byte
	var code string
	retryCount := 0
	var lastError error
	retryErr := retry.Constant(ctx, time.Duration(retryInterval)*time.Second, func(ctx context.Context) error {
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

			tflog.Trace(ctx, "Request Headers", map[string]interface{}{"Request Headers": request.Header})
		}

		if requestParameters, ok := d.Get("request_parameters").(map[string]interface{}); ok {
			parametersMap := make(map[string]string)

			for k, v := range requestParameters {
				strKey := fmt.Sprintf("%v", k)
				strValue := fmt.Sprintf("%v", v)
				parametersMap[strKey] = strValue
			}

			params := request.URL.Query()
			for k, v := range parametersMap {
				params.Add(k, v)
			}

			request.URL.RawQuery = params.Encode()

			tflog.Trace(ctx, "Request URL", map[string]interface{}{"Request Parameters": request.URL.String()})
		}

		d.Set("request_url_string", request.URL.String())

		if ctx.Err() != nil {
			return fmt.Errorf("context canceled, not retrying operation: %s", lastError)
		}

		if retryCount > maxRetry {
			return fmt.Errorf("request failed, retries exceeded: %s", lastError)
		}

		var resp *http.Response

		tflog.Trace(ctx, "calling endpoint", map[string]interface{}{"url": request.URL.String()})
		resp, err = Client.Do(request)
		if err != nil {
			tflog.Trace(ctx, "call failed, retrying", map[string]interface{}{"error": err})

			retryCount++
			lastError = err
			return retry.RetryableError(err)
		}

		body, _ = ioutil.ReadAll(resp.Body)
		code = strconv.Itoa(resp.StatusCode)

		if !responseCodeChecker(stringConversionList, code) {
			tflog.Trace(ctx, "call failed, retrying",
				map[string]interface{}{
					"statuscode": resp.StatusCode,
					"body":       string(body),
				})

			retryCount++
			lastError = fmt.Errorf("%s response received: %s", code, body)
			return retry.RetryableError(lastError)
		}

		tflog.Trace(ctx, "call succeded", map[string]interface{}{"statuscode": resp.StatusCode})

		return nil
	})

	if retryErr != nil {
		return diag.Errorf("unable to make request: %s", retryErr)
	}

	respBody.responseBody = string(body)

	d.Set("response", string(body))
	d.Set("status_code", code)

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
		jsonStr := []byte(reqBody)
		req.RequestBody = jsonStr
	}

	if destroyCertFile, ok := d.Get("destroy_cert_file").(string); ok {
		tlsConfig.CertFile = destroyCertFile
	}

	if destroyKeyFile, ok := d.Get("destroy_key_file").(string); ok {
		tlsConfig.KeyFile = destroyKeyFile
	}

	if destroyCaCertFile, ok := d.Get("destroy_ca_cert_file").(string); ok {
		tlsConfig.CaCertFile = destroyCaCertFile
	}

	if destroyCaCertDirectory, ok := d.Get("destroy_ca_cert_directory").(string); ok {
		tlsConfig.CaCertDirectory = destroyCaCertDirectory
	}

	if destroySkipTlsVerify, ok := d.Get("destroy_skip_tls_verify").(bool); ok {
		tlsConfig.SkipTlsVerify = destroySkipTlsVerify
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

	var stringConversionList []string

	destroyRespCodes := d.Get("destroy_response_codes").([]interface{})

	if len(destroyRespCodes) == 0 {
		destroyRespCodes = make([]interface{}, 3)
		destroyRespCodes[0] = "405"
		destroyRespCodes[1] = "404"
		destroyRespCodes[2] = "200"
		stringConversionList = make([]string, len(destroyRespCodes))
		for i, v := range destroyRespCodes {
			stringConversionList[i] = fmt.Sprint(v)
		}
	} else if len(destroyRespCodes) > 0 {
		//destroyRespCodes = make([]interface{}, len(destroyRespCodes))
		//strs := []interface{}{"200", "405", "404"}
		//stringConversionList = make([]string, len(strs))
		//for i, v := range strs {
		//	stringConversionList[i] = fmt.Sprint(v)
		//}
		stringConversionList = make([]string, len(destroyRespCodes))
		for i, v := range destroyRespCodes {
			stringConversionList[i] = fmt.Sprint(v)
		}

	}

	ok := false
	retryInterval := 10
	if retryInterval, ok = d.Get("destroy_retry_interval").(int); !ok {
		tflog.Warn(ctx, "using default value of 1s for retryInterval")
	}

	maxRetry := 0
	if maxRetry, ok = d.Get("destroy_max_retry").(int); !ok {
		tflog.Warn(ctx, "using default value of 1 for maxRetry")
	}

	var body []byte
	retryCount := 0
	var lastError error
	retryErr := retry.Constant(ctx, time.Duration(retryInterval)*time.Second, func(ctx context.Context) error {
		if ctx.Err() != nil {
			return fmt.Errorf("context canceled, not retrying operation: %s", lastError)
		}

		if retryCount > maxRetry {
			return fmt.Errorf("request failed, retries exceeded: %s", lastError)
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

		if requestParameters, ok := d.Get("destroy_parameters").(map[string]interface{}); ok {
			parametersMap := make(map[string]string)

			for k, v := range requestParameters {
				strKey := fmt.Sprintf("%v", k)
				strValue := fmt.Sprintf("%v", v)
				parametersMap[strKey] = strValue
			}

			params := request.URL.Query()
			for k, v := range parametersMap {
				params.Add(k, v)
			}

			request.URL.RawQuery = params.Encode()

			tflog.Trace(ctx, "Request URL", map[string]interface{}{"Request Parameters": request.URL.String()})
		}

		var resp *http.Response
		resp, err = Client.Do(request)
		if err != nil {
			tflog.Trace(ctx, "call failed, retrying",
				map[string]interface{}{
					"err": err,
				})

			retryCount++
			lastError = err
			return retry.RetryableError(err)
		}

		body, _ = ioutil.ReadAll(resp.Body)
		code := strconv.Itoa(resp.StatusCode)

		if !responseCodeChecker(stringConversionList, code) {
			tflog.Trace(ctx, "call failed, retrying",
				map[string]interface{}{
					"statuscode": resp.StatusCode,
					"body":       string(body),
				})

			retryCount++
			lastError = err
			return retry.RetryableError(fmt.Errorf("%s response received: %s", code, body))
		}

		return nil
	})

	if retryErr != nil {
		return diag.Errorf("unable to make request: %s", retryErr)
	}

	respBody.responseBody = string(body)

	return diags
}
