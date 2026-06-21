package conns

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRateLimit(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "5")
	h.Set("X-RateLimit-Remaining", "0")
	h.Set("X-RateLimit-Reset", "1470173023.123")
	h.Set("X-RateLimit-Reset-After", "1.5")
	h.Set("X-RateLimit-Bucket", "abcd1234")
	h.Set("X-RateLimit-Global", "true")
	h.Set("X-RateLimit-Scope", "user")

	rl := parseRateLimit(h)
	if !rl.HasData {
		t.Fatal("expected HasData")
	}
	if rl.Limit != 5 || rl.Remaining != 0 {
		t.Errorf("limit/remaining = %d/%d", rl.Limit, rl.Remaining)
	}
	if rl.ResetAfter != 1500*time.Millisecond {
		t.Errorf("ResetAfter = %v", rl.ResetAfter)
	}
	if rl.Bucket != "abcd1234" {
		t.Errorf("bucket = %q", rl.Bucket)
	}
	if !rl.Global || rl.Scope != "user" {
		t.Errorf("global/scope = %v/%q", rl.Global, rl.Scope)
	}
}

func TestParseRateLimitEmpty(t *testing.T) {
	if parseRateLimit(http.Header{}).HasData {
		t.Error("empty headers should yield HasData=false")
	}
}

func TestRetryAfter(t *testing.T) {
	// Body retry_after takes precedence.
	h := http.Header{}
	h.Set("Retry-After", "2")
	if got := retryAfter(h, 0.5); got != 500*time.Millisecond {
		t.Errorf("body precedence: got %v", got)
	}
	// Falls back to header when body is absent.
	if got := retryAfter(h, 0); got != 2*time.Second {
		t.Errorf("header fallback: got %v", got)
	}
	// No hint at all.
	if got := retryAfter(http.Header{}, 0); got != 0 {
		t.Errorf("no hint: got %v", got)
	}
}
