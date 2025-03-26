ephemeral "terracurl_request" "ephems" {
  method          = "POST"
  name            = "test"
  response_codes  = ["200"]
  url             = "https://example.com/open"
  cert_file       = "/path/to/cert/file"
  key_file        = "/path/to/key/file"
  ca_cert_file    = "/path/to/ca/cert/file"
  skip_tls_verify = true

  skip_renew            = false
  renew_interval        = "-10"
  renew_url             = "%s"
  renew_response_codes  = ["200"]
  renew_method          = "GET"
  renew_cert_file       = "/path/to/cert/file"
  renew_key_file        = "/path/to/key/file"
  renew_ca_cert_file    = "/path/to/ca/cert/file"
  renew_skip_tls_verify = false


  skip_close            = true
  close_url             = "%s"
  close_response_codes  = ["200"]
  close_method          = "DELETE"
  close_timeout         = "20"
  close_cert_file       = "/path/to/cert/file"
  close_key_file        = "/path/to/key/file"
  close_ca_cert_file    = "/path/to/ca/cert/file"
  close_skip_tls_verify = false

}
