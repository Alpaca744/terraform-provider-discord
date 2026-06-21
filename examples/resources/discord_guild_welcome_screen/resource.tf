resource "discord_guild_welcome_screen" "main" {
  guild_id    = "123456789012345678"
  enabled     = true
  description = "Welcome to our community!"

  welcome_channels = [
    {
      channel_id  = "111111111111111111"
      description = "Read the rules"
      emoji_name  = "📜"
    },
    {
      channel_id  = "222222222222222222"
      description = "Introduce yourself"
      emoji_name  = "👋"
    },
  ]
}
