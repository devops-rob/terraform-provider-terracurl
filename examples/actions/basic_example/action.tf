terraform {
  required_providers {
    terracurl = {
      source = "devops-rob/terracurl"
    }
  }
}

action "terracurl_request" "basic_example" {
  config {
    method         = "GET"
    response_codes = ["200"]
    url            = "http://example.com"
  }
}
