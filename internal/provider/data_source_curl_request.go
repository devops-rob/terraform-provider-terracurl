package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sethvargo/go-retry"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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
				Default:     0,
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
		},
	}
}

func dataSourceCurlRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	retryCount := 0
	var lastError error
	err = retry.Constant(ctx, time.Duration(retryInterval)*time.Second, func(ctx context.Context) error {
		if ctx.Err() != nil {
			return fmt.Errorf("context canceled, not retrying operation: %s", lastError)
		}

		if retryCount > maxRetry {
			return fmt.Errorf("request failed, retries exceeded: %s", lastError)
		}

		var err error
		var resp *http.Response
		resp, err = Client.Do(request)
		if err != nil {
			retryCount++
			lastError = err
			return retry.RetryableError(err)
		}

		body, _ = ioutil.ReadAll(resp.Body)
		code := strconv.Itoa(resp.StatusCode)

		if !responseCodeChecker(stringConversionList, code) {
			retryCount++
			lastError = fmt.Errorf("%s response received: %s", code, body)
			return retry.RetryableError(lastError)
		}

		return nil
	})

	if err != nil {
		return diag.Errorf("unable to make request: %s", err)
	}

	respBody.responseBody = string(body)

	d.Set("response", string(body))
	return diags
}
