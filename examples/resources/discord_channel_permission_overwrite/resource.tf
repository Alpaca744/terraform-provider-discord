resource "discord_channel_permission_overwrite" "mods_can_manage" {
  channel_id   = "123456789012345678"
  overwrite_id = "987654321098765432" # a role or member ID
  type         = "role"

  allow = ["VIEW_CHANNEL", "MANAGE_MESSAGES"]
  deny  = ["SEND_MESSAGES"]
}
