# Terraform Provider Discord Design

## Purpose

The Discord Terraform provider manages durable Discord configuration through
Terraform. Its quality bar is inspired by large, mature providers such as
`hashicorp/terraform-provider-aws`: broad API coverage, stable schemas, import
support, generated documentation, acceptance tests, clear diagnostics, and safe
release automation.

The provider should automate every stable Discord configuration surface exposed
by public APIs. Runtime/event-oriented APIs should only be included when they
represent durable configuration rather than transient behavior.

## Goals

- Manage Discord guild/server configuration declaratively.
- Manage Discord application and bot configuration exposed by public APIs.
- Support imports for every resource where Discord exposes stable object IDs.
- Provide generated Terraform Registry documentation for every resource and data
  source.
- Provide professional diagnostics for Discord API errors, especially missing
  permissions, missing OAuth scopes, rate limits, and not-found resources.
- Handle Discord rate limits dynamically from response headers; do not hard-code
  route limits.
- Treat Discord snowflakes as strings everywhere.
- Provide human-readable permission names while storing Discord-compatible
  bitfields internally.
- Maintain a test strategy covering unit tests, mocked API contract tests, and
  gated acceptance tests.

## Non-Goals

- Do not automate standard Discord user accounts or self-bot behavior.
- Do not perform interactive OAuth consent during `terraform plan` or
  `terraform apply`.
- Do not claim support for Developer Portal or Discord review workflows that are
  not exposed through public APIs.
- Do not model transient runtime APIs as ordinary Terraform resources, including:
  - Gateway connections
  - Gateway events
  - interaction responses
  - component/modal responses
  - voice connections
  - rich presence runtime state
  - activity runtime sessions

## Automation Boundary

The provider should automate every stable Discord configuration surface exposed
by public APIs.

Bootstrap, consent, and review flows that Discord does not expose through public
APIs must be documented clearly, but not represented as artificial provider
resources or data sources unless Discord exposes a real API object for them.

Examples of Discord-controlled steps that may remain outside Terraform:

- Initial Discord application creation, if no public API exists for it.
- Bot token creation or regeneration.
- User/admin OAuth consent.
- Bot/application installation authorization into a guild.
- Privileged intent approval.
- Partner-only or approval-only scopes/features.
- Discovery, verification, monetization, or policy review flows not exposed as
  public CRUD APIs.

The provider should minimize manual UI dependence after bootstrap by managing
all ongoing API-exposed configuration declaratively.

## Authentication Model

Provider configuration:

```hcl
provider "discord" {
  bot_token     = var.discord_bot_token
  bearer_token  = var.discord_bearer_token
  client_id     = var.discord_client_id
  client_secret = var.discord_client_secret

  api_base_url = "https://discord.com/api/v10"

  default_audit_log_reason = "Managed by Terraform"
}
```

Environment variables:

- `DISCORD_BOT_TOKEN`
- `DISCORD_BEARER_TOKEN`
- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_API_BASE_URL`
- `DISCORD_AUDIT_LOG_REASON`

Authentication rules:

- Bot token is the primary authentication method for guild/server management.
- Bearer token is optional and only required for APIs that Discord restricts to
  OAuth bearer authentication, such as application command permissions.
  Editing command permissions requires a Bearer token from the authorization-code
  flow, carrying the `applications.commands.permissions.update` scope, granted by
  a user who can manage the guild's commands. A bot token is rejected by this
  endpoint, and this scope is not obtainable via the client-credentials grant.
- Client credentials may be used only where Discord supports non-interactive
  client credential flows for required scopes. This grant yields only a narrow
  scope set (for team-owned applications, limited to `identify` and
  `applications.commands.update`), so it cannot back command-permission management.
- All credentials are sensitive.
- Provider configuration should validate only cheap, local consistency checks.
  API capability and permission failures should be surfaced by the relevant
  resource or data source with actionable diagnostics.

## Repository Layout

```text
terraform-provider-discord/
  main.go
  go.mod
  Makefile
  README.md
  ROADMAP.md
  CHANGELOG.md
  CONTRIBUTING.md
  SECURITY.md
  LICENSE

  internal/
    provider/
      provider.go
      factory.go
      config.go
      meta.go

    conns/
      client.go
      config.go
      auth.go
      rate_limiter.go
      retry.go
      errors.go

    service/
      application/
      guild/
      channel/
      role/
      automod/
      webhook/
      emoji/
      sticker/
      scheduledevent/
      stage/
      invite/
      soundboard/
      monetization/

    acctest/
      provider.go
      env.go
      sweepers.go
      checks.go

    fwtypes/
      snowflake.go
      bitfield.go
      image_data.go

    validators/
      snowflake.go
      discord_name.go
      permissions.go
      url.go
      locale.go

    expand/
    flatten/

  docs/
    index.md
    resources/
    data-sources/

  examples/
    provider/
    resources/
    data-sources/
    complete-guild/
    application-commands/
    automoderation/

  templates/
    resources/
    data-sources/

  .github/
    workflows/
      build.yml
      test.yml
      lint.yml
      docs.yml
      examples.yml
      acceptance.yml
      release.yml
```

## Implementation Framework

Use Go and the HashiCorp Terraform Plugin Framework.

Provider implementation must include:

- provider factory and version injection
- typed provider configuration model
- shared API client in provider metadata
- service/domain package boundaries
- schema validators and plan modifiers
- resource importers
- resource timeouts where useful
- generated documentation via Terraform plugin documentation tooling

## Discord Client Requirements

The API client is a core provider component and must be built before broad
resource implementation.

Required behavior:

- Use Discord API v10 by default (explicitly pinned; not relying on Discord's
  unspecified-version default).
- Send bot and bearer authentication correctly per endpoint.
- Support `X-Audit-Log-Reason` where Discord supports it.
- Parse rate-limit response headers:
  - `X-RateLimit-Limit`
  - `X-RateLimit-Remaining`
  - `X-RateLimit-Reset`
  - `X-RateLimit-Reset-After`
  - `X-RateLimit-Bucket`
  - `X-RateLimit-Global` (present only on 429 responses)
  - `X-RateLimit-Scope` (present only on 429 responses; values `user`, `global`,
    `shared`)
- Retry 429 responses using `Retry-After` or the response body `retry_after`
  field.
- Serialize requests by rate-limit bucket where necessary.
- Stop retrying on invalid credentials.
- Return structured errors with HTTP status, route, Discord error code, and
  response message.
- Map 404 reads to Terraform state removal for managed resources.
- Avoid retrying unsafe operations unless Discord response semantics make retry
  safe.

## Resource Coverage

### Tier 1: Essential Guild Management

- `discord_guild_settings`
- `discord_role`
- `discord_member_role`
- `discord_channel`
- `discord_channel_permission_overwrite`
- `discord_auto_moderation_rule`
- `discord_webhook`

### Tier 2: Application and Bot Configuration

- `discord_application_settings`
- `discord_application_command`
- `discord_guild_application_command`
- `discord_application_command_permissions`
- `discord_application_role_connection_metadata`

### Tier 3: Community Features

- `discord_guild_widget`
- `discord_guild_welcome_screen`
- `discord_guild_onboarding`
- `discord_guild_template`
- `discord_guild_scheduled_event`
- `discord_stage_instance`
- `discord_invite`

### Tier 4: Media and Customization

- `discord_guild_emoji`
- `discord_guild_sticker`
- `discord_soundboard_sound`

Emoji routes require special caution because Discord documents that emoji API
routes do not follow normal rate-limit conventions.

### Tier 5: Read-Only and Advanced Data Sources

- `discord_current_application`
- `discord_current_user`
- `discord_guild`
- `discord_guilds`
- `discord_guild_preview`
- `discord_channel`
- `discord_channels`
- `discord_role`
- `discord_roles`
- `discord_user`
- `discord_webhook`
- `discord_invite`
- `discord_audit_log`
- `discord_voice_regions`
- `discord_sku`
- `discord_entitlement`
- `discord_subscription`

Data sources should represent real Discord API objects or collections, not
provider-internal capability checks.

### Capability Caveats

Most resources support full CRUD, but several Discord APIs are read/limited-write
only. The following resources must not assume standard Create/Read/Update/Delete:

| Resource | Supported operations | Constraint |
|---|---|---|
| `discord_application_command_permissions` | GET, PUT (guild-scoped only) | PUT requires a Bearer token with `applications.commands.permissions.update`; bot token rejected. No global-command variant. |
| `discord_application_role_connection_metadata` | GET, PUT (replace-all) | PUT replaces the entire metadata array; model as one whole-object resource, not per-record. |
| `discord_guild_onboarding` | GET, PUT (full replace) | No partial PATCH and no DELETE; Update issues a full PUT of desired state. |
| `discord_invite` | create (via channel), GET, DELETE | No update endpoint; all mutable fields are `ForceNew`. |
| `discord_channel_permission_overwrite` | PUT, DELETE | No standalone GET; read by fetching the parent channel and extracting the overwrite. Affects drift detection. |
| `discord_guild_welcome_screen` | GET, PATCH | No create/delete; the object always exists on community guilds. |
| `discord_guild_widget` | GET, PATCH | No create/delete; settings object always exists. |

## Schema Standards

Every resource should provide:

- Create, read, update, and delete where the Discord API supports them.
- Import support.
- Clear `ForceNew` behavior for immutable Discord fields.
- Drift detection.
- Validation for documented Discord constraints.
- Diff suppression or normalization for Discord-normalized values.
- Plan modifiers for API-computed defaults where needed.
- Timeouts where Discord operations can be delayed or eventually consistent.
- Markdown descriptions suitable for generated Registry documentation.

Snowflakes must be strings:

```hcl
guild_id = "123456789012345678"
```

Permissions should be human-readable in configuration:

```hcl
permissions = [
  "VIEW_CHANNEL",
  "SEND_MESSAGES",
  "MANAGE_MESSAGES",
]
```

The provider should convert permission names to Discord bitfields internally.

## Diagnostics Standards

Diagnostics should be specific, actionable, and resource-local.

Example:

```text
Error: creating Discord channel

Discord API returned 403 Forbidden for POST /guilds/{guild_id}/channels.
The bot likely lacks MANAGE_CHANNELS in guild 123456789012345678.
```

Diagnostics should include:

- operation name
- Discord route or high-level endpoint description
- HTTP status
- Discord error code and message, when present
- likely missing permission or OAuth scope, when known
- affected guild/channel/application ID, when safe to show

Diagnostics must never expose tokens or secrets.

## Documentation Standards

Each resource and data source must document:

- Minimal example.
- Complete example where useful.
- Required Discord permissions.
- Required OAuth scopes, if any.
- Authentication type used: bot token, bearer token, or either.
- Import syntax for resources.
- Discord API caveats.
- Known drift behavior.
- Link to the relevant Discord documentation.

Example import format:

```bash
terraform import discord_role.moderator guild_id:role_id
```

## Testing Strategy

### Unit Tests

Unit tests should cover:

- permission bitfield conversion
- snowflake validation
- enum validation
- locale validation
- URL validation
- image data encoding
- expand/flatten round trips
- API error mapping
- rate-limit header parsing
- retry decision logic

### Mocked API Contract Tests

Use `httptest` to verify:

- endpoint paths
- HTTP methods
- auth headers
- audit-log reason headers
- request payloads
- response flattening
- 404 state removal
- 403 diagnostics
- 429 retry behavior

### Acceptance Tests

Acceptance tests are gated behind environment variables:

```text
TF_ACC=1
DISCORD_BOT_TOKEN=...
DISCORD_TEST_GUILD_ID=...
DISCORD_TEST_APPLICATION_ID=...
DISCORD_BEARER_TOKEN=...
```

Acceptance tests should be service-scoped:

```bash
make testacc TESTS=./internal/service/role
make testacc TESTS=./internal/service/channel
make testacc TESTS=./internal/service/automod
```

Acceptance tests must clean up after themselves. Sweepers should remove leftover
test roles, channels, webhooks, automod rules, commands, and other test objects
where safe.

## CI Standards

Required CI checks:

- `go test ./...`
- race-enabled tests where practical
- `go vet`
- `gofmt` check
- `golangci-lint`
- `govulncheck`
- generated documentation check
- example validation with `terraform validate`
- provider schema check
- changelog fragment check
- release build check

Acceptance tests should run manually, on a schedule, or in protected CI contexts
with Discord test credentials.

## Release Standards

- Use semantic versioning.
- Use signed release artifacts where possible.
- Generate changelog entries from fragments.
- Publish Terraform Registry-compatible documentation.
- Document breaking changes and state migrations clearly.
- Keep upgrade guides for major versions.

## Initial Milestones

### Milestone 1: Foundation

- Go module and provider entrypoint.
- Provider configuration.
- Discord API client.
- Dynamic rate-limit handling.
- Structured errors and diagnostics.
- Snowflake, bitfield, and image helper types.
- Documentation generation pipeline.
- CI for tests, linting, docs, and examples.

### Milestone 2: First Production Resources

- `discord_guild` data source.
- `discord_role` resource.
- `discord_channel` resource.
- Import support and acceptance tests for both resources.

### Milestone 3: Core Guild Management

- `discord_channel_permission_overwrite`.
- `discord_member_role`.
- `discord_guild_settings`.
- `discord_auto_moderation_rule`.
- `discord_webhook`.

### Milestone 4: Application Management

- `discord_current_application` data source.
- `discord_application_settings`.
- `discord_application_command`.
- `discord_guild_application_command`.
- `discord_application_role_connection_metadata`.

### Milestone 5: Broader Discord Coverage

- Community features.
- Scheduled events.
- Stage instances.
- Invites.
- Emojis, stickers, and soundboard sounds.
- Advanced read-only data sources.
