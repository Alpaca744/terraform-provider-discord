// Package stage implements the discord_stage_instance resource. A stage
// instance is keyed by its stage channel ID rather than a separate instance ID.
package stage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// StageInstance mirrors the Discord stage instance object. See
// https://discord.com/developers/docs/resources/stage-instance.
type StageInstance struct {
	ID                   string `json:"id"`
	GuildID              string `json:"guild_id"`
	ChannelID            string `json:"channel_id"`
	Topic                string `json:"topic"`
	PrivacyLevel         int64  `json:"privacy_level"`
	DiscoverableDisabled bool   `json:"discoverable_disabled"`
}

type createBody struct {
	ChannelID             string `json:"channel_id"`
	Topic                 string `json:"topic"`
	PrivacyLevel          *int64 `json:"privacy_level,omitempty"`
	SendStartNotification *bool  `json:"send_start_notification,omitempty"`
}

type modifyBody struct {
	Topic        *string `json:"topic,omitempty"`
	PrivacyLevel *int64  `json:"privacy_level,omitempty"`
}

func create(ctx context.Context, c *conns.Client, body createBody, reason string) (*StageInstance, error) {
	var out StageInstance
	err := c.Do(ctx, "creating Discord stage instance", http.MethodPost,
		"/stage-instances", conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// get reads the stage instance for a stage channel. The path key is the channel
// ID, not the instance ID.
func get(ctx context.Context, c *conns.Client, channelID string) (*StageInstance, error) {
	var out StageInstance
	err := c.Do(ctx, "reading Discord stage instance", http.MethodGet,
		fmt.Sprintf("/stage-instances/%s", channelID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, channelID string, body modifyBody, reason string) (*StageInstance, error) {
	var out StageInstance
	err := c.Do(ctx, "updating Discord stage instance", http.MethodPatch,
		fmt.Sprintf("/stage-instances/%s", channelID), conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteInstance(ctx context.Context, c *conns.Client, channelID, reason string) error {
	return c.Do(ctx, "deleting Discord stage instance", http.MethodDelete,
		fmt.Sprintf("/stage-instances/%s", channelID), conns.RequestOptions{AuditLogReason: reason})
}
