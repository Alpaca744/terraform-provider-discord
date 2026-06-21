// Package application implements application-scoped data sources and resources:
// the current application data source and (later) application settings and
// commands. All use the /applications/@me family of endpoints.
package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

// Application mirrors the subset of the Discord application object the provider
// reads or manages. See
// https://discord.com/developers/docs/resources/application#application-object.
type Application struct {
	ID                             string   `json:"id"`
	Name                           string   `json:"name"`
	Icon                           string   `json:"icon"`
	Description                    string   `json:"description"`
	BotPublic                      bool     `json:"bot_public"`
	BotRequireCodeGrant            bool     `json:"bot_require_code_grant"`
	TermsOfServiceURL              string   `json:"terms_of_service_url"`
	PrivacyPolicyURL               string   `json:"privacy_policy_url"`
	CustomInstallURL               string   `json:"custom_install_url"`
	RoleConnectionsVerificationURL string   `json:"role_connections_verification_url"`
	InteractionsEndpointURL        string   `json:"interactions_endpoint_url"`
	Flags                          int64    `json:"flags"`
	Tags                           []string `json:"tags"`
	ApproximateGuildCount          int64    `json:"approximate_guild_count"`
}

// GetCurrentApplication returns the application associated with the bot token.
func GetCurrentApplication(ctx context.Context, c *conns.Client) (*Application, error) {
	var out Application
	err := c.Do(ctx, "reading current Discord application", http.MethodGet,
		"/applications/@me", conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ApplicationSettingsBody is the PATCH /applications/@me payload. Pointer fields
// are sent only when set.
type ApplicationSettingsBody struct {
	Description                    *string  `json:"description,omitempty"`
	InteractionsEndpointURL        *string  `json:"interactions_endpoint_url,omitempty"`
	RoleConnectionsVerificationURL *string  `json:"role_connections_verification_url,omitempty"`
	CustomInstallURL               *string  `json:"custom_install_url,omitempty"`
	Tags                           []string `json:"tags,omitempty"`
}

// ModifyCurrentApplication patches the current application and returns it.
func ModifyCurrentApplication(ctx context.Context, c *conns.Client, body ApplicationSettingsBody) (*Application, error) {
	var out Application
	err := c.Do(ctx, "updating current Discord application", http.MethodPatch,
		"/applications/@me", conns.RequestOptions{Body: body, Out: &out})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// RoleConnectionMetadata is one application role connection metadata record.
// See https://discord.com/developers/docs/resources/application-role-connection-metadata.
type RoleConnectionMetadata struct {
	Type        int64  `json:"type"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetRoleConnectionMetadata returns all metadata records for an application.
func GetRoleConnectionMetadata(ctx context.Context, c *conns.Client, appID string) ([]RoleConnectionMetadata, error) {
	var out []RoleConnectionMetadata
	err := c.Do(ctx, "reading Discord application role connection metadata", http.MethodGet,
		fmt.Sprintf("/applications/%s/role-connections/metadata", appID),
		conns.RequestOptions{Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PutRoleConnectionMetadata replaces the full set of metadata records (the
// endpoint is replace-all; there is no per-record update). It returns the
// updated records.
func PutRoleConnectionMetadata(ctx context.Context, c *conns.Client, appID string, records []RoleConnectionMetadata) ([]RoleConnectionMetadata, error) {
	var out []RoleConnectionMetadata
	err := c.Do(ctx, "updating Discord application role connection metadata", http.MethodPut,
		fmt.Sprintf("/applications/%s/role-connections/metadata", appID),
		conns.RequestOptions{Body: records, Out: &out})
	if err != nil {
		return nil, err
	}
	return out, nil
}
