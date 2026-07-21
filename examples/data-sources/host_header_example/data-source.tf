# Override the HTTP Host header independently of the URL hostname.
# Useful for CDN origin checks, virtual host routing, and end-to-end tests.

data "terracurl_request" "virtual_host" {
  name   = "virtual-host"
  url    = "https://origin.example.com"
  method = "GET"

  headers = {
    Host = "www.example.com"
  }

  response_codes = ["200"]
}
