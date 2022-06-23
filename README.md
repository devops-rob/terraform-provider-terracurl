# Terraform Provider TerraCurl

Available in the [Terraform Registry.]()

This provider is designed to be a flexible extension of your terraform code to make managed and unamanged API calls to your target endpoint. Platform native providers should be preferred to TerraCurl but for instances where the platform provider does not have a resource or data source that you require, TerraCurl can be used to make substitute API calls.

## Managed API calls
When using TerraCurl, if the API call is creating a change on the target platform and you would like this change reversed upon a destroy, use the `terracurl_request` resource. This will allow you to enter the API call that should be run when `terraform destroy` is run.

```hcl
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
```
## Unmanaged API calls
For instances where there is no change required on the target platform when the `terraform destroy` command is run, use the `terracurl_request` data source

```hcl
data "terracurl_request" "test" {
  name           = "products"
  url            = "https://api.releases.hashicorp.com/v1/products"
  method         = "GET"
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

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
