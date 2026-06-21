resource "discord_webhook" "deploys" {
  channel_id = "123456789012345678"
  name       = "Deploy Notifications"
}

output "deploy_webhook_url" {
  value     = discord_webhook.deploys.url
  sensitive = true
}
