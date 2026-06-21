resource "discord_channel" "announcements" {
  guild_id = "123456789012345678"
  name     = "announcements"
  type     = 0 # GUILD_TEXT
  topic    = "Server news and updates"
  nsfw     = false
}

resource "discord_channel" "voice_lounge" {
  guild_id   = "123456789012345678"
  name       = "Lounge"
  type       = 2 # GUILD_VOICE
  bitrate    = 64000
  user_limit = 10
}
