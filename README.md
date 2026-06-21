# terraform-provider-discord

[![CI](https://github.com/alpaca744/terraform-provider-discord/actions/workflows/test.yml/badge.svg)](https://github.com/alpaca744/terraform-provider-discord/actions/workflows/test.yml)
[![Release](https://github.com/alpaca744/terraform-provider-discord/actions/workflows/release.yml/badge.svg)](https://github.com/alpaca744/terraform-provider-discord/releases)

A Terraform provider for managing durable Discord configuration through the
Discord REST API (v10), built on the
[HashiCorp Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

**21 resources · 17 data sources** — guild management, channels, webhooks,
roles, permissions, application commands, emoji, stickers, soundboard sounds,
scheduled events, onboarding, templates, monetization, and more.

---

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) ≥ 1.0
- A Discord bot token (see [Getting credentials](#getting-credentials))

## Installation

```hcl
terraform {
  required_providers {
    discord = {
      source  = "alpaca744/discord"
      version = "~> 0.1"
    }
  }
}
```

Run `terraform init` to download the provider from the
[Terraform Registry](https://registry.terraform.io/providers/alpaca744/discord).

## Provider configuration

```hcl
provider "discord" {
  bot_token = var.discord_bot_token
}
```

| Argument | Environment variable | Required | Description |
|---|---|---|---|
| `bot_token` | `DISCORD_BOT_TOKEN` | yes | Discord bot token |
| `bearer_token` | `DISCORD_BEARER_TOKEN` | no | OAuth2 bearer token (command-permission endpoints only) |
| `client_id` | `DISCORD_CLIENT_ID` | no | Application client ID |
| `client_secret` | `DISCORD_CLIENT_SECRET` | no | Application client secret |
| `api_base_url` | `DISCORD_API_BASE_URL` | no | Override the Discord API base URL (testing) |
| `default_audit_log_reason` | `DISCORD_AUDIT_LOG_REASON` | no | Default reason string written to the guild audit log |

Credentials are never written to state or surfaced in diagnostics.

## Quick example

```hcl
resource "discord_role" "moderator" {
  guild_id     = var.guild_id
  name         = "Moderator"
  color        = 0x3498db
  hoist        = true
  mentionable  = true
  permissions  = ["KICK_MEMBERS", "BAN_MEMBERS", "MANAGE_MESSAGES"]
}

resource "discord_channel" "announcements" {
  guild_id = var.guild_id
  name     = "announcements"
  type     = 5   # announcement channel
}

resource "discord_channel_permission_overwrite" "mod_view" {
  channel_id   = discord_channel.announcements.id
  overwrite_id = discord_role.moderator.id
  type         = "role"
  allow        = ["VIEW_CHANNEL", "SEND_MESSAGES"]
  deny         = []
}
```

More examples for every resource and data source live under [`examples/`](examples/).

## Getting credentials

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
   and create (or open) an application.
2. Navigate to **Bot** → **Token** → **Reset Token** and copy the value.
3. Export it before running Terraform:

   ```bash
   export DISCORD_BOT_TOKEN="your-bot-token"
   ```

4. Add the bot to your guild via the OAuth2 URL with the permissions your
   resources need (at minimum `bot` scope + the relevant guild permissions).

## Resources and data sources

| Resource | Description |
|---|---|
| `discord_role` | Guild role with permission bitfield |
| `discord_guild_settings` | Mutable guild properties (name, verification level, …) |
| `discord_channel` | Text, voice, category, announcement, forum, and stage channels |
| `discord_channel_permission_overwrite` | Per-channel permission overrides for roles and members |
| `discord_webhook` | Incoming webhooks |
| `discord_member_role` | Assigns a role to a guild member |
| `discord_auto_moderation_rule` | AutoMod keyword / mention-spam rules |
| `discord_guild_emoji` | Custom emoji |
| `discord_soundboard_sound` | Soundboard sounds |
| `discord_guild_sticker` | Custom stickers (PNG / APNG / GIF / Lottie) |
| `discord_invite` | Channel invites |
| `discord_guild_scheduled_event` | Guild scheduled events |
| `discord_stage_instance` | Stage channel instances |
| `discord_guild_template` | Guild templates |
| `discord_guild_widget` | Guild widget settings |
| `discord_guild_welcome_screen` | Community guild welcome screen |
| `discord_guild_onboarding` | Guild onboarding configuration |
| `discord_application_settings` | Application description, tags, install URL |
| `discord_application_role_connection_metadata` | Linked-role metadata records |
| `discord_application_command` | Global application commands |
| `discord_guild_application_command` | Guild-scoped application commands |

| Data source | Description |
|---|---|
| `discord_guild` | Reads a guild by ID |
| `discord_guilds` | Lists guilds the bot is in |
| `discord_guild_preview` | Public preview of a discoverable guild |
| `discord_role` | Reads a single guild role |
| `discord_roles` | Lists all roles in a guild |
| `discord_channel` | Reads a channel by ID |
| `discord_channels` | Lists channels in a guild |
| `discord_webhook` | Reads a webhook by ID |
| `discord_invite` | Reads an invite by code |
| `discord_audit_log` | Reads recent guild audit log entries |
| `discord_current_user` | Reads the authenticated bot user |
| `discord_user` | Reads any user by ID |
| `discord_voice_regions` | Lists available voice regions |
| `discord_current_application` | Reads the current application |
| `discord_sku` | Lists application SKUs |
| `discord_entitlement` | Lists application entitlements |
| `discord_subscription` | Lists subscriptions for a SKU |

## Development

Requirements: Go (version declared in `go.mod`) and `make`.

```bash
make build      # compile the provider binary
make test       # unit and contract tests (no credentials needed)
make check      # gofmt + go vet + tests
make testacc    # acceptance tests against the live Discord API (see below)
make docs       # regenerate docs/ from schema
make lint       # golangci-lint
```

### Running acceptance tests

Acceptance tests hit the real Discord API and are gated behind `TF_ACC=1`:

```bash
export TF_ACC=1
export DISCORD_BOT_TOKEN="..."
export DISCORD_TEST_GUILD_ID="..."          # ID of a bot-owned test guild
export DISCORD_TEST_APPLICATION_ID="..."   # bot application ID

# Optional — required only for specific test suites:
export DISCORD_TEST_COMMUNITY_GUILD_ID="..." # Community-enabled guild (welcome screen, stage channels)
export DISCORD_TEST_SOUND_DATA_URI="data:audio/mpeg;base64,..." # real MP3/OGG for soundboard tests

make testacc
```

Each test cleans up the Discord objects it creates. Never run acceptance tests
against a production guild.

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
