
data "terracurl_request" "basic_example" {
  method                 = "GET"
  name                   = "basics"
  response_codes         = ["200"]
  url                    = "http://example.com"
}

