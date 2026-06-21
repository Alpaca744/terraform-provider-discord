resource "discord_application_command" "ping" {
  application_id = "123456789012345678"
  name           = "ping"
  description     = "Replies with pong"
  type           = 1 # CHAT_INPUT (slash command)

  # Options are JSON to support arbitrarily nested subcommands and groups.
  options = jsonencode([
    {
      type        = 3 # STRING
      name        = "message"
      description = "Optional message to echo"
      required    = false
    },
  ])
}
