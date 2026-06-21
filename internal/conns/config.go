package conns

import (
	"fmt"
	"strings"
)

// DefaultAPIBaseURL is the Discord REST API base. v10 is pinned explicitly
// rather than relying on Discord's unspecified-version default.
const DefaultAPIBaseURL = "https://discord.com/api/v10"

// Config holds the resolved provider configuration used to build a Client.
type Config struct {
	BotToken       string
	BearerToken    string
	ClientID       string
	ClientSecret   string
	APIBaseURL     string
	AuditLogReason string
	// UserAgent is sent on every request; Discord requires a descriptive UA.
	UserAgent string
}

// Validate performs only cheap, local consistency checks. API capability and
// permission failures are intentionally left to the resource/data source that
// hits the relevant endpoint, so diagnostics are specific.
func (c *Config) Validate() error {
	if c.BotToken == "" && c.BearerToken == "" {
		return fmt.Errorf("at least one of bot_token or bearer_token must be configured")
	}
	if c.APIBaseURL != "" {
		if !strings.HasPrefix(c.APIBaseURL, "https://") && !strings.HasPrefix(c.APIBaseURL, "http://") {
			return fmt.Errorf("api_base_url must be an absolute http(s) URL, got %q", c.APIBaseURL)
		}
	}
	if (c.ClientID == "") != (c.ClientSecret == "") {
		return fmt.Errorf("client_id and client_secret must be set together")
	}
	return nil
}
