data "terracurl_request" "headers_example" {
  method                 = "GET"
  name                   = "basics"
  response_codes         = ["200"]
  url                    = "http://example.com"

  request_parameters = {
    parameter_key   = "parameter_value"
    parameter_key_2 = "another_parameter_value"
  }

  request_body = <<EOF
{
  "name": "devopsrob",
  "project": "TerraCurl V2"
}
EOF
}
