terraform {
  required_providers {
    terracurl = {
      source  = "devops-rob/terracurl"
      version = "1.0.1"
    }
  }
}

provider "terracurl" {}

resource "terracurl_request" "mount" {
  name         = "vault-mount"
  url          = "https://localhost:8200/v1/sys/mounts/aws"
  method       = "POST"
  request_body = <<EOF
{
  "type": "aws",
  "config": {
    "force_no_cache": true
  }
}

EOF

  headers = {
    X-Vault-Token = "root"
  }

  response_codes = [
    200,
    204
  ]

  cert_file       = "server-vault-0.pem"
  key_file        = "server-vault-0-key.pem"
  ca_cert_file    = "vault-server-ca.pem"
  skip_tls_verify = false


  destroy_url    = "https://localhost:8200/v1/sys/mounts/aws"
  destroy_method = "DELETE"

  destroy_headers = {
    X-Vault-Token = "root"
  }

  destroy_response_codes = [
    204
  ]

  destroy_cert_file       = "server-vault-0.pem"
  destroy_key_file        = "server-vault-0-key.pem"
  destroy_ca_cert_file    = "vault-server-ca.pem"
  destroy_skip_tls_verify = false

}

output "response" {
  value = terracurl_request.mount.response
}