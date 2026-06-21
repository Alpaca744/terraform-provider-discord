data "discord_guilds" "mine" {}

output "guild_names" {
  value = [for g in data.discord_guilds.mine.guilds : g.name]
}
