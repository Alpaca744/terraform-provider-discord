resource "discord_guild_settings" "main" {
  guild_id                      = "123456789012345678"
  name                          = "My Community"
  description                   = "A friendly place"
  verification_level            = 2
  default_message_notifications = 1
  explicit_content_filter       = 2
  afk_timeout                   = 300
  premium_progress_bar_enabled  = true
}
