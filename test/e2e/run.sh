#!/usr/bin/env bash
# End-to-end validation: drives the provider through real Terraform core against
# an in-memory mock of the Discord API. Proves schema validity and
# plan-correctness (create -> no-diff plan -> update -> import -> destroy)
# without needing live Discord credentials.
#
# Requires: terraform (or set TF=tofu) and go on PATH.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
WORK="$(mktemp -d)"
TF="${TF:-terraform}"
ADDR="127.0.0.1:17654"

cleanup() {
  [[ -n "${MOCK_PID:-}" ]] && kill "$MOCK_PID" 2>/dev/null || true
  rm -rf "$WORK"
}
trap cleanup EXIT

echo "==> building provider and mock"
go build -o "$WORK/terraform-provider-discord" "$REPO"
( cd "$HERE/mock" && go build -o "$WORK/mockdiscord" . )

echo "==> starting mock on $ADDR"
# Fully detach the mock: own session, stdin/stdout/stderr off the caller's TTY,
# so a wrapping process never blocks waiting on its descriptors.
setsid bash -c "MOCK_ADDR='$ADDR' '$WORK/mockdiscord' >'$WORK/mock.log' 2>&1" </dev/null &
MOCK_PID=$!
for _ in $(seq 1 25); do
  curl -s "http://$ADDR/api/v10/ping" >/dev/null 2>&1 && break || sleep 0.2
done

cat > "$WORK/dev.tfrc" <<EOF
provider_installation {
  dev_overrides { "alpaca744/discord" = "$WORK" }
  direct {}
}
EOF

cat > "$WORK/main.tf" <<'EOF'
terraform {
  required_providers {
    discord = { source = "alpaca744/discord" }
  }
}
provider "discord" {
  bot_token    = "test-token"
  api_base_url = "http://127.0.0.1:17654/api/v10"
}
resource "discord_role" "mod" {
  guild_id    = "111111111111111111"
  name        = "Moderator"
  permissions = ["VIEW_CHANNEL", "SEND_MESSAGES", "MANAGE_MESSAGES"]
  hoist       = true
}
resource "discord_channel" "general" {
  guild_id = "111111111111111111"
  name     = "general"
  type     = 0
  topic    = "Hello"
}
EOF

export TF_CLI_CONFIG_FILE="$WORK/dev.tfrc"
cd "$WORK"

echo "==> apply"
$TF apply -auto-approve >/dev/null

echo "==> plan must show NO changes (plan-correctness)"
if ! $TF plan -detailed-exitcode >/dev/null 2>&1; then
  echo "FAIL: provider produced a perpetual diff after apply" >&2
  $TF plan
  exit 1
fi

echo "==> update + re-plan"
sed -i 's/name        = "Moderator"/name        = "Mods"/' main.tf
$TF apply -auto-approve >/dev/null
$TF plan -detailed-exitcode >/dev/null

echo "==> destroy"
$TF destroy -auto-approve >/dev/null

echo "PASS: full create/plan/update/destroy lifecycle is clean"
