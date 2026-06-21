package conns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthType selects which credential a request uses.
type AuthType int

const (
	// AuthBot uses the bot token ("Authorization: Bot <token>"). It is the
	// default for guild/server management.
	AuthBot AuthType = iota
	// AuthBearer uses an OAuth2 bearer token ("Authorization: Bearer <token>"),
	// required by endpoints Discord restricts to OAuth, such as editing
	// application command permissions.
	AuthBearer
)

// maxRetries bounds automatic retries for rate limits and transient 5xx errors.
const maxRetries = 4

// Client is the Discord REST API client. It serializes requests per rate-limit
// bucket, retries 429 and transient errors safely, and maps failures to
// structured APIErrors.
type Client struct {
	cfg        Config
	httpClient *http.Client

	// buckets serializes requests sharing a rate-limit bucket. Discord assigns
	// buckets dynamically via X-RateLimit-Bucket, so the map grows as routes are
	// observed. The mutex guards the map itself.
	mu      sync.Mutex
	buckets map[string]*sync.Mutex
}

// NewClient builds a Client from validated config.
func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
	}
	cfg.APIBaseURL = strings.TrimRight(cfg.APIBaseURL, "/")
	if cfg.UserAgent == "" {
		cfg.UserAgent = "terraform-provider-discord (https://github.com/alpaca744/terraform-provider-discord)"
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		buckets:    make(map[string]*sync.Mutex),
	}, nil
}

// RequestOptions configures a single API call.
type RequestOptions struct {
	Auth AuthType
	// AuditLogReason overrides the provider default for this call. Empty uses the
	// configured default (if any). Only sent on endpoints that audit.
	AuditLogReason string
	// Body is marshaled to JSON when non-nil.
	Body any
	// Out, when non-nil, receives the decoded JSON response body.
	Out any
}

// discordError mirrors the top-level shape of a Discord error response.
type discordError struct {
	Code       int     `json:"code"`
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

// Do executes method+path (path is relative to the API base, e.g.
// "/guilds/123/channels"). operation is a human description used in diagnostics
// and should not contain secrets.
func (c *Client) Do(ctx context.Context, operation, method, path string, opts RequestOptions) error {
	idempotent := isIdempotent(method)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		body, err := c.marshalBody(opts.Body)
		if err != nil {
			return &APIError{Operation: operation, Method: method, Route: path, Message: err.Error()}
		}

		req, err := http.NewRequestWithContext(ctx, method, c.cfg.APIBaseURL+path, body)
		if err != nil {
			return &APIError{Operation: operation, Method: method, Route: path, Message: err.Error()}
		}
		c.setHeaders(req, opts)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Network-level error: retry idempotent requests, else fail.
			lastErr = &APIError{Operation: operation, Method: method, Route: path, Message: err.Error()}
			if idempotent && attempt < maxRetries {
				continue
			}
			return lastErr
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if opts.Out != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, opts.Out); err != nil {
					return &APIError{Operation: operation, Method: method, Route: path,
						Status: resp.StatusCode, Message: fmt.Sprintf("decoding response: %v", err)}
				}
			}
			return nil
		}

		// Decode Discord's structured error body, if any.
		var de discordError
		_ = json.Unmarshal(respBody, &de)

		apiErr := &APIError{
			Operation: operation,
			Method:    method,
			Route:     path,
			Status:    resp.StatusCode,
			Code:      de.Code,
			Message:   de.Message,
		}
		lastErr = apiErr

		if !shouldRetry(resp.StatusCode, idempotent) || attempt == maxRetries {
			return apiErr
		}

		// Honor rate-limit timing on 429; small backoff otherwise.
		wait := retryAfter(resp.Header, de.RetryAfter)
		if wait <= 0 {
			wait = time.Duration(attempt+1) * 250 * time.Millisecond
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return lastErr
}

func (c *Client) marshalBody(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encoding request body: %w", err)
	}
	return bytes.NewReader(raw), nil
}

func (c *Client) setHeaders(req *http.Request, opts RequestOptions) {
	switch opts.Auth {
	case AuthBearer:
		req.Header.Set("Authorization", "Bearer "+c.cfg.BearerToken)
	default:
		req.Header.Set("Authorization", "Bot "+c.cfg.BotToken)
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	if opts.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	reason := opts.AuditLogReason
	if reason == "" {
		reason = c.cfg.AuditLogReason
	}
	if reason != "" {
		req.Header.Set("X-Audit-Log-Reason", reason)
	}
}
