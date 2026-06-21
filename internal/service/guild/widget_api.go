package guild

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// WidgetSettings mirrors the Discord guild widget settings object
// (https://discord.com/developers/docs/resources/guild#guild-widget-settings-object).
type WidgetSettings struct {
	Enabled   bool    `json:"enabled"`
	ChannelID *string `json:"channel_id"`
}

// WidgetSettingsBody is the PATCH /guilds/{id}/widget payload.
type WidgetSettingsBody struct {
	Enabled   *bool   `json:"enabled,omitempty"`
	ChannelID *string `json:"channel_id,omitempty"`
}

// GetWidgetSettings reads the widget settings for a guild.
func GetWidgetSettings(ctx context.Context, c *conns.Client, guildID string) (*WidgetSettings, error) {
	var out WidgetSettings
	err := c.Do(ctx, "reading Discord guild widget", http.MethodGet,
		fmt.Sprintf("/guilds/%s/widget", guildID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyWidgetSettings updates the widget settings and returns them.
func ModifyWidgetSettings(ctx context.Context, c *conns.Client, guildID string, body WidgetSettingsBody, reason string) (*WidgetSettings, error) {
	var out WidgetSettings
	err := c.Do(ctx, "updating Discord guild widget", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/widget", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
