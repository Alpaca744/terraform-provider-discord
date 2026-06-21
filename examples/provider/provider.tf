terraform {
  required_providers {
    discord = {
      source = "alpaca744/discord"
    }
  }
}

provider "discord" {
  # bot_token may also be supplied via the DISCORD_BOT_TOKEN environment variable.
  bot_token = var.discord_bot_token
}

variable "discord_bot_token" {
  type      = string
  sensitive = true
}
