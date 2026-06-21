// Package user implements the discord_current_user and discord_user data
// sources.
package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// User mirrors the subset of the Discord user object the provider exposes. See
// https://discord.com/developers/docs/resources/user.
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot"`
}

// GetCurrentUser returns the user behind the configured token.
func GetCurrentUser(ctx context.Context, c *conns.Client) (*User, error) {
	var out User
	err := c.Do(ctx, "reading current Discord user", http.MethodGet,
		"/users/@me", conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUser returns a user by ID.
func GetUser(ctx context.Context, c *conns.Client, userID string) (*User, error) {
	var out User
	err := c.Do(ctx, "reading Discord user", http.MethodGet,
		fmt.Sprintf("/users/%s", userID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}
