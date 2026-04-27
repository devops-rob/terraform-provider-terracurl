package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	resource2 "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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

  request_parameters = {
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

  destroy_request_parameters = {
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

  request_parameters = {
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

  destroy_request_parameters = {
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

  request_parameters = {
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

  destroy_request_parameters = {
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

func TestAccresourceCurlResponseSensitive(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	secretBody := `{"token": "super-secret-token", "key": "key-1234"}`
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/keys",
		httpmock.NewStringResponder(200, secretBody),
	)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlResponseSensitive(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("terracurl_request.sensitive", "response_sensitive", "true"),
					resource.TestCheckResourceAttr("terracurl_request.sensitive", "response", ""),
					resource.TestCheckResourceAttr("terracurl_request.sensitive", "sensitive_response", secretBody),
					resource.TestCheckResourceAttr("terracurl_request.sensitive", "status_code", "200"),
				),
			},
		},
	})
}

func testAccresourceCurlResponseSensitive(name string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "sensitive" {
  name               = "%s"
  url                = "https://example.com/keys"
  method             = "POST"
  response_codes     = ["200"]
  response_sensitive = true
}`, name)
}

func TestAccresourceCurlResponseSensitiveDefault(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	body := `{"message": "ok"}`
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/default",
		httpmock.NewStringResponder(200, body),
	)

	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlResponseSensitiveDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("terracurl_request.default", "response_sensitive", "false"),
					resource.TestCheckResourceAttr("terracurl_request.default", "response", body),
					resource.TestCheckResourceAttr("terracurl_request.default", "sensitive_response", ""),
				),
			},
		},
	})
}

func testAccresourceCurlResponseSensitiveDefault(name string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "default" {
  name           = "%s"
  url            = "https://example.com/default"
  method         = "GET"
  response_codes = ["200"]
}`, name)
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
			"request_parameters":     schema.MapAttribute{ElementType: types.StringType, Optional: true},
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
			"sensitive_response":     schema.StringAttribute{Computed: true, Sensitive: true},
			"response_sensitive":     schema.BoolAttribute{Optional: true, Computed: true},
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
			"destroy_request_parameters": schema.MapAttribute{ElementType: types.StringType, Optional: true},
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
		ResponseCodes:            types.ListValueMust(types.StringType, []attr.Value{}),
		IgnoreResponseFields:     types.ListValueMust(types.StringType, []attr.Value{}),
		RequestParameters:        types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DestroyHeaders:           types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DestroyRequestParameters: types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DestroyResponseCodes:     types.ListValueMust(types.StringType, []attr.Value{}),
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

	if !upgradedState.ReadResponseCodes.IsNull() {
		t.Error("ReadResponseCodes should be null after upgrade")
	}

	if upgradedState.Id.ValueString() != "test-resource" {
		t.Error("Id was not preserved")
	}

	if upgradedState.Url.ValueString() != "https://api.example.com" {
		t.Error("Url was not preserved")
	}
}

// TestCurlResource_StateUpgrade_WithDestroyParameters tests the migration from destroy_parameters to destroy_request_parameters
func TestCurlResource_StateUpgrade_WithDestroyParameters(t *testing.T) {
	ctx := context.Background()
	r := &CurlResource{}

	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("No upgrader found for version 0")
	}

	// Simulate raw state JSON that contains the old "destroy_parameters" attribute
	rawStateJSON := `{
		"id": "test-resource-123",
		"name": "example_role",
		"url": "https://api.example.com/create",
		"method": "POST",
		"request_body": "{\"role\": \"test_role\"}",
		"headers": {"Authorization": "Bearer token", "Content-Type": "application/json"},
		"request_parameters": {"database": "test_db"},
		"response_codes": ["200", "201"],
		"skip_destroy": false,
	"destroy_url": "https://api.example.com/delete",
		"destroy_method": "DELETE",
		"destroy_parameters": {"role_id": "12345", "force": "true"},
		"destroy_headers": {"Authorization": "Bearer token"},
		"destroy_response_codes": ["200", "204"],
		"retry_interval": 10,
		"max_retry": 3,
		"timeout": 30,
		"cert_file": "/path/to/cert.pem",
		"key_file": "/path/to/key.pem",
		"skip_tls_verify": true
	}`

	// Create a minimal schema for the empty state to avoid nil pointer issues
	// Use the actual resource schema for consistency
	rUpgrade := &CurlResource{}
	schemaResp := &resource2.SchemaResponse{}
	rUpgrade.Schema(ctx, resource2.SchemaRequest{}, schemaResp)
	schemaVar := schemaResp.Schema

	// Create a minimal schema for the empty state to avoid nil pointer issues
	emptySchema1 := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
		},
	}

	req := resource2.UpgradeStateRequest{
		State: &tfsdk.State{Schema: emptySchema1}, // Provide schema to avoid nil pointer
		RawState: &tfprotov6.RawState{
			JSON: []byte(rawStateJSON),
		},
	}

	resp := &resource2.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: schemaVar,
		},
	}

	// Execute the upgrade
	upgrader.StateUpgrader(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade failed: %v", resp.Diagnostics)
	}

	// Get the upgraded state
	var upgradedState CurlResourceModel
	diags := resp.State.Get(ctx, &upgradedState)
	if diags.HasError() {
		t.Fatalf("error getting upgraded state: %v", diags)
	}

	// Verify that data was preserved and migrated correctly
	if upgradedState.Id.ValueString() != "test-resource-123" {
		t.Errorf("Expected id 'test-resource-123', got '%s'", upgradedState.Id.ValueString())
	}

	if upgradedState.Name.ValueString() != "example_role" {
		t.Errorf("Expected name 'example_role', got '%s'", upgradedState.Name.ValueString())
	}

	if upgradedState.Url.ValueString() != "https://api.example.com/create" {
		t.Errorf("Expected url 'https://api.example.com/create', got '%s'", upgradedState.Url.ValueString())
	}

	if upgradedState.Method.ValueString() != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", upgradedState.Method.ValueString())
	}

	if upgradedState.RequestBody.ValueString() != "{\"role\": \"test_role\"}" {
		t.Errorf("Expected request_body to be preserved, got '%s'", upgradedState.RequestBody.ValueString())
	}

	if upgradedState.DestroyUrl.ValueString() != "https://api.example.com/delete" {
		t.Errorf("Expected destroy_url to be preserved, got '%s'", upgradedState.DestroyUrl.ValueString())
	}

	if upgradedState.DestroyMethod.ValueString() != "DELETE" {
		t.Errorf("Expected destroy_method to be preserved, got '%s'", upgradedState.DestroyMethod.ValueString())
	}

	// Verify the key migration: destroy_parameters -> destroy_request_parameters
	if upgradedState.DestroyRequestParameters.IsNull() {
		t.Error("Expected destroy_request_parameters to be populated from destroy_parameters")
	} else {
		destroyParams := upgradedState.DestroyRequestParameters.Elements()
		if roleId, ok := destroyParams["role_id"]; !ok || roleId.(types.String).ValueString() != "12345" {
			t.Error("Expected destroy_request_parameters to contain role_id='12345'")
		}
		if force, ok := destroyParams["force"]; !ok || force.(types.String).ValueString() != "true" {
			t.Error("Expected destroy_request_parameters to contain force='true'")
		}
	}

	// Verify other complex types were preserved
	if upgradedState.Headers.IsNull() {
		t.Error("Expected headers to be preserved")
	} else {
		headers := upgradedState.Headers.Elements()
		if auth, ok := headers["Authorization"]; !ok || auth.(types.String).ValueString() != "Bearer token" {
			t.Error("Expected headers to contain Authorization='Bearer token'")
		}
	}

	if upgradedState.RequestParameters.IsNull() {
		t.Error("Expected request_parameters to be preserved")
	} else {
		reqParams := upgradedState.RequestParameters.Elements()
		if db, ok := reqParams["database"]; !ok || db.(types.String).ValueString() != "test_db" {
			t.Error("Expected request_parameters to contain database='test_db'")
		}
	}

	// Verify response_codes list was preserved
	if upgradedState.ResponseCodes.IsNull() {
		t.Error("Expected response_codes to be preserved")
	} else {
		responseCodes := upgradedState.ResponseCodes.Elements()
		if len(responseCodes) != 2 {
			t.Errorf("Expected 2 response codes, got %d", len(responseCodes))
		}
	}

	// Verify integer fields were preserved
	if upgradedState.RetryInterval.ValueInt64() != 10 {
		t.Errorf("Expected retry_interval 10, got %d", upgradedState.RetryInterval.ValueInt64())
	}

	if upgradedState.MaxRetry.ValueInt64() != 3 {
		t.Errorf("Expected max_retry 3, got %d", upgradedState.MaxRetry.ValueInt64())
	}

	if upgradedState.Timeout.ValueInt64() != 30 {
		t.Errorf("Expected timeout 30, got %d", upgradedState.Timeout.ValueInt64())
	}

	// Verify boolean fields were preserved
	if !upgradedState.SkipTlsVerify.ValueBool() {
		t.Error("Expected skip_tls_verify to be true")
	}

	if upgradedState.SkipDestroy.ValueBool() {
		t.Error("Expected skip_destroy to be false")
	}

	// Verify string fields were preserved
	if upgradedState.CertFile.ValueString() != "/path/to/cert.pem" {
		t.Errorf("Expected cert_file to be preserved, got '%s'", upgradedState.CertFile.ValueString())
	}

	if upgradedState.KeyFile.ValueString() != "/path/to/key.pem" {
		t.Errorf("Expected key_file to be preserved, got '%s'", upgradedState.KeyFile.ValueString())
	}

	// Verify v1 upgrade changes
	if !upgradedState.SkipRead.ValueBool() {
		t.Error("Expected skip_read to be set to true in v1 upgrade")
	}

	if !upgradedState.ReadUrl.IsNull() {
		t.Error("Expected read_url to be null in v1 upgrade")
	}

	if !upgradedState.ReadResponseCodes.IsNull() {
		t.Error("Expected read_response_codes to be null in v1 upgrade")
	}
}

// TestCurlResource_Read_ResponseSensitiveToggle tests that drift detection correctly
// handles the case where response_sensitive changes between operations
func TestCurlResource_Read_ResponseSensitiveToggle(t *testing.T) {
	t.Setenv("TF_ACC", "true")
	t.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Initial response data
	initialResponse := `{"key": "value1"}`
	// Changed response data (for drift detection)
	changedResponse := `{"key": "value2"}`

	callCount := 0
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/read",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return httpmock.NewStringResponse(200, initialResponse), nil
			}
			return httpmock.NewStringResponse(200, changedResponse), nil
		},
	)

	ctx := context.Background()
	r := &CurlResource{}
	r.client = &http.Client{Timeout: 30 * time.Second}

	// Get the schema
	schemaResp := &resource2.SchemaResponse{}
	r.Schema(ctx, resource2.SchemaRequest{}, schemaResp)

	// Test scenario 1: response_sensitive = false initially
	// Create initial state with response_sensitive = false
	initialState := CurlResourceModel{
		Id:                       types.StringValue("test"),
		Name:                     types.StringValue("test"),
		Url:                      types.StringValue("https://example.com/create"),
		Method:                   types.StringValue("POST"),
		SkipRead:                 types.BoolValue(false),
		ReadUrl:                  types.StringValue("https://example.com/read"),
		ReadMethod:               types.StringValue("GET"),
		ReadResponseCodes:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ResponseCodes:            types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ResponseSensitive:        types.BoolValue(false),
		Response:                 types.StringValue(initialResponse),
		SensitiveResponse:        types.StringValue(""),
		DriftMarker:              types.StringValue("initial"),
		IgnoreResponseFields:     types.ListNull(types.StringType),
		DestroyResponseCodes:     types.ListNull(types.StringType),
		SkipDestroy:              types.BoolValue(true),
		Headers:                  types.MapNull(types.StringType),
		RequestParameters:        types.MapNull(types.StringType),
		ReadHeaders:              types.MapNull(types.StringType),
		ReadParameters:           types.MapNull(types.StringType),
		DestroyHeaders:           types.MapNull(types.StringType),
		DestroyRequestParameters: types.MapNull(types.StringType),
	}

	state1 := tfsdk.State{Schema: schemaResp.Schema}
	diags := state1.Set(ctx, &initialState)
	if diags.HasError() {
		t.Fatalf("Failed to set initial state: %v", diags)
	}

	// Call Read() - should detect no drift (same response)
	readReq1 := resource2.ReadRequest{State: state1}
	readResp1 := &resource2.ReadResponse{State: state1}
	r.Read(ctx, readReq1, readResp1)

	if readResp1.Diagnostics.HasError() {
		t.Fatalf("Read failed: %v", readResp1.Diagnostics)
	}

	var stateAfterRead1 CurlResourceModel
	diags = readResp1.State.Get(ctx, &stateAfterRead1)
	if diags.HasError() {
		t.Fatalf("Failed to get state after first read: %v", diags)
	}

	// Verify no drift was detected
	if stateAfterRead1.DriftMarker.ValueString() != "initial" {
		t.Errorf("Expected no drift after first read, but drift_marker changed to: %s", stateAfterRead1.DriftMarker.ValueString())
	}

	// Test scenario 2: User toggles response_sensitive to true
	// The prior state still has response_sensitive = false with data in `response`
	// Simulate what happens when the user changes the config
	stateWithToggle := CurlResourceModel{
		Id:                       types.StringValue("test"),
		Name:                     types.StringValue("test"),
		Url:                      types.StringValue("https://example.com/create"),
		Method:                   types.StringValue("POST"),
		SkipRead:                 types.BoolValue(false),
		ReadUrl:                  types.StringValue("https://example.com/read"),
		ReadMethod:               types.StringValue("GET"),
		ReadResponseCodes:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ResponseCodes:            types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ResponseSensitive:        types.BoolValue(false),
		Response:                 types.StringValue(initialResponse),
		SensitiveResponse:        types.StringValue(""),
		DriftMarker:              types.StringValue("initial"),
		IgnoreResponseFields:     types.ListNull(types.StringType),
		DestroyResponseCodes:     types.ListNull(types.StringType),
		SkipDestroy:              types.BoolValue(true),
		Headers:                  types.MapNull(types.StringType),
		RequestParameters:        types.MapNull(types.StringType),
		ReadHeaders:              types.MapNull(types.StringType),
		ReadParameters:           types.MapNull(types.StringType),
		DestroyHeaders:           types.MapNull(types.StringType),
		DestroyRequestParameters: types.MapNull(types.StringType),
	}

	state2 := tfsdk.State{Schema: schemaResp.Schema}
	diags = state2.Set(ctx, &stateWithToggle)
	if diags.HasError() {
		t.Fatalf("Failed to set state with toggle: %v", diags)
	}

	// Call Read() again - this time the API returns different data
	// The Read function should:
	// 1. Look at the OLD response_sensitive value (false) from state
	// 2. Get the prior response from the `response` attribute
	// 3. Compare it with the new response
	// 4. Detect drift because data changed
	readReq2 := resource2.ReadRequest{State: state2}
	readResp2 := &resource2.ReadResponse{State: state2}
	r.Read(ctx, readReq2, readResp2)

	if readResp2.Diagnostics.HasError() {
		t.Fatalf("Read failed after toggle: %v", readResp2.Diagnostics)
	}

	var stateAfterRead2 CurlResourceModel
	diags = readResp2.State.Get(ctx, &stateAfterRead2)
	if diags.HasError() {
		t.Fatalf("Failed to get state after second read: %v", diags)
	}

	// Verify drift WAS detected (because response changed from value1 to value2)
	if stateAfterRead2.DriftMarker.ValueString() == "initial" {
		t.Error("Expected drift to be detected after response changed, but drift_marker remained 'initial'")
	}

	// Test scenario 3: Opposite direction - toggle from true to false
	// Prior state has response_sensitive = true with data in sensitive_response
	stateWithReverseToggle := CurlResourceModel{
		Id:                       types.StringValue("test"),
		Name:                     types.StringValue("test"),
		Url:                      types.StringValue("https://example.com/create"),
		Method:                   types.StringValue("POST"),
		SkipRead:                 types.BoolValue(false),
		ReadUrl:                  types.StringValue("https://example.com/read"),
		ReadMethod:               types.StringValue("GET"),
		ReadResponseCodes:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ResponseCodes:            types.ListValueMust(types.StringType, []attr.Value{types.StringValue("200")}),
		ResponseSensitive:        types.BoolValue(true),
		Response:                 types.StringValue(""),
		SensitiveResponse:        types.StringValue(changedResponse),
		DriftMarker:              types.StringValue("initial"),
		IgnoreResponseFields:     types.ListNull(types.StringType),
		DestroyResponseCodes:     types.ListNull(types.StringType),
		SkipDestroy:              types.BoolValue(true),
		Headers:                  types.MapNull(types.StringType),
		RequestParameters:        types.MapNull(types.StringType),
		ReadHeaders:              types.MapNull(types.StringType),
		ReadParameters:           types.MapNull(types.StringType),
		DestroyHeaders:           types.MapNull(types.StringType),
		DestroyRequestParameters: types.MapNull(types.StringType),
	}

	state3 := tfsdk.State{Schema: schemaResp.Schema}
	diags = state3.Set(ctx, &stateWithReverseToggle)
	if diags.HasError() {
		t.Fatalf("Failed to set state with reverse toggle: %v", diags)
	}

	// Reset call count to return initial response again
	callCount = 0

	readReq3 := resource2.ReadRequest{State: state3}
	readResp3 := &resource2.ReadResponse{State: state3}
	r.Read(ctx, readReq3, readResp3)

	if readResp3.Diagnostics.HasError() {
		t.Fatalf("Read failed after reverse toggle: %v", readResp3.Diagnostics)
	}

	var stateAfterRead3 CurlResourceModel
	diags = readResp3.State.Get(ctx, &stateAfterRead3)
	if diags.HasError() {
		t.Fatalf("Failed to get state after third read: %v", diags)
	}

	// Verify drift WAS detected (because response changed from value2 back to value1)
	if stateAfterRead3.DriftMarker.ValueString() == "initial" {
		t.Error("Expected drift to be detected after reverse toggle, but drift_marker remained 'initial'")
	}
}

// TestCurlResource_StateUpgrade_EmptyDestroyParameters tests handling of null destroy_parameters
func TestCurlResource_StateUpgrade_EmptyDestroyParameters(t *testing.T) {
	ctx := context.Background()
	r := &CurlResource{}

	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("No upgrader found for version 0")
	}

	// Simulate raw state JSON with null destroy_parameters
	rawStateJSON := `{
		"id": "test-null-params",
		"name": "test_resource",
		"url": "https://example.com",
		"method": "GET",
		"destroy_parameters": null
	}`

	rawState := &tfprotov6.RawState{
		JSON: []byte(rawStateJSON),
	}

	// Use the actual resource schema for consistency  
	rUpgrade2 := &CurlResource{}
	schemaResp2 := &resource2.SchemaResponse{}
	rUpgrade2.Schema(ctx, resource2.SchemaRequest{}, schemaResp2)
	schemaVar2 := schemaResp2.Schema

	emptySchema2 := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
		},
	}

	req := resource2.UpgradeStateRequest{
		State: &tfsdk.State{Schema: emptySchema2},
		RawState: rawState,
	}

	resp := &resource2.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: schemaVar2,
		},
	}

	upgrader.StateUpgrader(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("upgrade failed: %v", resp.Diagnostics)
	}

	var upgradedState CurlResourceModel
	diags := resp.State.Get(ctx, &upgradedState)
	if diags.HasError() {
		t.Fatalf("error getting upgraded state: %v", diags)
	}

	// Verify that null destroy_parameters results in null destroy_request_parameters
	if !upgradedState.DestroyRequestParameters.IsNull() {
		t.Error("Expected destroy_request_parameters to be null when destroy_parameters was null")
	}

	// Verify basic fields were still extracted
	if upgradedState.Id.ValueString() != "test-null-params" {
		t.Errorf("Expected id to be preserved, got '%s'", upgradedState.Id.ValueString())
	}
}
