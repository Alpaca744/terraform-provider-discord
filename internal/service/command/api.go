// Package command implements the application command resources:
// discord_application_command (global) and discord_guild_application_command.
// Both wrap the same Discord command object; they differ only in endpoint scope
// and a couple of scope-specific fields.
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Command mirrors the Discord application command object. The options array is
// carried as raw JSON because Discord allows arbitrarily nested options
// (subcommand groups → subcommands → options) that do not map cleanly to a
// fixed-depth Terraform schema.
// See https://discord.com/developers/docs/interactions/application-commands.
type Command struct {
	ID                       string          `json:"id"`
	ApplicationID            string          `json:"application_id"`
	GuildID                  string          `json:"guild_id,omitempty"`
	Type                     int64           `json:"type,omitempty"`
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	DefaultMemberPermissions *string         `json:"default_member_permissions"`
	DMPermission             *bool           `json:"dm_permission,omitempty"`
	NSFW                     bool            `json:"nsfw"`
	Options                  json.RawMessage `json:"options,omitempty"`
	Version                  string          `json:"version"`
}

// WriteBody is the create/edit payload. Type is only honored on create.
type WriteBody struct {
	Name                     string          `json:"name"`
	Description              string          `json:"description,omitempty"`
	Type                     int64           `json:"type,omitempty"`
	DefaultMemberPermissions *string         `json:"default_member_permissions"`
	DMPermission             *bool           `json:"dm_permission,omitempty"`
	NSFW                     bool            `json:"nsfw"`
	Options                  json.RawMessage `json:"options,omitempty"`
}

// CreateCommand creates a command at the given base path
// ("/applications/{app}/commands" or the guild-scoped variant).
func CreateCommand(ctx context.Context, c *conns.Client, basePath string, body WriteBody) (*Command, error) {
	var out Command
	err := c.Do(ctx, "creating Discord application command", http.MethodPost,
		basePath, conns.RequestOptions{Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCommand reads a command by ID.
func GetCommand(ctx context.Context, c *conns.Client, basePath, id string) (*Command, error) {
	var out Command
	err := c.Do(ctx, "reading Discord application command", http.MethodGet,
		fmt.Sprintf("%s/%s", basePath, id), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// EditCommand updates a command by ID. Discord uses PATCH for edits.
func EditCommand(ctx context.Context, c *conns.Client, basePath, id string, body WriteBody) (*Command, error) {
	var out Command
	err := c.Do(ctx, "updating Discord application command", http.MethodPatch,
		fmt.Sprintf("%s/%s", basePath, id), conns.RequestOptions{Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCommand removes a command by ID.
func DeleteCommand(ctx context.Context, c *conns.Client, basePath, id string) error {
	return c.Do(ctx, "deleting Discord application command", http.MethodDelete,
		fmt.Sprintf("%s/%s", basePath, id), conns.RequestOptions{})
}

// GlobalBasePath returns the base path for global commands of an application.
func GlobalBasePath(appID string) string {
	return fmt.Sprintf("/applications/%s/commands", appID)
}

// GuildBasePath returns the base path for guild-scoped commands.
func GuildBasePath(appID, guildID string) string {
	return fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID)
}
