resource "discord_role" "verified" {
  guild_id = "123456789012345678"
  name     = "Verified"
}

resource "discord_member_role" "alice_verified" {
  guild_id = "123456789012345678"
  user_id  = "222222222222222222"
  role_id  = discord_role.verified.id
}
