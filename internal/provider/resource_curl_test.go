package provider

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestAccresourceCurl(t *testing.T) {
	err := os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
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

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurl,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						`{"name": "devopsrob"}`),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "url",
						"https://example.com"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "name",
						"leader"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "method",
						"GET"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "destroy_url",
						"https://example.com"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "destroy_method",
						"GET"),
				),
			},
		},
	})
}

const testAccresourceCurl = `
resource "terracurl_request" "example" {
  name           = "leader"
  url            = "https://example.com"
  method         = "GET"
  response_codes = ["200"]
  destroy_url    = "https://example.com"
  destroy_method = "GET"
  destroy_response_codes = ["200"]
}
`

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Do(r *http.Request) (*http.Response, error) {
	args := m.Mock.Called(r)

	if resp, ok := args.Get(0).(*http.Response); ok {
		return resp, args.Error(1)
	}

	return nil, args.Error(1)
}

func TestAccresourceRetriesOnFailure(t *testing.T) {
	err := os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
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

	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	// Track the first call time and call count
	var firstCall time.Time
	var callCount int

	// Simulate a failure on the first call and success on the second
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

	httpmock.RegisterResponder("POST", "https://example.com/destroy",
		httpmock.NewStringResponder(204, ""),
	)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories, // Ensure this is correctly set up for your provider
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlBodyWithRetry(json),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if callCount != 2 {
							return fmt.Errorf("expected HTTP request to be made 2 times; made %v times", callCount)
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
func TestAccresourceRetriesOnError(t *testing.T) {
	err := os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()
	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var callCount int

	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				// Return a failure response for the first call
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil
			}

			// Return a success response for subsequent calls
			return httpmock.NewStringResponse(200, `{"name": "devopsrob"}`), nil
		})

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/destroy",
		httpmock.NewStringResponder(204, ""))

	// ensure default client is replaced after test
	defer func() {
		Client = &http.Client{}
	}()

	callCount = 0
	resource.UnitTest(t, resource.TestCase{
		CheckDestroy:      testAccCheckRequestDestroy,
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlBodyWithRetry(json),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if callCount != 2 {
							return fmt.Errorf("expected http request to be made 2 times. It was made %v times", callCount)
						}
						return nil
					}),
			},
		},
	})
}

func TestAccresourceCurlBody(t *testing.T) {
	err := os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()
	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, json),
	)

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/destroy",
		httpmock.NewStringResponder(204, ""))
	resource.UnitTest(t, resource.TestCase{
		CheckDestroy:      testAccCheckRequestDestroy,
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlBody(json),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						json),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "url",
						"https://example.com/create"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "name",
						"leader"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						json),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "method",
						"POST"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "destroy_url",
						"https://example.com/destroy"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "destroy_method",
						"POST"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response_codes.#", "1"),
				),
			},
		},
	})
}

func testAccresourceCurlBody(body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "example" {
  name           = "leader"
  url            = "https://example.com/create"
  request_body = <<EOF
%s
EOF
  method         = "POST"
  response_codes = ["200"]
  destroy_url    = "https://example.com/destroy"
  destroy_method = "POST"
  destroy_response_codes = ["204"]
}

`, body)
}

func testAccresourceCurlBodyWithRetry(body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "example" {
  name           = "leader"
  url            = "https://example.com/create"
  request_body = <<EOF
%s
EOF
  retry_interval = 1
  max_retry = 1
  method         = "POST"
  response_codes = ["200"]
  destroy_url    = "https://example.com/destroy"
  destroy_method = "POST"
  destroy_response_codes = ["204"]
}

`, body)
}

func testAccCheckRequestDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "terracurl_request" {
			continue
		}
	}
	return nil
}

func testAccresourceNoDestroy(body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "example" {
  name           = "leader"
  url            = "https://example.com/create"
  # destroy_url    = "http://example.com"

  request_body = <<EOF
%s
EOF
  method         = "POST"
  response_codes = ["200","204"]
}


`, body)
}

func TestAccresourceCurlParameters(t *testing.T) {
	err := os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()

	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, json),
	)

	httpmock.RegisterResponder(
		"POST",
		"https://example.com/destroy",
		httpmock.NewStringResponder(204, ""))
	resource.UnitTest(t, resource.TestCase{
		CheckDestroy:      testAccCheckRequestDestroy,
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlParameters(json),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						json),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "url",
						"https://example.com/create"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "name",
						"leader"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						json),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "method",
						"POST"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "request_url_string",
						"https://example.com/create?Action=parameter"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "destroy_url",
						"https://example.com/destroy"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "destroy_method",
						"POST"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response_codes.#", "1"),
				),
			},
		},
	})
}

func testAccresourceCurlParameters(body string) string {
	return fmt.Sprintf(`
resource "terracurl_request" "example" {
  name           = "leader"
  url            = "https://example.com/create"
  request_parameters = {
	"Action" = "parameter"
  }
  request_body = <<EOF
%s
EOF
  method         = "POST"
  response_codes = ["200"]
  destroy_url    = "https://example.com/destroy"
  destroy_method = "POST"
  destroy_response_codes = ["204"]
}


`, body)
}

func TestAccresourceNoDestroy(t *testing.T) {
	err := os.Setenv("USE_DEFAULT_CLIENT_FOR_TESTS", "true")
	if err != nil {
		return
	}
	defer func() {
		err := os.Unsetenv("USE_DEFAULT_CLIENT_FOR_TESTS")
		if err != nil {

		}
	}()

	json := `{"name": "leader"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, json),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"http://example.com",
		httpmock.NewStringResponder(405, ""),
	)

	resource.UnitTest(t, resource.TestCase{
		CheckDestroy:      testAccCheckRequestDestroy,
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceNoDestroy(json),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						json),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "url",
						"https://example.com/create"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "name",
						"leader"),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "response",
						json),
					resource.TestCheckResourceAttr(
						"terracurl_request.example", "method",
						"POST"),
				),
			},
		},
	})
}
