resource "terracurl_request" "tls_test" {
  method          = "POST"
  name            = "test"
  response_codes  = ["200", "201", "204"]
  url             = "https://localhost:8200/v1/sys/mounts/aws"
  cert_file       = "path/to/cert/file"
  key_file        = "path/to/cert/file"
  ca_cert_file    = "path/to/cert/file"
  skip_tls_verify = false


  request_body = jsonencode({
    type        = "aws"
    description = "Enabling to test terracurl"
  })

  headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

  retry_interval = "5"
  max_retry      = "2"

  skip_read            = false
  read_url             = "https://localhost:8200/v1/sys/mounts/aws"
  read_method          = "GET"
  read_response_codes  = ["200"]
  read_cert_file       = "path/to/cert/file"
  read_key_file        = "path/to/cert/file"
  read_ca_cert_file    = "path/to/cert/file"
  read_skip_tls_verify = false


  read_headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

  ignore_response_fields = ["request_id"]

  skip_destroy            = false
  destroy_method          = "DELETE"
  destroy_url             = "http://localhost:8200/v1/sys/mounts/aws"
  destroy_response_codes  = ["204"]
  destroy_cert_file       = "path/to/cert/file"
  destroy_key_file        = "path/to/cert/file"
  destroy_ca_cert_file    = "path/to/cert/file"
  destroy_skip_tls_verify = false


  destroy_headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

}
