resource "discord_guild_emoji" "blobwave" {
  guild_id = "123456789012345678"
  name     = "blobwave"
  # Image must be a data URI. filebase64 + a data: prefix is one way to supply it.
  image = "data:image/png;base64,${filebase64("${path.module}/blobwave.png")}"
}
