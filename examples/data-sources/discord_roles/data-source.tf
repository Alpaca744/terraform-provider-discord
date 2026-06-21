data "discord_roles" "all" {
  guild_id = "123456789012345678"
}

output "role_names" {
  value = [for r in data.discord_roles.all.roles : r.name]
}
