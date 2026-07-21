---
page_title: "HTTP Proxy Support"
subcategory: "Guides"
description: |-
  Configure outbound HTTP and HTTPS proxy settings for the TerraCurl provider.
---

# HTTP Proxy Support

TerraCurl routes all outbound HTTP requests through a proxy when configured. This applies to:

- `terracurl_request` resources (create, read, and destroy calls)
- `terracurl_request` data sources
- `terracurl_request` ephemeral resources (open, renew, and close calls)

Proxy settings apply to the Terraform process that runs the provider. They are not configured per resource.

## Environment variables

TerraCurl honors the standard Go proxy environment variables:

| Variable | Purpose |
|----------|---------|
| `HTTP_PROXY` / `http_proxy` | Proxy URL for HTTP requests |
| `HTTPS_PROXY` / `https_proxy` | Proxy URL for HTTPS requests |
| `NO_PROXY` / `no_proxy` | Comma-separated hosts that bypass the proxy |

Example shell configuration:

```bash
export HTTP_PROXY="http://proxy.example.com:8080"
export HTTPS_PROXY="http://proxy.example.com:8080"
export NO_PROXY="localhost,127.0.0.1,.internal.example.com"
```

In Terraform Cloud, Terraform Enterprise, or CI systems, set these variables in the workspace or job environment.

## Provider configuration

You can also configure proxy settings in the provider block. Provider attributes override environment variables when explicitly set.

```terraform
provider "terracurl" {
  http_proxy  = "http://proxy.example.com:8080"
  https_proxy = "http://proxy.example.com:8080"
  no_proxy    = "localhost,127.0.0.1,.internal.example.com"
}
```

Setting an attribute to an empty string disables that proxy setting, even if the corresponding environment variable is set.

## Precedence

1. Start from environment variable values
2. Override with any provider attribute that is explicitly set in Terraform configuration
3. Apply `no_proxy` matching to decide whether a given host bypasses the proxy

## Authenticated proxies

Proxy URLs may include credentials:

```terraform
provider "terracurl" {
  https_proxy = "http://user:password@proxy.example.com:8080"
}
```

TerraCurl uses Go's standard library proxy support. NTLM, Kerberos, and other enterprise authentication mechanisms are not supported.

## Limitations

- Proxy configuration is global to the provider instance; individual resources cannot specify different proxies.
- SOCKS proxies are not supported.
- Proxy settings affect outbound requests from the machine running Terraform, not remote APIs directly.
