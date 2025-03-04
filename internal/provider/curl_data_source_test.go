package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jarcoal/httpmock"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var RequestBody = `{"example": "request_body"}`

func TestAccCurlDataSourceBasic(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()

	responseBody := `{"response": "test_response"}`
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/data",
		httpmock.NewStringResponder(200, responseBody),
	)
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, // Use ProtoV6 Provider
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCurlRequestBasic(rName, RequestBody),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "method", "POST"),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "name", rName),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "url", "https://example.com/data"),
					resource.TestCheckResourceAttrSet("data.terracurl_request.basic_test", "response"),
					resource.TestCheckTypeSetElemAttr("data.terracurl_request.basic_test", "headers.*", "Bearer token"),
					resource.TestCheckTypeSetElemAttr("data.terracurl_request.basic_test", "headers.*", "application/json"),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "request_url_string", "https://example.com/data?id=12345&name=devopsrob"),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "request_body", RequestBody+"\n"),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "response_codes.#", "1"),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "response_codes.0", "200"),

					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "response", responseBody),
					resource.TestCheckResourceAttr("data.terracurl_request.basic_test", "status_code", "200"),
				),
			},
		},
	})
}
func testAccDataSourceCurlRequestBasic(name string, requestBody string) string {
	return fmt.Sprintf(`
data "terracurl_request" "basic_test" {
  method         = "POST"
  name           = "%s"
  response_codes = ["200"]
  url            = "https://example.com/data"
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
  
}
`, name, requestBody)
}

func TestAccCurlDataSourceRetries(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	rName := "devopsrob"
	json := `{"name": "` + rName + `"}`

	var firstCall time.Time
	var callCount int

	// Register the responder to simulate a failure followed by a success
	httpmock.RegisterResponder("POST", "https://example.com/data",
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
				Config: testAccdataSourceCurlBodyWithRetry(RequestBody),
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

func testAccdataSourceCurlBodyWithRetry(body string) string {
	return fmt.Sprintf(`
data "terracurl_request" "test" {
 name           = "leader"
 url            = "https://example.com/data"
 response_codes = ["200"]

 request_body = <<EOF
%s
EOF

 retry_interval = 1
 max_retry 	 	= 1
 method         = "POST"
}
`, body)

}

func TestAccdataSourceCurlRequestTimeout(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()

	rName := "devopsrob"
	json := `{"name": "` + rName + `"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			//time.Sleep(2 * time.Second) // Delay the response by 6 seconds
			return httpmock.NewStringResponse(200, `{"name": "devopsrob"}`), nil
		},
	)

	start := time.Now() // Record the start time
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, // Use ProtoV6 Provider
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config:   testAccdataSourceCurlWithTimeout(json),
				Check: resource.ComposeTestCheckFunc(
					testCheckDurationWithTimeout(),
				),
			},
		},
	})
	duration := time.Since(start) // Calculate the duration of the API call
	t.Logf("API call took %v", duration)
}

func testCheckDurationWithTimeout() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState := state.RootModule().Resources["data.terracurl_request.test"]
		durationStr := resourceState.Primary.Attributes["duration_milliseconds"]
		if durationStr == "" {
			return fmt.Errorf("duration_milliseconds attribute not set")
		}
		duration, err := time.ParseDuration(durationStr + "ms")
		if err != nil {
			return fmt.Errorf("failed to parse duration: %v", err)
		}
		if duration < 1*time.Second || duration > 30*time.Second {
			return fmt.Errorf("expected duration between 1 and 30 seconds, got %v", duration)
		}
		// Check for timeout error message
		if !strings.Contains(resourceState.Primary.Attributes["response"], "context canceled, not retrying operation") &&
			!strings.Contains(resourceState.Primary.Attributes["response"], "request failed, retries exceeded") {
			return fmt.Errorf("expected error message not found in response")
		}
		return nil
	}
}

func testAccdataSourceCurlWithTimeout(body string) string {
	return fmt.Sprintf(`
data "terracurl_request" "test" {
 name           = "leader"
 url            = "https://example.com/create"
 response_codes = ["200"]

 request_body = <<EOF
%s
EOF

 retry_interval = 1
 max_retry 	 	= 1
 method         = "POST"
 timeout = 1
}
`, body)

}

func TestAccCurlDataSourceTLS(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()
	server, certFile, keyFile, err := createTLSServer()
	if err != nil {
		t.Fatalf("failed to create TLS test server: %v. Cert file: %s", err, certFile)
	}
	fmt.Printf("CertFile: %s, KeyFile: %s\n", certFile, keyFile)
	defer server.Close()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(certFile)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(keyFile)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCurlTLS(server.URL, certFile, keyFile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.terracurl_request.tls_test", "method", "GET"),
					resource.TestCheckResourceAttr("data.terracurl_request.tls_test", "response", `{"message": "TLS test successful"}`),
				),
			},
		},
	})
}

// Terraform config for TLS test
func testAccDataSourceCurlTLS(url, certFile, keyFile string) string {
	return fmt.Sprintf(`
data "terracurl_request" "tls_test" {
  name 			   = "tls-test"
  method           = "GET"
  url              = "%s"
  response_codes   = ["200"]
  ca_cert_file     = "%s"
  cert_file = "%s"
  key_file  = "%s"
}
`, url, certFile, certFile, keyFile)
}
