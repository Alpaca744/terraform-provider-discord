resource "discord_role" "moderator" {
  guild_id    = "123456789012345678"
  name        = "Moderator"
  color       = 3447003 # blurple
  hoist       = true
  mentionable = false

  permissions = [
    "VIEW_CHANNEL",
    "SEND_MESSAGES",
    "MANAGE_MESSAGES",
    "KICK_MEMBERS",
  ]
}
