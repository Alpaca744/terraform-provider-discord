// Package acctest provides shared helpers for acceptance tests: the protocol-v6
// provider factory and environment-variable precondition checks. Acceptance
// tests are gated behind TF_ACC and require live Discord credentials.
package acctest

import (
	"os"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Environment variables that drive acceptance tests.
const (
	EnvBotToken             = "DISCORD_BOT_TOKEN"
	EnvBearerToken          = "DISCORD_BEARER_TOKEN"
	EnvTestGuildID          = "DISCORD_TEST_GUILD_ID"
	EnvTestCommunityGuildID = "DISCORD_TEST_COMMUNITY_GUILD_ID"
	EnvTestAppID            = "DISCORD_TEST_APPLICATION_ID"
)

// ProtoV6ProviderFactories builds the provider under test for resource.Test.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"discord": providerserver.NewProtocol6WithError(provider.New("acctest")()),
}

// PreCheck verifies the minimum credentials needed for guild-scoped acceptance
// tests are present. Call it from a test's PreCheck hook.
func PreCheck(t *testing.T) {
	t.Helper()
	requireEnv(t, EnvBotToken)
	requireEnv(t, EnvTestGuildID)
}

// PreCheckBearer additionally requires a bearer token, for tests that exercise
// OAuth-only endpoints such as application command permissions.
func PreCheckBearer(t *testing.T) {
	t.Helper()
	PreCheck(t)
	requireEnv(t, EnvBearerToken)
}

// TestGuildID returns the configured test guild ID.
func TestGuildID() string { return os.Getenv(EnvTestGuildID) }

// TestApplicationID returns the configured test application ID.
func TestApplicationID() string { return os.Getenv(EnvTestAppID) }

// TestCommunityGuildID returns a Community-enabled guild ID for tests that
// require Community features (welcome screen, etc.). Skips the test if the
// env var is not set.
func TestCommunityGuildID(t *testing.T) string {
	t.Helper()
	id := os.Getenv(EnvTestCommunityGuildID)
	if id == "" {
		t.Skipf("%s not set — skipping Community guild test", EnvTestCommunityGuildID)
	}
	return id
}

func requireEnv(t *testing.T, key string) {
	t.Helper()
	if os.Getenv(key) == "" {
		t.Fatalf("%s must be set for acceptance tests", key)
	}
}
