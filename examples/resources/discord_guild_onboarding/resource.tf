resource "discord_guild_onboarding" "main" {
  guild_id            = "123456789012345678"
  enabled             = true
  mode                = 0 # ONBOARDING_DEFAULT
  default_channel_ids = ["111111111111111111"]

  # Prompts are JSON to support the deeply nested prompt/option structure.
  prompts = jsonencode([
    {
      id          = "0"
      type        = 0 # MULTIPLE_CHOICE
      title       = "What are you interested in?"
      single_select = false
      required    = false
      in_onboarding = true
      options = [
        {
          id          = "0"
          title       = "Gaming"
          channel_ids = []
          role_ids    = ["333333333333333333"]
        },
      ]
    },
  ])
}
