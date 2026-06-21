// Package sticker implements the discord_guild_sticker resource. Sticker
// creation is a multipart/form-data upload; edits and reads are JSON.
package sticker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Sticker mirrors the Discord sticker object. See
// https://discord.com/developers/docs/resources/sticker.
type Sticker struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	FormatType  int64  `json:"format_type"`
	Available   bool   `json:"available"`
	GuildID     string `json:"guild_id"`
}

// formatUpload maps a format name to the part Content-Type and filename Discord
// uses to infer the sticker format.
var formatUpload = map[string]struct {
	contentType string
	filename    string
}{
	"png":    {"image/png", "sticker.png"},
	"apng":   {"image/apng", "sticker.png"},
	"gif":    {"image/gif", "sticker.gif"},
	"lottie": {"application/json", "sticker.json"},
}

// IsFormat reports whether name is a supported sticker file format.
func IsFormat(name string) bool {
	_, ok := formatUpload[name]
	return ok
}

func create(ctx context.Context, c *conns.Client, guildID, name, description, tags, format string, content []byte, reason string) (*Sticker, error) {
	up, ok := formatUpload[format]
	if !ok {
		return nil, fmt.Errorf("unsupported sticker format %q", format)
	}
	var out Sticker
	err := c.DoMultipart(ctx, "creating Discord sticker", http.MethodPost,
		fmt.Sprintf("/guilds/%s/stickers", guildID),
		map[string]string{"name": name, "description": description, "tags": tags},
		conns.MultipartFile{FieldName: "file", FileName: up.filename, ContentType: up.contentType, Content: content},
		conns.RequestOptions{AuditLogReason: reason, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func get(ctx context.Context, c *conns.Client, guildID, stickerID string) (*Sticker, error) {
	var out Sticker
	err := c.Do(ctx, "reading Discord sticker", http.MethodGet,
		fmt.Sprintf("/guilds/%s/stickers/%s", guildID, stickerID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type modifyBody struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description"`
	Tags        *string `json:"tags,omitempty"`
}

func modify(ctx context.Context, c *conns.Client, guildID, stickerID string, body modifyBody, reason string) (*Sticker, error) {
	var out Sticker
	err := c.Do(ctx, "updating Discord sticker", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/stickers/%s", guildID, stickerID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteSticker(ctx context.Context, c *conns.Client, guildID, stickerID, reason string) error {
	return c.Do(ctx, "deleting Discord sticker", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/stickers/%s", guildID, stickerID),
		conns.RequestOptions{AuditLogReason: reason})
}
