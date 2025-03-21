terraform {
  required_providers {
    terracurl = {
      source  = "devops-rob/terracurl"
      version = "2.0.0"
    }
  }
}

provider "terracurl" {}
