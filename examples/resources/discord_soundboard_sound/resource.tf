resource "discord_soundboard_sound" "airhorn" {
  guild_id = "123456789012345678"
  name     = "airhorn"
  volume   = 0.8
  sound    = "data:audio/mpeg;base64,${filebase64("${path.module}/airhorn.mp3")}"
}
