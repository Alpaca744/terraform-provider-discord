// Package scheduledevent implements the discord_guild_scheduled_event resource.
package scheduledevent

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Event mirrors the Discord guild scheduled event object. See
// https://discord.com/developers/docs/resources/guild-scheduled-event.
type Event struct {
	ID                 string          `json:"id"`
	GuildID            string          `json:"guild_id"`
	ChannelID          *string         `json:"channel_id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	ScheduledStartTime string          `json:"scheduled_start_time"`
	ScheduledEndTime   *string         `json:"scheduled_end_time"`
	PrivacyLevel       int64           `json:"privacy_level"`
	Status             int64           `json:"status"`
	EntityType         int64           `json:"entity_type"`
	EntityMetadata     *EntityMetadata `json:"entity_metadata"`
}

// EntityMetadata holds the location for EXTERNAL events.
type EntityMetadata struct {
	Location string `json:"location,omitempty"`
}

type createBody struct {
	ChannelID          *string         `json:"channel_id,omitempty"`
	Name               string          `json:"name"`
	Description        *string         `json:"description,omitempty"`
	ScheduledStartTime string          `json:"scheduled_start_time"`
	ScheduledEndTime   *string         `json:"scheduled_end_time,omitempty"`
	PrivacyLevel       int64           `json:"privacy_level"`
	EntityType         int64           `json:"entity_type"`
	EntityMetadata     *EntityMetadata `json:"entity_metadata,omitempty"`
	Image              *string         `json:"image,omitempty"`
}

type modifyBody struct {
	ChannelID          *string         `json:"channel_id,omitempty"`
	Name               *string         `json:"name,omitempty"`
	Description        *string         `json:"description,omitempty"`
	ScheduledStartTime *string         `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   *string         `json:"scheduled_end_time,omitempty"`
	PrivacyLevel       *int64          `json:"privacy_level,omitempty"`
	Status             *int64          `json:"status,omitempty"`
	EntityType         *int64          `json:"entity_type,omitempty"`
	EntityMetadata     *EntityMetadata `json:"entity_metadata,omitempty"`
	Image              *string         `json:"image,omitempty"`
}

func create(ctx context.Context, c *conns.Client, guildID string, body createBody, reason string) (*Event, error) {
	var out Event
	err := c.Do(ctx, "creating Discord scheduled event", http.MethodPost,
		fmt.Sprintf("/guilds/%s/scheduled-events", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func get(ctx context.Context, c *conns.Client, guildID, eventID string) (*Event, error) {
	var out Event
	err := c.Do(ctx, "reading Discord scheduled event", http.MethodGet,
		fmt.Sprintf("/guilds/%s/scheduled-events/%s", guildID, eventID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, guildID, eventID string, body modifyBody, reason string) (*Event, error) {
	var out Event
	err := c.Do(ctx, "updating Discord scheduled event", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/scheduled-events/%s", guildID, eventID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteEvent(ctx context.Context, c *conns.Client, guildID, eventID, reason string) error {
	return c.Do(ctx, "deleting Discord scheduled event", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/scheduled-events/%s", guildID, eventID),
		conns.RequestOptions{AuditLogReason: reason})
}
