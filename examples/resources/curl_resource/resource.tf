resource "curl_request" "hashicorp_products" {
  name           = "releases"
  url            = "https://api.releases.hashicorp.com/v1/products"
  method         = "GET"
  destroy_url    = "https://api.releases.hashicorp.com/v1/products"
  destroy_method = ""
}