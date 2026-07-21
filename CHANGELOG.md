## 2.5.0

FEATURES:

- Add optional `response_sensitive` and sensitive response attributes to `terracurl_request` resource, data source, and ephemeral resource to prevent secret values appearing in plan output. Based on #144 by @ohaibbq. Closes #142.

## 2.4.2

BUG FIXES:

- Fix false drift when `Create()` stored unsanitized HTTP responses while `Read()` used sanitized JSON (fixes Create/Read mismatch for `response` and `ignore_response_fields`). Based on #141 by @rrrix.

## 2.4.1

ENHANCEMENTS:

- Upgrade HashiCorp plugin dependencies: `terraform-plugin-framework` v1.19.0, `terraform-plugin-go` v0.31.0, `terraform-plugin-testing` v1.16.0
- Supersedes Dependabot PRs #138, #139, and #140

## 2.4.0

FEATURES:

- Add `terracurl_request` action support for on-demand HTTP requests via `terraform apply -invoke` (Terraform 1.14+). Based on #127 by @jen20.

## 2.3.1

BUG FIXES:

- Fix Host header override when `Host` is set in request header maps (fixes #79). Based on #97 by @JoshBlades.

## 2.3.0

FEATURES:

- Add HTTP and HTTPS proxy support via environment variables and provider attributes (`http_proxy`, `https_proxy`, `no_proxy`)
- Proxy support applies to TLS/mTLS requests as well as plain HTTP requests
- Add HTTP Proxy Support guide, examples, and tests

## 2.2.0 

FEATURES:

- Bug fixes with state upgrade
- Documentation updates
