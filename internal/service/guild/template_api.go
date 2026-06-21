package guild

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Template mirrors the Discord guild template object. See
// https://discord.com/developers/docs/resources/guild-template.
type Template struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	UsageCount    int64  `json:"usage_count"`
	CreatorID     string `json:"creator_id"`
	SourceGuildID string `json:"source_guild_id"`
	IsDirty       *bool  `json:"is_dirty"`
}

type templateWriteBody struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateTemplate creates a template from a guild's current state.
func CreateTemplate(ctx context.Context, c *conns.Client, guildID string, body templateWriteBody, reason string) (*Template, error) {
	var out Template
	err := c.Do(ctx, "creating Discord guild template", http.MethodPost,
		fmt.Sprintf("/guilds/%s/templates", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTemplate reads a template by code.
func GetTemplate(ctx context.Context, c *conns.Client, code string) (*Template, error) {
	var out Template
	err := c.Do(ctx, "reading Discord guild template", http.MethodGet,
		fmt.Sprintf("/guilds/templates/%s", code), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyTemplate updates a template's name/description.
func ModifyTemplate(ctx context.Context, c *conns.Client, guildID, code string, body templateWriteBody, reason string) (*Template, error) {
	var out Template
	err := c.Do(ctx, "updating Discord guild template", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/templates/%s", guildID, code),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTemplate removes a template.
func DeleteTemplate(ctx context.Context, c *conns.Client, guildID, code, reason string) error {
	return c.Do(ctx, "deleting Discord guild template", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/templates/%s", guildID, code),
		conns.RequestOptions{AuditLogReason: reason})
}
