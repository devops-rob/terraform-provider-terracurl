data "terracurl_request" "headers_example" {
  method         = "GET"
  name           = "basics"
  response_codes = ["200"]
  url            = "http://example.com"

  request_body = jsonencode(
    {
      name    = "devopsrob"
      project = "TerraCurl v2"
    }
  )
}