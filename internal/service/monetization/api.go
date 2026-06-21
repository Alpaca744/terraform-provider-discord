// Package monetization implements the discord_sku, discord_entitlement, and
// discord_subscription read-only data sources.
package monetization

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// SKU mirrors the Discord SKU object.
type SKU struct {
	ID            string `json:"id"`
	Type          int64  `json:"type"`
	ApplicationID string `json:"application_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Flags         int64  `json:"flags"`
}

// Entitlement mirrors the subset of the Discord entitlement object exposed.
type Entitlement struct {
	ID            string  `json:"id"`
	SKUID         string  `json:"sku_id"`
	ApplicationID string  `json:"application_id"`
	UserID        *string `json:"user_id"`
	GuildID       *string `json:"guild_id"`
	Type          int64   `json:"type"`
	Deleted       bool    `json:"deleted"`
	StartsAt      *string `json:"starts_at"`
	EndsAt        *string `json:"ends_at"`
}

// Subscription mirrors the subset of the Discord subscription object exposed.
type Subscription struct {
	ID                 string   `json:"id"`
	UserID             string   `json:"user_id"`
	SKUIDs             []string `json:"sku_ids"`
	Status             int64    `json:"status"`
	CurrentPeriodStart string   `json:"current_period_start"`
	CurrentPeriodEnd   string   `json:"current_period_end"`
}

// ListSKUs returns all SKUs for an application.
func ListSKUs(ctx context.Context, c *conns.Client, appID string) ([]SKU, error) {
	var out []SKU
	err := c.Do(ctx, "listing Discord SKUs", http.MethodGet,
		fmt.Sprintf("/applications/%s/skus", appID), conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListEntitlements returns entitlements for an application, optionally filtered
// by user or SKU.
func ListEntitlements(ctx context.Context, c *conns.Client, appID, userID, skuID string) ([]Entitlement, error) {
	values := url.Values{}
	if userID != "" {
		values.Set("user_id", userID)
	}
	if skuID != "" {
		values.Set("sku_ids", skuID)
	}
	path := fmt.Sprintf("/applications/%s/entitlements", appID)
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []Entitlement
	err := c.Do(ctx, "listing Discord entitlements", http.MethodGet, path, conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListSubscriptions returns subscriptions containing a SKU, filtered by user.
func ListSubscriptions(ctx context.Context, c *conns.Client, skuID, userID string) ([]Subscription, error) {
	values := url.Values{}
	if userID != "" {
		values.Set("user_id", userID)
	}
	path := fmt.Sprintf("/skus/%s/subscriptions", skuID)
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []Subscription
	err := c.Do(ctx, "listing Discord subscriptions", http.MethodGet, path, conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}
