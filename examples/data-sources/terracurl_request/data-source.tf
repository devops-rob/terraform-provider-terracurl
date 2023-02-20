terraform {
  required_providers {
    terracurl = {
      source  = "devops-rob/terracurl"
      version = "1.0.1"
    }
  }
}

provider "terracurl" {}

data "terracurl_request" "test" {
  name   = "products"
  url    = "https://api.releases.hashicorp.com/v1/products"
  method = "GET"

  response_codes = [
    200
  ]

  max_retry      = 1
  retry_interval = 10
}

output "response" {
  value = jsondecode(data.terracurl_request.test.response)
}