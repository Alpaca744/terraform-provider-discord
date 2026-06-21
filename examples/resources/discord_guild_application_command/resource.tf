resource "discord_guild_application_command" "config" {
  application_id = "123456789012345678"
  guild_id       = "222222222222222222"
  name           = "config"
  description     = "Server configuration"
  type           = 1

  default_member_permissions = ["MANAGE_GUILD"]
}
