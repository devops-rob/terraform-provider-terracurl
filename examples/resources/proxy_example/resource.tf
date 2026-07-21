# Outbound requests use provider-level proxy settings.
# No per-resource proxy configuration is required.

resource "terracurl_request" "via_proxy" {
  name           = "proxy-example"
  url            = "https://api.example.com/v1/resources"
  method         = "GET"
  response_codes = ["200"]
  skip_destroy   = true
}
