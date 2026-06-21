package guild

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Member mirrors the subset of the Discord guild member object needed to verify
// role assignments. See
// https://discord.com/developers/docs/resources/guild#guild-member-object.
type Member struct {
	Roles []string `json:"roles"`
}

// GetMember fetches a guild member by user ID.
func GetMember(ctx context.Context, c *conns.Client, guildID, userID string) (*Member, error) {
	var out Member
	err := c.Do(ctx, "reading Discord guild member", http.MethodGet,
		fmt.Sprintf("/guilds/%s/members/%s", guildID, userID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AddMemberRole assigns a role to a member (idempotent).
func AddMemberRole(ctx context.Context, c *conns.Client, guildID, userID, roleID, reason string) error {
	return c.Do(ctx, "assigning Discord member role", http.MethodPut,
		fmt.Sprintf("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID),
		conns.RequestOptions{AuditLogReason: reason})
}

// RemoveMemberRole removes a role from a member (idempotent).
func RemoveMemberRole(ctx context.Context, c *conns.Client, guildID, userID, roleID, reason string) error {
	return c.Do(ctx, "removing Discord member role", http.MethodDelete,
		fmt.Sprintf("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID),
		conns.RequestOptions{AuditLogReason: reason})
}

// MemberHasRole reports whether the member currently holds roleID.
func (m *Member) MemberHasRole(roleID string) bool {
	for _, r := range m.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}
