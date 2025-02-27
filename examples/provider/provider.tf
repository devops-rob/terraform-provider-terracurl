terraform {
  required_providers {
    terracurl = {
      source  = "devops-rob/terracurl"
      version = "1.0.0"
    }
  }
}

provider "terracurl" {}


#locals {
#  vault_token = jsondecode(ephemeral.terracurl_request.ephems.response).data
#  token = local.vault_token
#}
data "terracurl_request" "test" {
  method         = "POST"
  name           = "test"
  response_codes = ["200", "201", "204"]
  url            = "http://localhost:8200/v1/auth/token/create-orphan"

  headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

  request_body = jsonencode({
    policies        = ["web", "stage"]
    meta = {
      user = "devopsrob"
    }
    ttl = "1h"
    renewable = true
  })


}

output "response" {
  value = data.terracurl_request.test.response
}

#resource "terracurl_request" "test" {
#  method         = "POST"
#  name           = "test"
#  response_codes = ["200", "201", "204"]
#  url            = "http://localhost:8200/v1/sys/mounts/aws"
#
#  request_body = jsonencode({
#    type        = "aws"
#    description = "Enabling to test terracurl"
#  })
#
#  headers = {
#    X-Vault-Token = "root"
#    Content-Type  = "application/json"
#  }
#
#  skip_read           = false
#  read_url            = "http://localhost:8200/v1/sys/mounts/aws"
#  read_method         = "GET"
#  read_response_codes = ["200"]
#
#  read_headers = {
#    X-Vault-Token = "root"
#    Content-Type  = "application/json"
#  }
#
#  skip_destroy           = false
#  destroy_method         = "DELETE"
#  destroy_url            = "http://localhost:8200/v1/sys/mounts/aws"
#  destroy_response_codes = ["204"]
#
#  destroy_headers = {
#    X-Vault-Token = "root"
#    Content-Type  = "application/json"
#  }
#
#  ignore_response_fields = ["request_id"]
#
#}
#
##output "response" {
##  value = ephemeral.terracurl_request.ephems.response
##  ephemeral = true
##}
#
#ephemeral "terracurl_request" "ephems" {
#  method         = "POST"
#  name           = "test"
#  response_codes = ["200", "201", "204"]
#  url            = "http://localhost:8200/v1/sys/unseal"
#
#  request_body = <<EOF
#{
#  "key": "pgqeCXHfz4Tx/cwlX+Y7ONIVGT48z9Vj3dsbf0B8Vow="
#}
#
#EOF
#
#  headers = {
#    X-Vault-Token = "root"
#    Content-Type  = "application/json"
#  }
#
#}
