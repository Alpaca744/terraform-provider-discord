data "discord_guild_preview" "example" {
  id = "123456789012345678"
}

output "member_count" {
  value = data.discord_guild_preview.example.approximate_member_count
}
