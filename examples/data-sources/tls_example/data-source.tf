
data "terracurl_request" "basic_example" {
  method          = "GET"
  name            = "basics"
  response_codes  = ["200"]
  url             = "https://example.com"
  cert_file       = "path/to/cert/file"
  key_file        = "path/to/cert/file"
  ca_cert_file    = "path/to/cert/file"
  skip_tls_verify = false
}

