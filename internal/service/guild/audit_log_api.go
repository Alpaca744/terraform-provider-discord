package guild

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// AuditLog mirrors the subset of the Discord audit log object exposed by the
// provider. See https://discord.com/developers/docs/resources/audit-log.
type AuditLog struct {
	Entries []AuditLogEntry `json:"audit_log_entries"`
}

// AuditLogEntry is one audit log entry.
type AuditLogEntry struct {
	ID         string  `json:"id"`
	UserID     *string `json:"user_id"`
	TargetID   *string `json:"target_id"`
	ActionType int64   `json:"action_type"`
	Reason     string  `json:"reason"`
}

// AuditLogQuery holds optional filters for an audit log request.
type AuditLogQuery struct {
	UserID     string
	ActionType *int64
	Limit      *int64
}

// GetAuditLog reads a guild's audit log with optional filters.
func GetAuditLog(ctx context.Context, c *conns.Client, guildID string, q AuditLogQuery) (*AuditLog, error) {
	values := url.Values{}
	if q.UserID != "" {
		values.Set("user_id", q.UserID)
	}
	if q.ActionType != nil {
		values.Set("action_type", strconv.FormatInt(*q.ActionType, 10))
	}
	if q.Limit != nil {
		values.Set("limit", strconv.FormatInt(*q.Limit, 10))
	}

	path := fmt.Sprintf("/guilds/%s/audit-logs", guildID)
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out AuditLog
	err := c.Do(ctx, "reading Discord audit log", http.MethodGet, path, conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
