# Some APIs acknowledge a write with a status envelope rather than the object that
# was created, for example:
#
#   POST /api/domain        -> {"Success": true, "Message": "Domain registered"}
#   POST /api/domain/detail -> {"Domain": "example.com", "Owners": [...], ...}
#
# Storing the write acknowledgement in state makes drift detection compare the
# envelope against the object returned by the read endpoint. Those two shapes can
# never match, so every plan reports drift and proposes a replacement.
#
# Setting read_after_write causes TerraCurl to call the read endpoint immediately
# after a successful create or update and persist that response instead, so later
# reads compare like for like.

resource "terracurl_request" "domain" {
  name           = "domain-registration"
  url            = "https://api.example.com/api/domain"
  method         = "POST"
  response_codes = ["200"]

  request_body = jsonencode({
    domain = "example.com"
    owners = ["team@example.com"]
  })

  headers = {
    Content-Type = "application/json"
  }

  # Persist the read response in state instead of the write acknowledgement.
  read_after_write = true

  # Setting skip_read to false additionally re-reads on refresh, so changes made
  # outside Terraform are detected and reconciled.
  skip_read           = false
  read_url            = "https://api.example.com/api/domain/detail"
  read_method         = "POST"
  read_response_codes = ["200"]

  read_request_body = jsonencode({
    domain = "example.com"
  })

  read_headers = {
    Content-Type = "application/json"
  }

  # Exclude volatile fields the API returns on read but that are not configuration.
  ignore_response_fields = ["LastModified"]

  skip_destroy           = false
  destroy_url            = "https://api.example.com/api/domain"
  destroy_method         = "DELETE"
  destroy_response_codes = ["200", "204"]

  destroy_request_parameters = {
    domain = "example.com"
  }
}
