resource "terracurl_request" "test" {
  method         = "POST"
  name           = "test"
  response_codes = ["200", "201", "204"]
  url            = "http://localhost:8200/v1/sys/mounts/aws"

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

  skip_read           = false
  read_url            = "http://localhost:8200/v1/sys/mounts/aws"
  read_method         = "GET"
  read_response_codes = ["200"]

  read_headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

  ignore_response_fields = ["request_id"]

  skip_destroy           = false
  destroy_method         = "DELETE"
  destroy_url            = "http://localhost:8200/v1/sys/mounts/aws"
  destroy_response_codes = ["204", "503"]

  destroy_headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

}
