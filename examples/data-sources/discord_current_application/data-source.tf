data "discord_current_application" "this" {}

output "application_id" {
  value = data.discord_current_application.this.id
}
