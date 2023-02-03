package provider

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jarcoal/httpmock"
	"testing"
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
  destroy_url    = "http://example.com"

  request_body = <<EOF
%s
EOF
  method         = "POST"
  response_codes = ["200","204"]
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

//func TestAccresourceTls(t *testing.T) {
//	rName := sdkacctest.RandomWithPrefix("devopsrob")
//	json := `{"name": "` + rName + `"}`
//
//	httpmock.ActivateTLSTransport()
//	defer httpmock.DeactivateAndReset()
//	httpmock.RegisterResponder("GET", "https://example.com/api",
//		httpmock.NewStringResponder(200, `{"message": "Hello, world!"}`))
//
//	resource.UnitTest(t, resource.TestCase{
//		CheckDestroy:      testAccCheckRequestDestroy,
//		PreCheck:          func() { testAccPreCheck(t) },
//		ProviderFactories: providerFactories,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccresourceTls(json),
//				Check: func(s *terraform.State) error {
//					// Send a request to the mock responder
//					req, _ := http.NewRequest("GET", "https://example.com/api", nil)
//					res, err := http.DefaultClient.Do(req)
//					if err != nil {
//						return err
//					}
//					defer res.Body.Close()
//
//					// Check the response from the mock responder
//					if res.StatusCode != 200 {
//						return fmt.Errorf("unexpected status code: %d", res.StatusCode)
//					}
//
//					return nil
//				},
//			},
//		},
//	})
//}
//
//func testAccresourceTls(body string) string {
//	return fmt.Sprintf(`
//resource "terracurl_request" "example" {
//  name           = "leader"
//  url            = "https://example.com/create"
//  request_body = <<EOF
//%s
//EOF
//  method         = "POST"
//  response_codes = ["200"]
//  destroy_url    = "https://example.com/destroy"
//  destroy_method = "POST"
//  destroy_response_codes = ["204"]
//  # cert_file = "/tls/server-terracurl-0.pem"
//  # key_file  = "/tls/server-terracurl-0-key.pem"
//  # ca_cert_file = "/tls/terracurl-ca.pem"
//  # skip_tls_verify = true
//
//}
//
//`, body)
//}
