package provider

import (
	"fmt"
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

func testMockEndpointCount(endpoint string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		usage := httpmock.GetCallCountInfo()
		if usage[endpoint] != expected {
			return fmt.Errorf("endpoint called %d times, expected %d", usage[endpoint], expected)
		}
		return nil
	}
}
