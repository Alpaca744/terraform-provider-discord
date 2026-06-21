// Package soundboard implements the discord_soundboard_sound resource.
package soundboard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Sound mirrors the Discord soundboard sound object. See
// https://discord.com/developers/docs/resources/soundboard.
type Sound struct {
	SoundID   string  `json:"sound_id"`
	Name      string  `json:"name"`
	Volume    float64 `json:"volume"`
	EmojiID   *string `json:"emoji_id"`
	EmojiName *string `json:"emoji_name"`
	GuildID   string  `json:"guild_id"`
	Available bool    `json:"available"`
}

// createBody carries the audio as a data URI; the sound is set only at creation.
type createBody struct {
	Name      string   `json:"name"`
	Sound     string   `json:"sound"`
	Volume    *float64 `json:"volume,omitempty"`
	EmojiID   *string  `json:"emoji_id,omitempty"`
	EmojiName *string  `json:"emoji_name,omitempty"`
}

type modifyBody struct {
	Name      *string  `json:"name,omitempty"`
	Volume    *float64 `json:"volume,omitempty"`
	EmojiID   *string  `json:"emoji_id"`
	EmojiName *string  `json:"emoji_name"`
}

func create(ctx context.Context, c *conns.Client, guildID string, body createBody, reason string) (*Sound, error) {
	var out Sound
	err := c.Do(ctx, "creating Discord soundboard sound", http.MethodPost,
		fmt.Sprintf("/guilds/%s/soundboard-sounds", guildID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func get(ctx context.Context, c *conns.Client, guildID, soundID string) (*Sound, error) {
	var out Sound
	err := c.Do(ctx, "reading Discord soundboard sound", http.MethodGet,
		fmt.Sprintf("/guilds/%s/soundboard-sounds/%s", guildID, soundID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func modify(ctx context.Context, c *conns.Client, guildID, soundID string, body modifyBody, reason string) (*Sound, error) {
	var out Sound
	err := c.Do(ctx, "updating Discord soundboard sound", http.MethodPatch,
		fmt.Sprintf("/guilds/%s/soundboard-sounds/%s", guildID, soundID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func deleteSound(ctx context.Context, c *conns.Client, guildID, soundID, reason string) error {
	return c.Do(ctx, "deleting Discord soundboard sound", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/soundboard-sounds/%s", guildID, soundID),
		conns.RequestOptions{AuditLogReason: reason})
}
