# Changelog

All notable changes to this provider are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-21

### Added

- Initial provider release covering the full Discord configuration surface:
  **21 resources** and **17 data sources** across guild management, application
  and command management, community features, media, and monetization.
- Discord API v10 client with dynamic rate-limit handling (X-RateLimit-Global +
  X-RateLimit-Scope), safe retries, structured diagnostics, and multipart upload
  support for sticker creation.
- Human-readable permission names (`"KICK_MEMBERS"`, `"BAN_MEMBERS"`, …) backed
  by Discord-compatible bitfields; stored in Terraform state as string sets.
- Full acceptance-test suite validated against the live Discord API — no mocks.
- End-to-end plan-correctness harness (`test/e2e/`) asserting a clean
  create → no-diff plan → update → import → destroy lifecycle via real Terraform
  core against an in-memory Discord mock.
- Generated Terraform Registry documentation and HCL examples for every resource
  and data source.
- GoReleaser configuration for multi-platform signed binaries
  (linux/darwin/windows/freebsd, amd64/arm64/arm/386).

### Resources

`discord_role`, `discord_guild_settings`, `discord_channel`,
`discord_channel_permission_overwrite`, `discord_webhook`, `discord_member_role`,
`discord_auto_moderation_rule`, `discord_guild_emoji`, `discord_soundboard_sound`,
`discord_guild_sticker`, `discord_invite`, `discord_guild_scheduled_event`,
`discord_stage_instance`, `discord_guild_template`, `discord_guild_widget`,
`discord_guild_welcome_screen`, `discord_guild_onboarding`,
`discord_application_settings`, `discord_application_role_connection_metadata`,
`discord_application_command`, `discord_guild_application_command`

### Data sources

`discord_guild`, `discord_guilds`, `discord_guild_preview`, `discord_role`,
`discord_roles`, `discord_channel`, `discord_channels`, `discord_webhook`,
`discord_invite`, `discord_audit_log`, `discord_current_user`, `discord_user`,
`discord_voice_regions`, `discord_current_application`, `discord_sku`,
`discord_entitlement`, `discord_subscription`

[0.1.0]: https://github.com/alpaca744/terraform-provider-discord/releases/tag/v0.1.0
