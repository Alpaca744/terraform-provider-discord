data "discord_audit_log" "recent_bans" {
  guild_id    = "123456789012345678"
  action_type = 22 # MEMBER_BAN_ADD
  limit       = 10
}
