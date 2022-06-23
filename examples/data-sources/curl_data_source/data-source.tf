terraform {
  required_providers {
    terracurl = {
      source  = "local/devops-rob/terracurl"
      version = "0.1.0"
    }
  }
}

provider "terracurl" {}

data "terracurl_request" "test" {
  name           = "products"
  url            = "https://api.releases.hashicorp.com/v1/products"
  method         = "GET"
}

output "response" {
  value = jsondecode(data.terracurl_request.test.response)
}