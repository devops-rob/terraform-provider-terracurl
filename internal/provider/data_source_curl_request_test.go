package provider

import (
	"bytes"
	"context"
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccdataSourceCurlRequest(t *testing.T) {

	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccdataSourceCurlRequest(json),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.terracurl_request.test", "name", regexp.MustCompile("^leader")),
					resource.TestMatchResourceAttr(
						"data.terracurl_request.test", "response", regexp.MustCompile(`^{"name": "devopsrob"}`)),
					resource.TestMatchResourceAttr(
						"data.terracurl_request.test", "url", regexp.MustCompile("^https://example.com")),
					resource.TestMatchResourceAttr(
						"data.terracurl_request.test", "request_body", regexp.MustCompile(json)),
					resource.TestMatchResourceAttr(
						"data.terracurl_request.test", "method", regexp.MustCompile("^POST")),
				),
			},
		},
	})
}

func testAccdataSourceCurlRequest(body string) string {
	return fmt.Sprintf(`
data "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  request_body = <<EOF
%s
EOF
  method         = "POST"
  response_codes = [200]
}
`, body)
}

func TestAccdataSourceRetriesOnFailure(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	mc := &mockClient{}

	resp1 := &http.Response{}
	resp1.StatusCode = http.StatusInternalServerError
	resp1.Body = ioutil.NopCloser(bytes.NewReader([]byte("boom")))

	resp2 := &http.Response{}
	resp2.StatusCode = http.StatusOK
	resp2.Body = ioutil.NopCloser(bytes.NewReader([]byte(json)))

	mc.On("Do", mock.Anything).Return(resp1, nil).Once()
	mc.On("Do", mock.Anything).Return(resp2, nil)

	Client = mc

	// ensure default client is replaced after test
	defer func() {
		Client = &http.Client{}
	}()

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly: true, // Read is called for the plan and apply phase, we only need to test read once
				Config:   testAccdataSourceCurlBodyWithRetry(json),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if len(mc.Calls) != 2 {
							return fmt.Errorf("expected http request to be made 2 times. It was made %v times", len(mc.Calls))
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
 url            = "https://example.com/create"
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

func testAccdataSourceCurlBodyWithTimeout(body string) string {
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
func TestAccdataSourceCurlRequestTimeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("devopsrob")
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
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				Config:   testAccdataSourceCurlBodyWithTimeout(json),
				Check: resource.ComposeTestCheckFunc(
					testCheckDurationWithTimeout(),
				),
			},
		},
	})
	duration := time.Since(start)        // Calculate the duration of the API call
	t.Logf("API call took %v", duration) // Log the duration
}

func testCheckDurationWithTimeout() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resource := state.RootModule().Resources["data.terracurl_request.test"]
		durationStr := resource.Primary.Attributes["duration_milliseconds"]
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
		if !strings.Contains(resource.Primary.Attributes["response"], "context canceled, not retrying operation") &&
			!strings.Contains(resource.Primary.Attributes["response"], "request failed, retries exceeded") {
			return fmt.Errorf("expected error message not found in response")
		}
		return nil
	}
}

func TestAccMyResourceWithTimeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("devopsrob")
	json := `{"name": "` + rName + `"}`

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			time.Sleep(1 * time.Second) // Delay the response by 6 seconds
			return httpmock.NewStringResponse(200, `{"name": "devopsrob"}`), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:   testAccdataSourceCurlBodyWithTimeout(json),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMyResourceWithTimeout("my_resource", 1*time.Second),
				),
			},
		},
	})
}

func testAccCheckMyResourceWithTimeout(name string, timeout time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				// Timeout exceeded, test passed
				return nil
			}
			return fmt.Errorf("unexpected context error: %v", ctx.Err())
		case <-time.After(2 * time.Second):
			// Timeout not exceeded, test failed
			return fmt.Errorf("resource %s did not timeout", name)
		}
	}
}
