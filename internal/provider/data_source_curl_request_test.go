package provider

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jarcoal/httpmock"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccdataSourceCurlRequest(t *testing.T) {
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

//func TestAccdataSourceRetriesOnFailure(t *testing.T) {
//
//	rName := sdkacctest.RandomWithPrefix("devopsrob")
//	json := `{"name": "` + rName + `"}`
//
//	mc := &mockClient{}
//
//	resp1 := &http.Response{}
//	resp1.StatusCode = http.StatusInternalServerError
//	resp1.Body = ioutil.NopCloser(bytes.NewReader([]byte("boom")))
//
//	resp2 := &http.Response{}
//	resp2.StatusCode = http.StatusOK
//	resp2.Body = ioutil.NopCloser(bytes.NewReader([]byte(json)))
//
//	mc.On("Do", mock.Anything).Return(resp1, nil).Once()
//	mc.On("Do", mock.Anything).Return(resp2, nil)
//
//	Client = mc
//
//	// ensure default client is replaced after test
//	defer func() {
//		Client = &http.Client{}
//	}()
//
//	resource.UnitTest(t, resource.TestCase{
//		PreCheck:          func() { testAccPreCheck(t) },
//		ProviderFactories: providerFactories,
//		Steps: []resource.TestStep{
//			{
//				PlanOnly: true, // Read is called for the plan and apply phase, we only need to test read once
//				Config:   testAccdataSourceCurlBodyWithRetry(json),
//				Check: resource.ComposeTestCheckFunc(
//					func(s *terraform.State) error {
//						if len(mc.Calls) != 2 {
//							return fmt.Errorf("expected http request to be made 2 times. It was made %v times", len(mc.Calls))
//						}
//						return nil
//					},
//				),
//			},
//		},
//	})
//}

func TestAccdataSourceRetriesOnFailure(t *testing.T) {
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

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories, // Ensure this is correctly set up for your provider
		Steps: []resource.TestStep{
			{
				PlanOnly: true, // Assuming Read is called for the plan and apply phase
				Config:   testAccdataSourceCurlBodyWithRetry(json),
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
