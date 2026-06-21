package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Onboarding mirrors the Discord guild onboarding object. Prompts are carried as
// raw JSON because they nest deeply (prompts → options → channel/role IDs and
// emoji). See https://discord.com/developers/docs/resources/guild#guild-onboarding-object.
type Onboarding struct {
	GuildID           string          `json:"guild_id"`
	Prompts           json.RawMessage `json:"prompts"`
	DefaultChannelIDs []string        `json:"default_channel_ids"`
	Enabled           bool            `json:"enabled"`
	Mode              int64           `json:"mode"`
}

// OnboardingBody is the PUT /guilds/{id}/onboarding payload (replace-all).
type OnboardingBody struct {
	Prompts           json.RawMessage `json:"prompts"`
	DefaultChannelIDs []string        `json:"default_channel_ids"`
	Enabled           bool            `json:"enabled"`
	Mode              int64           `json:"mode"`
}

// GetOnboarding reads the onboarding configuration for a guild.
func GetOnboarding(ctx context.Context, c *conns.Client, guildID string) (*Onboarding, error) {
	var out Onboarding
	err := c.Do(ctx, "reading Discord guild onboarding", http.MethodGet,
		fmt.Sprintf("/guilds/%s/onboarding", guildID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PutOnboarding replaces the onboarding configuration and returns it.
func PutOnboarding(ctx context.Context, c *conns.Client, guildID string, body OnboardingBody, reason string) (*Onboarding, error) {
	var out Onboarding
	err := c.Do(ctx, "updating Discord guild onboarding", http.MethodPut,
		fmt.Sprintf("/guilds/%s/onboarding", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
