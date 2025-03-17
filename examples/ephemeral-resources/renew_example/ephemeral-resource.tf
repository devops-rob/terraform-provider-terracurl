ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["200"]
  url            = "http://example.com/open"

  skip_renew           = false
  renew_interval       = "-10"
  renew_url            = "http://example.com/renew"
  renew_response_codes = ["200"]
  renew_method         = "GET"

  skip_close = true

}
