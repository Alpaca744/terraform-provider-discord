# Contributing

Thanks for your interest in improving the Discord Terraform provider.

## Development setup

Requirements: Go (version declared in `go.mod`) and `make`.

```bash
make build      # compile the provider binary
make check      # gofmt check + go vet + unit tests
make testrace   # unit and contract tests with the race detector
make docs       # regenerate Terraform Registry documentation
make lint       # golangci-lint (if installed)
```

## Tests

### Unit and contract tests

These run without credentials and must pass on every change:

```bash
make test
```

Contract tests use `httptest.NewServer` to assert API endpoint paths, methods,
payloads, and error mapping — they do not require network access.

### Acceptance tests

Acceptance tests hit the live Discord API. They are gated behind `TF_ACC=1`:

```bash
export TF_ACC=1
export DISCORD_BOT_TOKEN="Bot <token>"
export DISCORD_TEST_GUILD_ID="<guild-snowflake>"
export DISCORD_TEST_APPLICATION_ID="<app-snowflake>"
make testacc
```

**Additional env vars for optional test suites:**

| Variable | Test suite | Requirement |
|---|---|---|
| `DISCORD_TEST_COMMUNITY_GUILD_ID` | `TestAccGuildWelcomeScreen`, `TestAccStageInstance_basic` | A guild with Community features enabled |
| `DISCORD_TEST_SOUND_DATA_URI` | `TestAccSoundboardSound_basic` | A valid `data:audio/mpeg;base64,...` or `data:audio/ogg;...` data URI. Discord validates audio server-side, so a real file is required. |

Tests that lack their required env var emit a `t.Skip` and are not counted as
failures.

Each test cleans up every Discord object it creates. Never point tests at a
production guild — use a dedicated test server owned by your bot.

## Adding a resource or data source

1. Add typed API call(s) under `internal/service/<area>/api.go`.
2. Implement the resource or data source with the Terraform Plugin Framework.
3. Register it in `internal/provider/provider.go`.
4. Add a contract test (`*_test.go`) and an acceptance test (`acc_test.go`).
5. Add an example under `examples/`.
6. Run `make docs` and commit the regenerated `docs/`.

## Conventions

- Treat Discord snowflakes as `types.String` everywhere (uint64 exceeds the
  JavaScript safe integer range).
- Surface permissions as human-readable names (e.g. `"KICK_MEMBERS"`); convert
  to bitfields only at the API boundary.
- Map HTTP 404 reads to `resp.State.RemoveResource` — never return an error for
  a missing resource on Read.
- Never retry on 401/403 (bad credentials, insufficient permissions).
- Diagnostics must be resource-local and must never surface tokens or secrets.
- Write-only attributes (image data URIs, audio data URIs, file content) must be
  marked `Sensitive: true` and excluded from import verification
  (`ImportStateVerifyIgnore`).
- Singleton resources (guild_settings, application_settings, …) implement Delete
  as a no-op.
- Replace-all PUT resources (role_connection_metadata, guild_onboarding) own
  their entire list.
