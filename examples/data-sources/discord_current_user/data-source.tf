data "discord_current_user" "this" {}

output "bot_username" {
  value = data.discord_current_user.this.username
}
