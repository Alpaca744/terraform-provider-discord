data "discord_channels" "all" {
  guild_id = "123456789012345678"
}

output "text_channels" {
  value = [for c in data.discord_channels.all.channels : c.name if c.type == 0]
}
