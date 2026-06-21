resource "discord_invite" "welcome" {
  channel_id = "123456789012345678"
  max_age    = 86400 # 24 hours
  max_uses   = 0     # unlimited
  temporary  = false
}

output "invite_url" {
  value = discord_invite.welcome.url
}
