data "discord_guild" "example" {
  id = "123456789012345678"
}

output "guild_name" {
  value = data.discord_guild.example.name
}
