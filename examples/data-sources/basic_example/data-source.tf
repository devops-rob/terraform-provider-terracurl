data "terracurl_request" "headers_example" {
  method                 = "GET"
  name                   = "basics"
  response_codes         = ["200"]
  url                    = "http://example.com"

  headers = {
    Authorization = "Bearer root"
  }
}
