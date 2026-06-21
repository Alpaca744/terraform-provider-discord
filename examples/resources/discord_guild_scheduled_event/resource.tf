resource "discord_guild_scheduled_event" "launch" {
  guild_id             = "123456789012345678"
  name                 = "Product Launch"
  description          = "Join us for the big reveal"
  entity_type          = 3 # EXTERNAL
  privacy_level        = 2 # GUILD_ONLY
  scheduled_start_time = "2026-07-01T17:00:00Z"
  scheduled_end_time   = "2026-07-01T18:00:00Z"
  location             = "https://example.com/live"
}
