// Package diagutil builds consistent, resource-local Terraform diagnostics from
// Discord API errors. It keeps diagnostics specific and actionable while never
// exposing tokens or secrets.
package diagutil

import (
	"fmt"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// APIError converts an error from the Discord client into a diagnostic.
// action is a verb phrase such as "creating Discord role"; subject is an
// optional affected ID (guild, channel, application) safe to display.
func APIError(action, subject string, err error) diag.Diagnostic {
	summary := action
	detail := err.Error()

	var apiErr *conns.APIError
	if ae, ok := err.(*conns.APIError); ok {
		apiErr = ae
	}

	if apiErr != nil {
		switch {
		case conns.IsForbidden(err):
			detail += "\n\nThis is usually a missing permission or missing OAuth scope. " +
				"Verify the bot has the required permission in the target guild."
		case conns.IsUnauthorized(err):
			detail += "\n\nThe credentials were rejected. Verify the configured bot or bearer token."
		}
	}
	if subject != "" {
		detail += fmt.Sprintf("\n\nAffected object: %s", subject)
	}

	return diag.NewErrorDiagnostic(summary, detail)
}
