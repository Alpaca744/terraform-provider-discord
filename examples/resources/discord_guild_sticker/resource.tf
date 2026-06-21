resource "discord_guild_sticker" "wave" {
  guild_id    = "123456789012345678"
  name        = "wave"
  description = "A friendly wave"
  tags        = "wave"
  format      = "png"
  # Use filebase64 to read and encode the sticker file.
  file_content_base64 = filebase64("${path.module}/wave.png")
}
