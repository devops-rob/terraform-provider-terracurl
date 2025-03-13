package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jarcoal/httpmock"
	"net/http"
	"os"
	"testing"
	"time"
)

const testAccEphemeralResource = `
ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201"]
  url            = "https://example.com/open"
 
  skip_renew           = false
  renew_interval	   = "-10"
  renew_url            = "https://example.com/renew"
  renew_response_codes = ["200"]
  renew_method         = "GET"

  skip_close           = false
  close_url            = "https://example.com/close"
  close_response_codes = ["204"]
  close_method         = "DELETE"
  close_timeout		   = "20"

}

resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]
  method		 = "POST"
  retry_interval = "6"
  max_retry		 = "1"


  skip_destroy = true
  skip_read    = true

}`

func TestAccEphemeralResource(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("TF_LOG", "debug")
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

	var firstCall time.Time
	var callCount int

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/open",
		httpmock.NewStringResponder(201, `{"name": "devopsrob"}`),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/renew",
		httpmock.NewStringResponder(200, `{"name": "devopsrob"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/close",
		httpmock.NewStringResponder(204, ""),
	)
	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				firstCall = time.Now()
				time.Sleep(2 * time.Second)
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil

			}
			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEphemeralResource,
				Check: resource.ComposeTestCheckFunc(
					testMockEndpointRegister("POST https://example.com/open"),
					testMockEndpointRegister("GET https://example.com/renew"),
					//testMockEndpointCount("GET https://example.com/renew", 1),
					//testMockEndpointCount("DELETE https://example.com/close", 1),

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
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					//testMockEndpointRegister("GET https://example.com/renew"),
					testMockEndpointRegister("DELETE https://example.com/close"),
				),
			},
		},
	})

}

const testAccEphemeralResourceWithHeaders = `
ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201"]
  url            = "https://example.com/open"

  headers = {
    Authorization = "Bearer root"
    Content-Type  = "application/json"
  }
  
  skip_renew           = false
  renew_interval	   = "-10"
  renew_url            = "https://example.com/renew"
  renew_response_codes = ["200"]
  renew_method         = "GET"

  renew_headers = {
    Authorization = "Bearer root"
    Content-Type  = "application/json"
  }

  skip_close           = false
  close_url            = "https://example.com/close"
  close_response_codes = ["204"]
  close_method         = "DELETE"
  close_timeout		   = "20"

  close_headers = {
    Authorization = "Bearer root"
  }


}

resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]
  method		 = "POST"
  retry_interval = "6"
  max_retry		 = "1"


  skip_destroy = true
  skip_read    = true

}


`

func TestAccEphemeralResourceWithHeaders(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("TF_LOG", "debug")
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

	//var firstCall time.Time
	var callCount int

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				time.Sleep(2 * time.Second)
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil

			}
			return httpmock.NewStringResponse(200, ""), nil
		},
	)
	httpmock.RegisterResponder("POST", "https://example.com/open",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") == "Bearer root" {
				return httpmock.NewStringResponse(201, `{"message": "success"}`), nil
			}
			return httpmock.NewStringResponse(400, `{"error": "missing or invalid header"}`), nil
		})
	httpmock.RegisterResponder("GET", "https://example.com/renew",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") == "Bearer root" {
				return httpmock.NewStringResponse(200, `{"message": "success"}`), nil
			}
			return httpmock.NewStringResponse(400, `{"error": "missing or invalid header"}`), nil
		})
	httpmock.RegisterResponder("DELETE", "https://example.com/close",
		func(req *http.Request) (*http.Response, error) {
			// Check for the required header
			if req.Header.Get("Authorization") == "Bearer root" {
				return httpmock.NewStringResponse(204, `{"message": "success"}`), nil
			}
			// Return a different response if the header is missing or incorrect
			return httpmock.NewStringResponse(400, `{"error": "missing or invalid header"}`), nil
		})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEphemeralResourceWithHeaders,
				Check: resource.ComposeTestCheckFunc(
					testMockEndpointRegister("POST https://example.com/open"),
					testMockEndpointRegister("GET https://example.com/renew"),
					testMockEndpointRegister("DELETE https://example.com/close"),
				),
			},
		},
	})

}

const testAccEphemeralResourceWithParameters = `
ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201"]
  url            = "https://example.com/open"

  request_parameters = {
	params = "true"
  }
 
  skip_renew           = false
  renew_interval	   = "-10"
  renew_url            = "https://example.com/renew"
  renew_response_codes = ["200"]
  renew_method         = "GET"
  renew_request_parameters = {
	params = "true"
  }


  skip_close           = false
  close_url            = "https://example.com/close"
  close_response_codes = ["204"]
  close_method         = "DELETE"
  close_timeout		   = "20"

  close_request_parameters = {
	params = "true"
  }

}

resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]
  method		 = "POST"
  retry_interval = "6"
  max_retry		 = "1"


  skip_destroy = true
  skip_read    = true

}`

func TestAccEphemeralResourceWithParameters(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("TF_LOG", "debug")
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

	//var firstCall time.Time
	var callCount int

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/open?params=true",
		httpmock.NewStringResponder(201, `{"message": "success"}`),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/renew?params=true",
		httpmock.NewStringResponder(200, `{"message": "success"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/close?params=true",
		httpmock.NewStringResponder(204, ""),
	)
	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				time.Sleep(2 * time.Second)
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil

			}
			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEphemeralResourceWithParameters,
				Check: resource.ComposeTestCheckFunc(
					testMockEndpointRegister("POST https://example.com/open?params=true"),
					testMockEndpointRegister("GET https://example.com/renew?params=true"),
					testMockEndpointRegister("DELETE https://example.com/close?params=true"),
				),
			},
		},
	})

}

const testAccEphemeralResourceSkipRenew = `
ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201"]
  url            = "https://example.com/open"
 
  skip_renew           = true
  renew_interval	   = "-10"
  renew_url            = "https://example.com/renew"
  renew_response_codes = ["200"]
  renew_method         = "GET"

  skip_close           = false
  close_url            = "https://example.com/close"
  close_response_codes = ["204"]
  close_method         = "DELETE"
  close_timeout		   = "20"


}

resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]
  method		 = "POST"
  retry_interval = "6"
  max_retry		 = "1"


  skip_destroy = true
  skip_read    = true

}`

func TestAccEphemeralResourceSkipRenew(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("TF_LOG", "debug")
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

	//var firstCall time.Time
	var callCount int

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/open",
		httpmock.NewStringResponder(201, `{"message": "success"}`),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/renew",
		httpmock.NewStringResponder(200, `{"message": "success"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/close",
		httpmock.NewStringResponder(204, ""),
	)
	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				time.Sleep(2 * time.Second)
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil

			}
			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEphemeralResourceSkipRenew,
				Check: resource.ComposeTestCheckFunc(
					testMockEndpointCount("GET https://example.com/renew", 0),
				),
			},
		},
	})

}

const testAccEphemeralResourceSkipClose = `
ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201"]
  url            = "https://example.com/open"
 
  skip_renew           = true

  skip_close           = true
  close_url            = "https://example.com/close"
  close_response_codes = ["204"]
  close_method         = "DELETE"
  close_timeout		   = "20"


}

resource "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  response_codes = ["200"]
  method		 = "POST"
  retry_interval = "6"
  max_retry		 = "1"


  skip_destroy = true
  skip_read    = true

}`

func TestAccEphemeralResourceSkipClose(t *testing.T) {
	err := os.Setenv("TF_ACC", "true")
	if err != nil {
		return
	}
	err = os.Setenv("TF_LOG", "debug")
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

	//var firstCall time.Time
	var callCount int

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		"POST",
		"https://example.com/open",
		httpmock.NewStringResponder(201, `{"message": "success"}`),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://example.com/renew",
		httpmock.NewStringResponder(200, `{"message": "success"}`),
	)
	httpmock.RegisterResponder(
		"DELETE",
		"https://example.com/close",
		httpmock.NewStringResponder(204, ""),
	)
	httpmock.RegisterResponder("POST", "https://example.com/create",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				time.Sleep(2 * time.Second)
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil

			}
			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEphemeralResourceSkipClose,
				Check: resource.ComposeTestCheckFunc(
					testMockEndpointCount("GET https://example.com/close", 0),
				),
			},
		},
	})

}
