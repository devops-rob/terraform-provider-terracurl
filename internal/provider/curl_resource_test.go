package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resource2 "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jarcoal/httpmock"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestAccresourceCurl(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, // Use ProtoV6 Provider
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurl(rName, RequestBody),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("terracurl_request.basic", "name", rName),
					resource.TestCheckResourceAttr("terracurl_request.basic", "url", "https://example.com"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "method", "GET"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "response_codes.#", "3"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "response_codes.0", "200"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "response_codes.1", "201"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "response_codes.2", "204"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "headers.Authorization", "Bearer token"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "headers.Content-Type", "application/json"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "request_body", RequestBody+"\n"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "request_url_string", "https://example.com?id=12345&name=devopsrob"),

					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_method", "POST"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_url", "https://example.com/destroy"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_response_codes.#", "2"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_response_codes.0", "200"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_response_codes.1", "204"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_headers.Authorization", "Bearer token"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_headers.Content-Type", "application/json"),
					resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_request_body", RequestBody+"\n"),

					//resource.TestCheckResourceAttr("terracurl_request.basic", "destroy_request_url_string", "https://example.com/destroy?id=12345&name=devopsrob"),
				),
			},
		},
	})

}

func testAccresourceCurl(name string, requestBody string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "basic" {
  name           = "%s"
  url            = "https://example.com"
  method         = "GET"

  request_body	 = <<EOF
%s
EOF

  headers = {
	Authorization = "Bearer token"
	Content-Type  = "application/json"
  }

  parameters = {
    id 	 = "12345"
	name = "devopsrob"
  }

  response_codes = ["200", "201", "204"]
  destroy_url    = "https://example.com/destroy"
  destroy_method = "POST"
  destroy_response_codes = ["200", "204"]

  destroy_headers = {
	Authorization = "Bearer token"
	Content-Type  = "application/json"
  }

  destroy_parameters = {
    id 	 = "12345"
	name = "devopsrob"
  }

  destroy_request_body	 = <<EOF
%s
EOF


}`, name, requestBody, requestBody)
}

func TestAccresourceCurlDestroy(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/destroy",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testMockEndpointCount("DELETE https://example.com/destroy", 1),
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlDestroy(rName, RequestBody),
			},
			{
				Config: testAccresourceCurlDestroy(rName, RequestBody),
			},
		},
	})

}

func testAccresourceCurlDestroy(name string, requestBody string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "basic" {
  name           = "%s"
  url            = "https://example.com"
  method         = "GET"

  request_body	 = <<EOF
%s
EOF

  headers = {
	Authorization = "Bearer token"
	Content-Type  = "application/json"
  }

  parameters = {
    id 	 = "12345"
	name = "devopsrob"
  }

  
  response_codes = ["200", "201", "204"]
  skip_destroy   = false
  skip_read = true

  destroy_url    = "https://example.com/destroy"
  destroy_method = "DELETE"
  destroy_response_codes = ["200", "204"]

  destroy_headers = {
	Authorization = "Bearer token"
	Content-Type  = "application/json"
  }

  destroy_parameters = {
    id 	 = "12345"
	name = "devopsrob"
  }

  destroy_request_body	 = <<EOF
%s
EOF


}`, name, requestBody, requestBody)
}

func TestAccresourceCurlSkipDestroy(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/destroy",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testMockEndpointCount("DELETE https://example.com/destroy", 0),
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlSkipDestroy(rName, RequestBody),
			},
		},
	})

}

func testAccresourceCurlSkipDestroy(name string, requestBody string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "skip_destroy" {
  name           = "%s"
  url            = "https://example.com"
  method         = "GET"

  request_body	 = <<EOF
%s
EOF

  headers = {
	Authorization = "Bearer token"
	Content-Type  = "application/json"
  }

  parameters = {
    id 	 = "12345"
	name = "devopsrob"
  }

  
  response_codes = ["200", "201", "204"]
  skip_destroy   = true
  skip_read = true

  destroy_url    = "https://example.com/destroy"
  destroy_method = "DELETE"
  destroy_response_codes = ["200", "204"]

  destroy_headers = {
	Authorization = "Bearer token"
	Content-Type  = "application/json"
  }

  destroy_parameters = {
    id 	 = "12345"
	name = "devopsrob"
  }

  destroy_request_body	 = <<EOF
%s
EOF


}`, name, requestBody, requestBody)
}

func TestAccCurlresourceRetries(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	rName := "devopsrob"
	json := `{"name": "` + rName + `"}`

	var firstCall time.Time
	var callCount int

	// Register the responder to simulate a failure followed by a success
	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				firstCall = time.Now()
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil
			}
			return httpmock.NewStringResponse(200, json), nil
		},
	)
	httpmock.RegisterResponder("DELETE", "https://example.com/destroy",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				firstCall = time.Now()
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil
			}
			return httpmock.NewStringResponse(200, json), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, // Use ProtoV6 Provider
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlWithRetry(RequestBody),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if callCount != 2 {
							return fmt.Errorf("expected http request to be made 2 times. It was made %v times", callCount)
						}

						// Ensure the test has run for longer than the retry interval
						duration := time.Since(firstCall)
						if duration < 1*time.Second {
							return fmt.Errorf("expected test to have taken longer than the retry interval of 1s, test duration: %s", duration)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccresourceCurlWithRetry(body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]

  request_body = <<EOF
%s
EOF

  retry_interval = 1
  max_retry 	 	= 1
  method         = "POST"

  skip_destroy  = false
  destroy_url    = "https://example.com/destroy"
  destroy_method = "DELETE"

  destroy_response_codes = [
	"200", 
	"204"
  ]

  destroy_retry_interval 	= 1
  destroy_max_retry 	 	= 1

}
`, body)

}

func TestAccresourceCurlSkipRead(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("TF_LOG", "DEBUG")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/read",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testMockEndpointCount("GET https://example.com/read", 0),
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlSkipRead(rName, RequestBody),
			},
		},
	})

}

func testAccresourceCurlSkipRead(name string, body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "test" {
  name           = "%s"
  url            = "https://example.com/create"
  response_codes = ["200"]

  request_body = <<EOF
%s
EOF

  retry_interval = 1
  max_retry 	 	= 1
  method         = "POST"

  skip_destroy  = true
  skip_read  	= true
  read_url    	= "https://example.com/read"
  read_method 	= "GET"

  read_response_codes = [
	"200", 
	"204"
  ]

}
`, name, body)

}

func TestAccresourceCurlSkipReadNoReadFields(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlSkipReadNoReadFields(rName, RequestBody),
			},
		},
	})
}

func testAccresourceCurlSkipReadNoReadFields(name string, body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "test" {
    name           = "%s"
    url            = "https://example.com/create"
    response_codes = ["200"]

    request_body = <<EOF
%s
EOF

    retry_interval = 1
    max_retry     = 1
    method        = "POST"

    skip_destroy  = true
    skip_read     = true
}
`, name, body)
}

func TestAccresourceCurlRead(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/read",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testMockEndpointCount("GET https://example.com/read", 1),
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlRead(rName, RequestBody),
			},
		},
	})

}

func testAccresourceCurlRead(name string, body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "read" {
  name           = "%s"
  url            = "https://example.com/create"
  response_codes = ["200"]

  request_body = <<EOF
%s
EOF

  retry_interval = 1
  max_retry 	 	= 1
  method         = "POST"

  skip_destroy  = true
  skip_read  	= false
  read_url    	= "https://example.com/read"
  read_method 	= "GET"

  read_response_codes = [
	"200", 
	"204"
  ]

}
`, name, body)

}

func testAccresourceCurlTls(name, url, caCertFile, certFile, keyFile, readUrl, readCaCertFile, readCertFile, readKeyFile, destroyUrl, destroyCaCertFile, destroyCertFile, destroyKeyFile string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "tls_test" {
  name           = "%s"
  url            = "%s"
  response_codes = ["200"]


  retry_interval = 1
  max_retry 	 	= 1
  method         = "POST"
  ca_cert_file     = "%s"
  cert_file = "%s"
  key_file  = "%s"

  skip_read  	= false
  read_url    	= "%s"
  read_method 	= "GET"
  read_ca_cert_file     = "%s"
  read_cert_file = "%s"
  read_key_file  = "%s"

  read_response_codes = [
	"200", 
	"204"
  ]

  skip_destroy  = false
  destroy_url    	= "%s"
  destroy_method 	= "GET"
  destroy_ca_cert_file     = "%s"
  destroy_cert_file = "%s"
  destroy_key_file  = "%s"

  destroy_response_codes = [
	"200", 
	"204"
  ]


}
`, name, url, caCertFile, certFile, keyFile, readUrl, readCaCertFile, readCertFile, readKeyFile, destroyUrl, destroyCaCertFile, destroyCertFile, destroyKeyFile)

}

func TestAccCurlResourceWithTLS(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	server, certFile, keyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server for Create operation: %v. Cert file: %s", err, certFile)
	}
	fmt.Printf("Open server URL: %s\n", server.URL)

	readServer, readCertFile, readKeyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server for Read operation: %v. Cert file: %s", err, certFile)
	}

	destroyServer, destroyCertFile, destroyKeyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server for Destroy operation: %v. Cert file: %s", err, certFile)
	}

	//fmt.Printf("CertFile: %s, KeyFile: %s\n", certFile, keyFile)
	defer server.Close()
	defer readServer.Close()
	defer destroyServer.Close()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(certFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(keyFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(readCertFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(readKeyFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(destroyCertFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(destroyKeyFile)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlTls("tls_test", server.URL, certFile, certFile, keyFile, readServer.URL, readCertFile, readCertFile, readKeyFile, destroyServer.URL, destroyCertFile, destroyCertFile, destroyKeyFile),
				Check:  resource.TestCheckResourceAttr("terracurl_request.tls_test", "response", `{"message": "TLS test successful"}`),
			},
		},
	})
}

func testAccresourceCurlTlsSkipVerify(name, url, certFile, keyFile, readUrl, readCertFile, readKeyFile, destroyUrl, destroyCertFile, destroyKeyFile string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "tls_test" {
  name           = "%s"
  url            = "%s"
  response_codes = ["200"]


  retry_interval = 1
  max_retry 	 	= 1
  method         = "POST"
  cert_file = "%s"
  key_file  = "%s"
  skip_tls_verify = true

  skip_read  	= false
  read_url    	= "%s"
  read_method 	= "GET"
  read_cert_file = "%s"
  read_key_file  = "%s"
  read_skip_tls_verify = true

  read_response_codes = [
	"200", 
	"204"
  ]

  skip_destroy  = false
  destroy_url    	= "%s"
  destroy_method 	= "GET"
  destroy_cert_file = "%s"
  destroy_key_file  = "%s"
  destroy_skip_tls_verify = true

  destroy_response_codes = [
	"200", 
	"204"
  ]


}
`, name, url, certFile, keyFile, readUrl, readCertFile, readKeyFile, destroyUrl, destroyCertFile, destroyKeyFile)

}

func TestAccCurlResourceWithTLSSkipVerify(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	server, certFile, keyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server for Create operation: %v. Cert file: %s", err, certFile)
	}
	fmt.Printf("Open server URL: %s\n", server.URL)

	readServer, readCertFile, readKeyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server for Read operation: %v. Cert file: %s", err, certFile)
	}

	destroyServer, destroyCertFile, destroyKeyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server for Destroy operation: %v. Cert file: %s", err, certFile)
	}

	defer server.Close()
	defer readServer.Close()
	defer destroyServer.Close()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(certFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(keyFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(readCertFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(readKeyFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(destroyCertFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			fmt.Println(err)
		}
	}(destroyKeyFile)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlTlsSkipVerify("tls_test", server.URL, certFile, keyFile, readServer.URL, readCertFile, readKeyFile, destroyServer.URL, destroyCertFile, destroyKeyFile),
				Check:  resource.TestCheckResourceAttr("terracurl_request.tls_test", "response", `{"message": "TLS test successful"}`),
			},
		},
	})
}

func testMockEndpointCount(endpoint string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usage := httpmock.GetCallCountInfo()
		if usage[endpoint] != expected {
			return fmt.Errorf("endpoint called %d times, expected %d", usage[endpoint], expected)
		}
		return nil
	}
}

func testMockEndpointRegister(endpoint string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usage := httpmock.GetCallCountInfo()
		if usage[endpoint] < 1 {
			return fmt.Errorf("endpoint not called")
		}
		return nil
	}
}

func TestCurlResource_StateUpgrade(t *testing.T) {
	ctx := context.Background()
	r := &CurlResource{}

	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("No upgrader found for version 0")
	}

	// Define the complete schema
	schemaVar := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                     schema.StringAttribute{Computed: true},
			"name":                   schema.StringAttribute{Optional: true},
			"url":                    schema.StringAttribute{Required: true},
			"method":                 schema.StringAttribute{Optional: true},
			"headers":                schema.MapAttribute{ElementType: types.StringType, Optional: true},
			"parameters":             schema.MapAttribute{ElementType: types.StringType, Optional: true},
			"request_body":           schema.StringAttribute{Optional: true},
			"cert_file":              schema.StringAttribute{Optional: true},
			"key_file":               schema.StringAttribute{Optional: true},
			"ca_cert_file":           schema.StringAttribute{Optional: true},
			"ca_cert_directory":      schema.StringAttribute{Optional: true},
			"skip_tls_verify":        schema.BoolAttribute{Optional: true},
			"timeout":                schema.Int64Attribute{Optional: true},
			"response_codes":         schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"status_code":            schema.StringAttribute{Computed: true},
			"response":               schema.StringAttribute{Computed: true},
			"request_url_string":     schema.StringAttribute{Computed: true},
			"max_retry":              schema.Int64Attribute{Optional: true},
			"retry_interval":         schema.Int64Attribute{Optional: true},
			"ignore_response_fields": schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"drift_marker":           schema.StringAttribute{Optional: true},

			// Read-related fields
			"skip_read":              schema.BoolAttribute{Optional: true},
			"read_url":               schema.StringAttribute{Optional: true},
			"read_method":            schema.StringAttribute{Optional: true},
			"read_headers":           schema.MapAttribute{ElementType: types.StringType, Optional: true},
			"read_parameters":        schema.MapAttribute{ElementType: types.StringType, Optional: true},
			"read_request_body":      schema.StringAttribute{Optional: true},
			"read_cert_file":         schema.StringAttribute{Optional: true},
			"read_key_file":          schema.StringAttribute{Optional: true},
			"read_ca_cert_file":      schema.StringAttribute{Optional: true},
			"read_ca_cert_directory": schema.StringAttribute{Optional: true},
			"read_skip_tls_verify":   schema.BoolAttribute{Optional: true},
			"read_response_codes":    schema.ListAttribute{ElementType: types.StringType, Optional: true},

			// Destroy-related fields
			"skip_destroy":               schema.BoolAttribute{Optional: true},
			"destroy_url":                schema.StringAttribute{Optional: true},
			"destroy_method":             schema.StringAttribute{Optional: true},
			"destroy_headers":            schema.MapAttribute{ElementType: types.StringType, Optional: true},
			"destroy_parameters":         schema.MapAttribute{ElementType: types.StringType, Optional: true},
			"destroy_request_body":       schema.StringAttribute{Optional: true},
			"destroy_cert_file":          schema.StringAttribute{Optional: true},
			"destroy_key_file":           schema.StringAttribute{Optional: true},
			"destroy_ca_cert_file":       schema.StringAttribute{Optional: true},
			"destroy_ca_cert_directory":  schema.StringAttribute{Optional: true},
			"destroy_skip_tls_verify":    schema.BoolAttribute{Optional: true},
			"destroy_response_codes":     schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"destroy_timeout":            schema.Int64Attribute{Optional: true},
			"destroy_max_retry":          schema.Int64Attribute{Optional: true},
			"destroy_retry_interval":     schema.Int64Attribute{Optional: true},
			"destroy_request_url_string": schema.StringAttribute{Computed: true},
		},
	}

	// Create initial state
	oldState := &CurlResourceModel{
		Id:             types.StringValue("test-resource"),
		Name:           types.StringValue("test"),
		Url:            types.StringValue("https://api.example.com"),
		Method:         types.StringValue("POST"),
		ReadUrl:        types.StringValue("https://api.example.com/read"),
		Headers:        types.MapValueMust(types.StringType, map[string]attr.Value{}),
		ReadHeaders:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
		ReadParameters: types.MapValueMust(types.StringType, map[string]attr.Value{}),
		ReadResponseCodes: types.ListValueMust(
			types.StringType,
			[]attr.Value{
				types.StringValue("200"),
			},
		),
		ResponseCodes:        types.ListValueMust(types.StringType, []attr.Value{}),
		IgnoreResponseFields: types.ListValueMust(types.StringType, []attr.Value{}),
		Parameters:           types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DestroyHeaders:       types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DestroyParameters:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DestroyResponseCodes: types.ListValueMust(types.StringType, []attr.Value{}),
	}

	state := tfsdk.State{
		Schema: schemaVar,
	}
	diags := state.Set(ctx, oldState)
	if diags.HasError() {
		t.Fatalf("error setting initial state: %v", diags)
	}

	req := resource2.UpgradeStateRequest{
		State: &state,
	}

	resp := &resource2.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: schemaVar,
		},
	}

	upgrader.StateUpgrader(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade failed: %v", resp.Diagnostics)
	}

	var upgradedState CurlResourceModel
	diags = resp.State.Get(ctx, &upgradedState)
	if diags.HasError() {
		t.Fatalf("error getting upgraded state: %v", diags)
	}

	// Verify the results
	if !upgradedState.ReadUrl.IsNull() {
		t.Error("ReadUrl should be null after upgrade")
	}
	if len(upgradedState.ReadResponseCodes.Elements()) != 0 {
		t.Errorf("ReadResponseCodes should be an empty list after upgrade. Got: %v", upgradedState.ReadResponseCodes.Elements())
	}

	if upgradedState.Id.ValueString() != "test-resource" {
		t.Error("Id was not preserved")
	}

	if upgradedState.Url.ValueString() != "https://api.example.com" {
		t.Error("Url was not preserved")
	}
}
