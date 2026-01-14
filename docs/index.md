---
page_title: "Provider: TerraCurl"
description: |-
  TODO
---

# TERRACURL Provider

TOOD

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
```

## Limitations
