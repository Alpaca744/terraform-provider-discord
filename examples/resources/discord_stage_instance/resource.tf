resource "discord_stage_instance" "weekly_qa" {
  channel_id    = "123456789012345678" # a stage channel
  topic         = "Weekly Q&A"
  privacy_level = 2 # GUILD_ONLY
}
