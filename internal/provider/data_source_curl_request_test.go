package provider

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/jarcoal/httpmock"
	"regexp"
	"testing"

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

//const testAccdataSourceCurlRequest = `
//data "terracurl_request" "test" {
//  name           = "leader"
//  url            = "https://example.com"
//  method         = "GET"
//}
//`

func testAccdataSourceCurlRequest(body string) string {
	return fmt.Sprintf(`
data "terracurl_request" "test" {
  name           = "leader"
  url            = "https://example.com/create"
  request_body = <<EOF
%s
EOF
  method         = "POST"
}
`, body)
}
