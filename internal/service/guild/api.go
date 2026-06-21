// Package guild implements Discord guild-scoped resources and data sources:
// roles, and the guild data source. It wraps the shared conns.Client with typed
// request/response models for the relevant Discord endpoints.
package guild

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Role mirrors the Discord role object fields the provider manages.
// See https://discord.com/developers/docs/topics/permissions#role-object.
type Role struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Color        int64  `json:"color"`
	Hoist        bool   `json:"hoist"`
	Position     int64  `json:"position"`
	Permissions  string `json:"permissions"`
	Managed      bool   `json:"managed"`
	Mentionable  bool   `json:"mentionable"`
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
}

// roleWriteBody is the create/modify payload. Pointers distinguish "unset" from
// zero values so PATCH only sends intended changes.
type roleWriteBody struct {
	Name         *string `json:"name,omitempty"`
	Permissions  *string `json:"permissions,omitempty"`
	Color        *int64  `json:"color,omitempty"`
	Hoist        *bool   `json:"hoist,omitempty"`
	Mentionable  *bool   `json:"mentionable,omitempty"`
	UnicodeEmoji *string `json:"unicode_emoji,omitempty"`
}

// CreateRole creates a role in a guild and returns the created object.
func CreateRole(ctx context.Context, c *conns.Client, guildID string, body roleWriteBody, reason string) (*Role, error) {
	var out Role
	err := c.Do(ctx, "creating Discord role", http.MethodPost,
		fmt.Sprintf("/guilds/%s/roles", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRole fetches a single role by ID.
func GetRole(ctx context.Context, c *conns.Client, guildID, roleID string) (*Role, error) {
	var out Role
	err := c.Do(ctx, "reading Discord role", http.MethodGet,
		fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyRole updates a role and returns the updated object.
func ModifyRole(ctx context.Context, c *conns.Client, guildID, roleID string, body roleWriteBody, reason string) (*Role, error) {
	var out Role
	err := c.Do(ctx, "updating Discord role", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteRole removes a role from a guild.
func DeleteRole(ctx context.Context, c *conns.Client, guildID, roleID, reason string) error {
	return c.Do(ctx, "deleting Discord role", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID),
		conns.RequestOptions{AuditLogReason: reason})
}

// Guild mirrors the subset of the Discord guild object exposed by the data source.
// See https://discord.com/developers/docs/resources/guild#guild-object.
type Guild struct {
	ID                          string   `json:"id"`
	Name                        string   `json:"name"`
	Description                 string   `json:"description"`
	OwnerID                     string   `json:"owner_id"`
	Icon                        string   `json:"icon"`
	Splash                      string   `json:"splash"`
	Banner                      string   `json:"banner"`
	AFKChannelID                string   `json:"afk_channel_id"`
	AFKTimeout                  int64    `json:"afk_timeout"`
	VerificationLevel           int64    `json:"verification_level"`
	DefaultMessageNotifications int64    `json:"default_message_notifications"`
	ExplicitContentFilter       int64    `json:"explicit_content_filter"`
	PreferredLocale             string   `json:"preferred_locale"`
	PremiumTier                 int64    `json:"premium_tier"`
	PremiumSubscriptionCount    int64    `json:"premium_subscription_count"`
	SystemChannelID             string   `json:"system_channel_id"`
	RulesChannelID              string   `json:"rules_channel_id"`
	PublicUpdatesChannelID      string   `json:"public_updates_channel_id"`
	PremiumProgressBarEnabled   bool     `json:"premium_progress_bar_enabled"`
	Features                    []string `json:"features"`
}

// GuildSettingsBody is the PATCH /guilds/{id} payload. Pointer fields are only
// sent when set, so the modify endpoint receives just the intended changes.
type GuildSettingsBody struct {
	Name                        *string   `json:"name,omitempty"`
	Description                 *string   `json:"description,omitempty"`
	VerificationLevel           *int64    `json:"verification_level,omitempty"`
	DefaultMessageNotifications *int64    `json:"default_message_notifications,omitempty"`
	ExplicitContentFilter       *int64    `json:"explicit_content_filter,omitempty"`
	AFKChannelID                *string   `json:"afk_channel_id,omitempty"`
	AFKTimeout                  *int64    `json:"afk_timeout,omitempty"`
	SystemChannelID             *string   `json:"system_channel_id,omitempty"`
	RulesChannelID              *string   `json:"rules_channel_id,omitempty"`
	PublicUpdatesChannelID      *string   `json:"public_updates_channel_id,omitempty"`
	PreferredLocale             *string   `json:"preferred_locale,omitempty"`
	PremiumProgressBarEnabled   *bool     `json:"premium_progress_bar_enabled,omitempty"`
	Features                    *[]string `json:"features,omitempty"`
}

// GuildPreview mirrors the Discord guild preview object (available for
// discoverable guilds). See
// https://discord.com/developers/docs/resources/guild#guild-preview-object.
type GuildPreview struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	Icon                     string   `json:"icon"`
	Splash                   string   `json:"splash"`
	DiscoverySplash          string   `json:"discovery_splash"`
	Features                 []string `json:"features"`
	ApproximateMemberCount   int64    `json:"approximate_member_count"`
	ApproximatePresenceCount int64    `json:"approximate_presence_count"`
}

// GetGuildPreview fetches the preview of a discoverable guild.
func GetGuildPreview(ctx context.Context, c *conns.Client, guildID string) (*GuildPreview, error) {
	var out GuildPreview
	err := c.Do(ctx, "reading Discord guild preview", http.MethodGet,
		fmt.Sprintf("/guilds/%s/preview", guildID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PartialGuild mirrors an entry from GET /users/@me/guilds.
type PartialGuild struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Owner       bool     `json:"owner"`
	Permissions string   `json:"permissions"`
	Features    []string `json:"features"`
}

// ListCurrentUserGuilds returns the guilds the current user/bot is a member of.
func ListCurrentUserGuilds(ctx context.Context, c *conns.Client) ([]PartialGuild, error) {
	var out []PartialGuild
	err := c.Do(ctx, "listing current user's Discord guilds", http.MethodGet,
		"/users/@me/guilds", conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListRoles returns all roles in a guild.
func ListRoles(ctx context.Context, c *conns.Client, guildID string) ([]Role, error) {
	var out []Role
	err := c.Do(ctx, "listing Discord roles", http.MethodGet,
		fmt.Sprintf("/guilds/%s/roles", guildID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetGuild fetches a guild by ID.
func GetGuild(ctx context.Context, c *conns.Client, guildID string) (*Guild, error) {
	var out Guild
	err := c.Do(ctx, "reading Discord guild", http.MethodGet,
		fmt.Sprintf("/guilds/%s", guildID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyGuild applies guild settings via PATCH and returns the updated guild.
func ModifyGuild(ctx context.Context, c *conns.Client, guildID string, body GuildSettingsBody, reason string) (*Guild, error) {
	var out Guild
	err := c.Do(ctx, "updating Discord guild settings", http.MethodPatch,
		fmt.Sprintf("/guilds/%s", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
