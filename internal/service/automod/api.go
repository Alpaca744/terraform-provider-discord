// Package automod implements the discord_auto_moderation_rule resource.
package automod

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Rule mirrors the Discord auto moderation rule object. See
// https://discord.com/developers/docs/resources/auto-moderation.
type Rule struct {
	ID              string           `json:"id"`
	GuildID         string           `json:"guild_id"`
	Name            string           `json:"name"`
	CreatorID       string           `json:"creator_id"`
	EventType       int64            `json:"event_type"`
	TriggerType     int64            `json:"trigger_type"`
	TriggerMetadata *TriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []Action         `json:"actions"`
	Enabled         bool             `json:"enabled"`
	ExemptRoles     []string         `json:"exempt_roles"`
	ExemptChannels  []string         `json:"exempt_channels"`
}

// TriggerMetadata holds trigger-type-specific configuration.
type TriggerMetadata struct {
	KeywordFilter                []string `json:"keyword_filter,omitempty"`
	RegexPatterns                []string `json:"regex_patterns,omitempty"`
	Presets                      []int64  `json:"presets,omitempty"`
	AllowList                    []string `json:"allow_list,omitempty"`
	MentionTotalLimit            int64    `json:"mention_total_limit,omitempty"`
	MentionRaidProtectionEnabled bool     `json:"mention_raid_protection_enabled,omitempty"`
}

// Action is an automoderation action taken when the rule triggers.
type Action struct {
	Type     int64           `json:"type"`
	Metadata *ActionMetadata `json:"metadata,omitempty"`
}

// ActionMetadata holds action-type-specific configuration.
type ActionMetadata struct {
	ChannelID       string `json:"channel_id,omitempty"`
	DurationSeconds int64  `json:"duration_seconds,omitempty"`
	CustomMessage   string `json:"custom_message,omitempty"`
}

// createBody is the create payload. trigger_type is required and immutable.
type createBody struct {
	Name            string           `json:"name"`
	EventType       int64            `json:"event_type"`
	TriggerType     int64            `json:"trigger_type"`
	TriggerMetadata *TriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []Action         `json:"actions"`
	Enabled         bool             `json:"enabled"`
	ExemptRoles     []string         `json:"exempt_roles,omitempty"`
	ExemptChannels  []string         `json:"exempt_channels,omitempty"`
}

// modifyBody omits trigger_type, which Discord does not allow changing.
type modifyBody struct {
	Name            string           `json:"name"`
	EventType       int64            `json:"event_type"`
	TriggerMetadata *TriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         []Action         `json:"actions"`
	Enabled         bool             `json:"enabled"`
	ExemptRoles     []string         `json:"exempt_roles"`
	ExemptChannels  []string         `json:"exempt_channels"`
}

func create(ctx context.Context, c *conns.Client, guildID string, body createBody, reason string) (*Rule, error) {
	var out Rule
	err := c.Do(ctx, "creating Discord auto moderation rule", http.MethodPost,
		fmt.Sprintf("/guilds/%s/auto-moderation/rules", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func get(ctx context.Context, c *conns.Client, guildID, ruleID string) (*Rule, error) {
	var out Rule
	err := c.Do(ctx, "reading Discord auto moderation rule", http.MethodGet,
		fmt.Sprintf("/guilds/%s/auto-moderation/rules/%s", guildID, ruleID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, guildID, ruleID string, body modifyBody, reason string) (*Rule, error) {
	var out Rule
	err := c.Do(ctx, "updating Discord auto moderation rule", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/auto-moderation/rules/%s", guildID, ruleID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteRule(ctx context.Context, c *conns.Client, guildID, ruleID, reason string) error {
	return c.Do(ctx, "deleting Discord auto moderation rule", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/auto-moderation/rules/%s", guildID, ruleID),
		conns.RequestOptions{AuditLogReason: reason})
}
