// Package channel implements the discord_channel resource: guild channels of
// the common manageable types (text, voice, category, announcement, forum).
package channel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Channel mirrors the subset of the Discord channel object the provider manages.
// See https://discord.com/developers/docs/resources/channel#channel-object.
type Channel struct {
	ID               string `json:"id"`
	Type             int64  `json:"type"`
	GuildID          string `json:"guild_id,omitempty"`
	Name             string `json:"name"`
	Topic            string `json:"topic,omitempty"`
	Position         int64  `json:"position"`
	NSFW             bool   `json:"nsfw"`
	ParentID         string `json:"parent_id,omitempty"`
	RateLimitPerUser int64  `json:"rate_limit_per_user,omitempty"`
	Bitrate          int64  `json:"bitrate,omitempty"`
	UserLimit        int64  `json:"user_limit,omitempty"`

	PermissionOverwrites []Overwrite `json:"permission_overwrites,omitempty"`
}

// Overwrite is a channel permission overwrite entry. type is 0 for a role and
// 1 for a member; allow and deny are decimal-string permission bitfields.
// See https://discord.com/developers/docs/resources/channel#overwrite-object.
type Overwrite struct {
	ID    string `json:"id"`
	Type  int64  `json:"type"`
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// overwriteBody is the PUT /channels/{id}/permissions/{overwrite.id} payload.
type overwriteBody struct {
	Type  int64  `json:"type"`
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// createBody is the POST /guilds/{id}/channels payload. type is required on
// create; the rest are optional.
type createBody struct {
	Name             string  `json:"name"`
	Type             int64   `json:"type"`
	Topic            *string `json:"topic,omitempty"`
	NSFW             *bool   `json:"nsfw,omitempty"`
	ParentID         *string `json:"parent_id,omitempty"`
	RateLimitPerUser *int64  `json:"rate_limit_per_user,omitempty"`
	Bitrate          *int64  `json:"bitrate,omitempty"`
	UserLimit        *int64  `json:"user_limit,omitempty"`
}

// modifyBody is the PATCH /channels/{id} payload. type is intentionally omitted;
// the resource treats type as ForceNew.
type modifyBody struct {
	Name             *string `json:"name,omitempty"`
	Topic            *string `json:"topic,omitempty"`
	NSFW             *bool   `json:"nsfw,omitempty"`
	ParentID         *string `json:"parent_id,omitempty"`
	RateLimitPerUser *int64  `json:"rate_limit_per_user,omitempty"`
	Bitrate          *int64  `json:"bitrate,omitempty"`
	UserLimit        *int64  `json:"user_limit,omitempty"`
}

func create(ctx context.Context, c *conns.Client, guildID string, body createBody, reason string) (*Channel, error) {
	var out Channel
	err := c.Do(ctx, "creating Discord channel", http.MethodPost,
		fmt.Sprintf("/guilds/%s/channels", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListChannels returns all channels in a guild.
func ListChannels(ctx context.Context, c *conns.Client, guildID string) ([]Channel, error) {
	var out []Channel
	err := c.Do(ctx, "listing Discord channels", http.MethodGet,
		fmt.Sprintf("/guilds/%s/channels", guildID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func get(ctx context.Context, c *conns.Client, channelID string) (*Channel, error) {
	var out Channel
	err := c.Do(ctx, "reading Discord channel", http.MethodGet,
		fmt.Sprintf("/channels/%s", channelID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, channelID string, body modifyBody, reason string) (*Channel, error) {
	var out Channel
	err := c.Do(ctx, "updating Discord channel", http.MethodPatch,
		fmt.Sprintf("/channels/%s", channelID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteChannel(ctx context.Context, c *conns.Client, channelID, reason string) error {
	return c.Do(ctx, "deleting Discord channel", http.MethodDelete,
		fmt.Sprintf("/channels/%s", channelID),
		conns.RequestOptions{AuditLogReason: reason})
}

// putOverwrite creates or replaces a channel permission overwrite.
func putOverwrite(ctx context.Context, c *conns.Client, channelID, overwriteID string, body overwriteBody, reason string) error {
	return c.Do(ctx, "setting Discord channel permission overwrite", http.MethodPut,
		fmt.Sprintf("/channels/%s/permissions/%s", channelID, overwriteID),
		conns.RequestOptions{AuditLogReason: reason, Body: body})
}

// deleteOverwrite removes a channel permission overwrite.
func deleteOverwrite(ctx context.Context, c *conns.Client, channelID, overwriteID, reason string) error {
	return c.Do(ctx, "deleting Discord channel permission overwrite", http.MethodDelete,
		fmt.Sprintf("/channels/%s/permissions/%s", channelID, overwriteID),
		conns.RequestOptions{AuditLogReason: reason})
}

// findOverwrite reads the parent channel and returns the overwrite with the
// given ID. Discord exposes no standalone GET for overwrites, so they are read
// through the channel object. The bool is false when no such overwrite exists.
func findOverwrite(ctx context.Context, c *conns.Client, channelID, overwriteID string) (*Overwrite, bool, error) {
	ch, err := get(ctx, c, channelID)
	if err != nil {
		return nil, false, err
	}
	for i := range ch.PermissionOverwrites {
		if ch.PermissionOverwrites[i].ID == overwriteID {
			return &ch.PermissionOverwrites[i], true, nil
		}
	}
	return nil, false, nil
}
