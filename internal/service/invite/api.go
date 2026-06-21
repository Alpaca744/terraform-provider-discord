// Package invite implements the discord_invite resource. Invites are immutable
// after creation: there is no update endpoint, so every mutable field is
// ForceNew and Read only verifies the invite still exists.
package invite

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Invite mirrors the subset of the Discord invite object the provider uses.
// Note: GET /invites/{code} does not reliably return invite metadata such as
// max_age or max_uses; those are create-time inputs and treated as authoritative
// in state. See https://discord.com/developers/docs/resources/invite.
type Invite struct {
	Code    string       `json:"code"`
	Channel inviteEntity `json:"channel"`
	Guild   inviteEntity `json:"guild"`
}

type inviteEntity struct {
	ID string `json:"id"`
}

type createBody struct {
	MaxAge    int64 `json:"max_age"`
	MaxUses   int64 `json:"max_uses"`
	Temporary bool  `json:"temporary"`
	Unique    bool  `json:"unique"`
}

// createOnChannel creates an invite on a channel.
func createOnChannel(ctx context.Context, c *conns.Client, channelID string, body createBody, reason string) (*Invite, error) {
	var out Invite
	err := c.Do(ctx, "creating Discord invite", http.MethodPost,
		fmt.Sprintf("/channels/%s/invites", channelID),
		conns.RequestOptions{AuditLogReason: reason, Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// get reads an invite by code.
func get(ctx context.Context, c *conns.Client, code string) (*Invite, error) {
	var out Invite
	err := c.Do(ctx, "reading Discord invite", http.MethodGet,
		fmt.Sprintf("/invites/%s", code), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// deleteByCode revokes an invite.
func deleteByCode(ctx context.Context, c *conns.Client, code, reason string) error {
	return c.Do(ctx, "deleting Discord invite", http.MethodDelete,
		fmt.Sprintf("/invites/%s", code), conns.RequestOptions{AuditLogReason: reason})
}
