package provider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
)

func TestAccresourceCurl(t *testing.T) {
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
	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	mc := &mockClient{}

	// First failed response
	resp1 := &http.Response{}
	resp1.StatusCode = http.StatusInternalServerError
	resp1.Body = ioutil.NopCloser(bytes.NewReader([]byte("boom")))

	// Second successful response
	resp2 := &http.Response{}
	resp2.StatusCode = http.StatusOK
	resp2.Body = ioutil.NopCloser(bytes.NewReader([]byte(json)))

	// Destroy response
	resp3 := &http.Response{}
	resp3.StatusCode = http.StatusNoContent
	resp3.Body = ioutil.NopCloser(bytes.NewReader([]byte(json)))

	c := mc.On("Do", mock.Anything).Once().Return(resp1, nil)
	mc.On("Do", mock.Anything).Once().Return(resp2, nil)
	mc.On("Do", mock.Anything).Once().Return(resp3, nil)

	var firstCall time.Time
	c.RunFn = func(args mock.Arguments) {
		firstCall = time.Now()
	}

	Client = mc

	// ensure default client is replaced after test
	defer func() {
		Client = &http.Client{}
	}()

	resource.UnitTest(t, resource.TestCase{
		CheckDestroy:      testAccCheckRequestDestroy,
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlBodyWithRetry(json),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if len(mc.Calls) != 2 {
							return fmt.Errorf("expected http request to be made 2 times")
						}

						// ensure the test has run for longer than the retry interval
						duration := time.Since(firstCall)
						if duration < 1*time.Second {
							return fmt.Errorf("expected test to have taken longer than the retry interval of 1s, test duration: %s", duration)
						}

						return nil
					}),
			},
		},
	})
}

func TestAccresourceRetriesOnError(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	mc := &mockClient{}

	// Second successful response
	resp2 := &http.Response{}
	resp2.StatusCode = http.StatusOK
	resp2.Body = ioutil.NopCloser(bytes.NewReader([]byte(json)))

	// Destroy response
	resp3 := &http.Response{}
	resp3.StatusCode = http.StatusNoContent
	resp3.Body = ioutil.NopCloser(bytes.NewReader([]byte(json)))

	mc.On("Do", mock.Anything).Once().Return(nil, fmt.Errorf("boom"))
	mc.On("Do", mock.Anything).Once().Return(resp2, nil)
	mc.On("Do", mock.Anything).Once().Return(resp3, nil)

	Client = mc

	// ensure default client is replaced after test
	defer func() {
		Client = &http.Client{}
	}()

	resource.UnitTest(t, resource.TestCase{
		CheckDestroy:      testAccCheckRequestDestroy,
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccresourceCurlBodyWithRetry(json),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if len(mc.Calls) != 2 {
							return fmt.Errorf("expected http request to be made 2 times")
						}
						return nil
					}),
			},
		},
	})
}

func TestAccresourceCurlBody(t *testing.T) {
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
