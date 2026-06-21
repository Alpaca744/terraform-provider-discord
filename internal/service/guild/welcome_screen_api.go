package guild

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// WelcomeScreen mirrors the Discord welcome screen object. `enabled` is a
// modify-only parameter and is not returned by GET. See
// https://discord.com/developers/docs/resources/guild#welcome-screen-object.
type WelcomeScreen struct {
	Description     *string          `json:"description"`
	WelcomeChannels []WelcomeChannel `json:"welcome_channels"`
}

// WelcomeChannel is one channel shown on the welcome screen.
type WelcomeChannel struct {
	ChannelID   string  `json:"channel_id"`
	Description string  `json:"description"`
	EmojiID     *string `json:"emoji_id"`
	EmojiName   *string `json:"emoji_name"`
}

// WelcomeScreenBody is the PATCH /guilds/{id}/welcome-screen payload.
type WelcomeScreenBody struct {
	Enabled         *bool            `json:"enabled,omitempty"`
	Description     *string          `json:"description,omitempty"`
	WelcomeChannels []WelcomeChannel `json:"welcome_channels"`
}

// GetWelcomeScreen reads the welcome screen for a guild.
func GetWelcomeScreen(ctx context.Context, c *conns.Client, guildID string) (*WelcomeScreen, error) {
	var out WelcomeScreen
	err := c.Do(ctx, "reading Discord welcome screen", http.MethodGet,
		fmt.Sprintf("/guilds/%s/welcome-screen", guildID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyWelcomeScreen updates the welcome screen and returns it.
func ModifyWelcomeScreen(ctx context.Context, c *conns.Client, guildID string, body WelcomeScreenBody, reason string) (*WelcomeScreen, error) {
	var out WelcomeScreen
	err := c.Do(ctx, "updating Discord welcome screen", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/welcome-screen", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
