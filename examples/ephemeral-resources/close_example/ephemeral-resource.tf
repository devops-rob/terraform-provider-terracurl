ephemeral "terracurl_request" "ephems" {
  method         = "POST"
  name           = "test"
  response_codes = ["201", "200", "205"]
  url            = "http://localhost:8200/v1/sys/unseal"

  headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

  request_body = jsonencode({
    "key" = "pgqeCXHfz4Tx/cwlX+Y7ONIVGT48z9Vj3dsbf0B8Vow="
  })

  skip_renew           = true

  skip_close           = false
  close_url            = "http://localhost:8200/v1/sys/seal"
  close_response_codes = ["201", "200", "204"]
  close_method         = "POST"

  close_headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

}
