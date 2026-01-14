---
page_title: "Provider: TerraCurl"
description: |-
  TODO
---

# TERRACURL Provider

TerraCurl is an open-source Terraform provider that enables declarative, configurable HTTP API interactions directly from Terraform. It allows infrastructure and platform teams to integrate with most REST or HTTP-based services, including those without native Terraform providers, using standard Terraform workflows.

With TerraCurl, you can define create, read, update, and delete operations as HTTP requests, complete with support for custom headers, authentication, TLS configuration, retries, and response parsing. This makes it possible to manage third-party APIs, internal services, and bespoke platforms as first-class Terraform resources, without resorting to brittle null_resource hacks or external scripts.

TerraCurl is designed for reliability and correctness, supporting state reconciliation, drift detection, idempotency, and lifecycle control. It can trigger resource recreation when remote state diverges from expected responses, handle ephemeral resources, and work with multipart and form-based APIs. By bringing arbitrary HTTP endpoints under Terraform’s declarative model, TerraCurl bridges the gap between “infrastructure as code” and “API as code”, enabling consistent automation, auditability, and repeatability across the entire platform stack.

Use the navigation to the left to read about the available resources.

## Example Usage

As of Terraform 1.8 and later, providers can implement functions that you can call from the Terraform configuration.

Define the provider as a `required_provider` to use its functions

```terraform
terraform {
  required_providers {
    terracurl = {
      source  = "devops-rob/terracurl"
      version = "2.0.0"
    }
  }
}

provider "terracurl" {}

ephemeral "terracurl_request" "ephems" {
  method         = "GET"
  name           = "test"
  response_codes = ["201", "200", "205"]
  url            = "http://localhost:8200/v1/rabbitmq/creds/administrator"

  headers = {
    X-Vault-Token = "root"
    Content-Type  = "application/json"
  }

  skip_renew = true
  skip_close = true

}

```

## Limitations

- Currently, this provider does not support AWS APIs due to the nature of how they authenticate requests. Future AWS support is planned.
- Write-only attributes are currently not supported; however, there are plans to introduce this in the short-to-medium term future.
- Terraform Actions support is currently not supported, but work is well under way to introduce this functionality.
