# Terraform Provider TerraCurl

[![Terraform](https://img.shields.io/badge/terraform-%235835CC.svg?style=for-the-badge&logo=terraform&logoColor=white)](https://registry.terraform.io/providers/devops-rob/terracurl/latest/docs)

TerraCurl is an open-source Terraform provider that enables declarative, configurable HTTP API interactions directly from Terraform. It allows infrastructure and platform teams to integrate with most REST or HTTP-based services, including those without native Terraform providers, using standard Terraform workflows.

With TerraCurl, you can define create, read, update, and delete operations as HTTP requests, complete with support for custom headers, authentication, TLS configuration, retries, and response parsing. This makes it possible to manage third-party APIs, internal services, and bespoke platforms as first-class Terraform resources, without resorting to brittle null_resource hacks or external scripts.

TerraCurl is designed for reliability and correctness, supporting state reconciliation, drift detection, idempotency, and lifecycle control. It can trigger resource recreation when a remote state diverges from expected responses, handle ephemeral resources, and work with multipart and form-based APIs. By bringing arbitrary HTTP endpoints under Terraform’s declarative model, TerraCurl bridges the gap between “infrastructure as code” and “API as code”, enabling consistent automation, auditability, and repeatability across the entire platform stack.
## Join the community

[![Discord](https://img.shields.io/badge/Discord-%235865F2.svg?style=for-the-badge&logo=discord&logoColor=white)](https://discord.gg/Zc4raDkX4C)

## Managed API calls
When using TerraCurl, if the API call is creating a change on the target platform and you would like this change reversed upon a destroy, use the `terracurl_request` resource. This will allow you to enter the API call that should be run when `terraform destroy` is run.

```hcl
resource "terracurl_request" "mount" {
  name           = "vault-mount"
  url            = "https://localhost:8200/v1/sys/mounts/aws"
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
```
## Unmanaged API calls
For instances where there is no change required on the target platform when the `terraform destroy` command is run, use the `terracurl_request` data source

```hcl
data "terracurl_request" "test" {
  name           = "products"
  url            = "https://api.releases.hashicorp.com/v1/products"
  method         = "GET"

  response_codes = [
    200
  ]

  max_retry      = 1
  retry_interval = 10
}
```
## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.17

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:
```sh
$ go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, update the makefile with your OS architecture, then run `make install`. This will build the provider and put the provider binary in the `~/.terraform.d/plugins/local/` directory.

To generate or update documentation, run `make docs`.

In order to run the full suite of Acceptance tests, run `make test`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make test
```