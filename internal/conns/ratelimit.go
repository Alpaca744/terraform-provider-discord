package conns

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimit captures the rate-limit state Discord reports on a response.
//
// Header semantics follow https://discord.com/developers/docs/topics/rate-limits.
// Limit/Remaining/Reset/ResetAfter/Bucket appear on normal rate-limited
// responses; Global and Scope are only present on 429 responses.
type RateLimit struct {
	Limit      int
	Remaining  int
	Reset      time.Time     // absolute reset time (X-RateLimit-Reset, unix seconds)
	ResetAfter time.Duration // X-RateLimit-Reset-After
	Bucket     string
	Global     bool   // X-RateLimit-Global (429 only)
	Scope      string // X-RateLimit-Scope: "user", "global", or "shared" (429 only)

	// HasData is true when at least one rate-limit header was present.
	HasData bool
}

// parseRateLimit extracts rate-limit headers from a response. Missing or
// malformed headers are skipped rather than erroring; the client treats absent
// data as "no constraint known".
func parseRateLimit(h http.Header) RateLimit {
	var rl RateLimit

	if v := h.Get("X-RateLimit-Limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.Limit = n
			rl.HasData = true
		}
	}
	if v := h.Get("X-RateLimit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.Remaining = n
			rl.HasData = true
		}
	}
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			sec := int64(f)
			nsec := int64((f - float64(sec)) * float64(time.Second))
			rl.Reset = time.Unix(sec, nsec)
			rl.HasData = true
		}
	}
	if v := h.Get("X-RateLimit-Reset-After"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			rl.ResetAfter = time.Duration(f * float64(time.Second))
			rl.HasData = true
		}
	}
	if v := h.Get("X-RateLimit-Bucket"); v != "" {
		rl.Bucket = v
		rl.HasData = true
	}
	if v := h.Get("X-RateLimit-Global"); v != "" {
		rl.Global = v == "true"
		rl.HasData = true
	}
	if v := h.Get("X-RateLimit-Scope"); v != "" {
		rl.Scope = v
		rl.HasData = true
	}

	return rl
}

// retryAfter computes how long to wait before retrying a 429, preferring the
// JSON body's retry_after (seconds, possibly fractional) when provided and
// falling back to the Retry-After header. A zero duration means "no hint".
func retryAfter(h http.Header, bodyRetryAfter float64) time.Duration {
	if bodyRetryAfter > 0 {
		return time.Duration(bodyRetryAfter * float64(time.Second))
	}
	if v := h.Get("Retry-After"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(f * float64(time.Second))
		}
	}
	return 0
}
