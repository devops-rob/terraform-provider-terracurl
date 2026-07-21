# Proxy can be configured with provider attributes:
#
# provider "terracurl" {
#   http_proxy  = "http://proxy.example.com:8080"
#   https_proxy = "http://proxy.example.com:8080"
#   no_proxy    = "localhost,127.0.0.1,.internal.example.com"
# }
#
# Alternatively, configure the environment where Terraform runs:
#
# export HTTP_PROXY="http://proxy.example.com:8080"
# export HTTPS_PROXY="http://proxy.example.com:8080"
# export NO_PROXY="localhost,127.0.0.1,.internal.example.com"

provider "terracurl" {
  http_proxy  = "http://proxy.example.com:8080"
  https_proxy = "http://proxy.example.com:8080"
  no_proxy    = "localhost,127.0.0.1,.internal.example.com"
}
