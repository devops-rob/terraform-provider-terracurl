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
