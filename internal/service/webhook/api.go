// Package webhook implements the discord_webhook resource. Webhooks are created
// on a channel and can be moved between channels via modify.
package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Webhook mirrors the Discord webhook object fields the provider manages.
// See https://discord.com/developers/docs/resources/webhook#webhook-object.
type Webhook struct {
	ID        string `json:"id"`
	Type      int64  `json:"type"`
	GuildID   string `json:"guild_id,omitempty"`
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar,omitempty"`
	Token     string `json:"token,omitempty"`
	URL       string `json:"url,omitempty"`
}

type createBody struct {
	Name   string  `json:"name"`
	Avatar *string `json:"avatar,omitempty"`
}

// modifyBody patches name, avatar, and/or the channel the webhook posts to.
type modifyBody struct {
	Name      *string `json:"name,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
	ChannelID *string `json:"channel_id,omitempty"`
}

func create(ctx context.Context, c *conns.Client, channelID string, body createBody, reason string) (*Webhook, error) {
	var out Webhook
	err := c.Do(ctx, "creating Discord webhook", http.MethodPost,
		fmt.Sprintf("/channels/%s/webhooks", channelID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func get(ctx context.Context, c *conns.Client, webhookID string) (*Webhook, error) {
	var out Webhook
	err := c.Do(ctx, "reading Discord webhook", http.MethodGet,
		fmt.Sprintf("/webhooks/%s", webhookID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, webhookID string, body modifyBody, reason string) (*Webhook, error) {
	var out Webhook
	err := c.Do(ctx, "updating Discord webhook", http.MethodPatch,
		fmt.Sprintf("/webhooks/%s", webhookID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteWebhook(ctx context.Context, c *conns.Client, webhookID, reason string) error {
	return c.Do(ctx, "deleting Discord webhook", http.MethodDelete,
		fmt.Sprintf("/webhooks/%s", webhookID),
		conns.RequestOptions{AuditLogReason: reason})
}
