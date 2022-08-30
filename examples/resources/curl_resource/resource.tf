terraform {
  required_providers {
    terracurl = {
      source  = "local/devops-rob/terracurl"
      version = "0.1.1"
    }
  }
}

provider "terracurl" {}

resource "terracurl_request" "mount" {
  name           = "vault-mount"
  url            = "http://localhost:8200/v1/sys/mounts/aws"
  method         = "POST"
  request_body   = <<EOF
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

  response_codes = [200,204]

  destroy_url    = "http://localhost:8200/v1/sys/mounts/aws"
  destroy_method = "DELETE"

  destroy_headers = {
    X-Vault-Token = "root"
  }

  destroy_response_codes = [204]
}

output "response" {
  value = terracurl_request.mount.response
}