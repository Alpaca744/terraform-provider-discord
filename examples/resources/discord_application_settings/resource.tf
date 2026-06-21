resource "discord_application_settings" "this" {
  description                       = "My bot does useful things."
  interactions_endpoint_url         = "https://example.com/interactions"
  role_connections_verification_url = "https://example.com/verify"
  tags                              = ["utility", "moderation"]
}
