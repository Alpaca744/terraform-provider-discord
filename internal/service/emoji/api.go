// Package emoji implements the discord_guild_emoji resource.
package emoji

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Emoji mirrors the Discord emoji object. See
// https://discord.com/developers/docs/resources/emoji.
type Emoji struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Roles     []string `json:"roles"`
	Animated  bool     `json:"animated"`
	Available bool     `json:"available"`
	Managed   bool     `json:"managed"`
}

// createBody carries the image as a data URI; the image is set only at creation.
type createBody struct {
	Name  string   `json:"name"`
	Image string   `json:"image"`
	Roles []string `json:"roles,omitempty"`
}

type modifyBody struct {
	Name  *string  `json:"name,omitempty"`
	Roles []string `json:"roles"`
}

func create(ctx context.Context, c *conns.Client, guildID string, body createBody, reason string) (*Emoji, error) {
	var out Emoji
	err := c.Do(ctx, "creating Discord emoji", http.MethodPost,
		fmt.Sprintf("/guilds/%s/emojis", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func get(ctx context.Context, c *conns.Client, guildID, emojiID string) (*Emoji, error) {
	var out Emoji
	err := c.Do(ctx, "reading Discord emoji", http.MethodGet,
		fmt.Sprintf("/guilds/%s/emojis/%s", guildID, emojiID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, guildID, emojiID string, body modifyBody, reason string) (*Emoji, error) {
	var out Emoji
	err := c.Do(ctx, "updating Discord emoji", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/emojis/%s", guildID, emojiID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteEmoji(ctx context.Context, c *conns.Client, guildID, emojiID, reason string) error {
	return c.Do(ctx, "deleting Discord emoji", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/emojis/%s", guildID, emojiID),
		conns.RequestOptions{AuditLogReason: reason})
}
