// Package conns implements the Discord REST API client used by every provider
// resource and data source: authentication, dynamic rate limiting, retry, and
// structured error mapping.
package conns

import (
	"fmt"
	"net/http"
)

// APIError is a structured Discord API error. It carries enough context for
// resource-local diagnostics (operation, route, HTTP status, Discord error code
// and message) without ever embedding credentials.
type APIError struct {
	// Operation is a human-readable description such as "creating Discord channel".
	Operation string
	// Method and Route describe the request, e.g. "POST" and
	// "/guilds/{guild_id}/channels". Route should use placeholders, not real IDs.
	Method string
	Route  string
	// Status is the HTTP status code Discord returned.
	Status int
	// Code is the Discord JSON error code (the "code" field), 0 if absent.
	Code int
	// Message is the Discord error message (the "message" field), if present.
	Message string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("Discord API returned %d %s for %s %s",
		e.Status, http.StatusText(e.Status), e.Method, e.Route)
	if e.Code != 0 {
		msg += fmt.Sprintf(" (code %d)", e.Code)
	}
	if e.Message != "" {
		msg += ": " + e.Message
	}
	return msg
}

// IsNotFound reports whether the error is a 404, which callers map to Terraform
// state removal for managed resources.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Status == http.StatusNotFound
}

// IsForbidden reports whether the error is a 403, typically a missing-permission
// condition worth a targeted diagnostic.
func IsForbidden(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Status == http.StatusForbidden
}

// IsUnauthorized reports whether the error is a 401, an invalid-credentials
// condition the client must not retry.
func IsUnauthorized(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Status == http.StatusUnauthorized
}
