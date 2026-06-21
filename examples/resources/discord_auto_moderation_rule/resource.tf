resource "discord_auto_moderation_rule" "block_keywords" {
  guild_id     = "123456789012345678"
  name         = "Block banned words"
  event_type   = 1 # MESSAGE_SEND
  trigger_type = 1 # KEYWORD
  enabled      = true

  # Nested attributes use HCL assignment syntax (`=`), not blocks.
  trigger_metadata = {
    keyword_filter = ["badword1", "badword2"]
  }

  actions = [
    {
      type           = 1 # BLOCK_MESSAGE
      custom_message = "That word is not allowed here."
    },
  ]
}
