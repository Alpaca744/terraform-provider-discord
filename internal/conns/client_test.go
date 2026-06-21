package conns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{BotToken: "test-bot-token", APIBaseURL: baseURL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestClientSuccessAndHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bot test-bot-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("X-Audit-Log-Reason"); got != "managed by tf" {
			t.Errorf("audit reason = %q", got)
		}
		if r.Method != http.MethodPost || r.URL.Path != "/guilds/1/channels" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"42","name":"general"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var out struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := c.Do(context.Background(), "creating channel", http.MethodPost, "/guilds/1/channels", RequestOptions{
		AuditLogReason: "managed by tf",
		Body:           map[string]string{"name": "general"},
		Out:            &out,
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if out.ID != "42" || out.Name != "general" {
		t.Errorf("decoded = %+v", out)
	}
}

func TestClientBearerAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer brr" {
			t.Errorf("Authorization = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BotToken: "bot", BearerToken: "brr", APIBaseURL: srv.URL})
	if err := c.Do(context.Background(), "op", http.MethodPut, "/x", RequestOptions{Auth: AuthBearer}); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestClientNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":10003,"message":"Unknown Channel"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.Do(context.Background(), "reading channel", http.MethodGet, "/channels/9", RequestOptions{})
	if !IsNotFound(err) {
		t.Fatalf("expected not-found, got %v", err)
	}
	apiErr := err.(*APIError)
	if apiErr.Code != 10003 || apiErr.Message != "Unknown Channel" {
		t.Errorf("error fields = %+v", apiErr)
	}
}

func TestClientForbiddenNoRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":50013,"message":"Missing Permissions"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.Do(context.Background(), "creating channel", http.MethodPost, "/guilds/1/channels", RequestOptions{})
	if !IsForbidden(err) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("403 should not retry, got %d calls", got)
	}
}

func TestClient429Retry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.Header().Set("X-RateLimit-Scope", "user")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"You are being rate limited.","retry_after":0.001,"global":false}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.Do(context.Background(), "op", http.MethodGet, "/x", RequestOptions{})
	if err != nil {
		t.Fatalf("Do after retry: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", got)
	}
}

func TestClient500RetryThenFail(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.Do(context.Background(), "op", http.MethodGet, "/x", RequestOptions{})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// 1 initial + maxRetries retries.
	if got := atomic.LoadInt32(&calls); got != int32(maxRetries+1) {
		t.Errorf("expected %d calls, got %d", maxRetries+1, got)
	}
}

func TestErrorString(t *testing.T) {
	e := &APIError{Method: "POST", Route: "/guilds/{guild_id}/channels", Status: http.StatusForbidden, Code: 50013, Message: "Missing Permissions"}
	got := e.Error()
	want := "Discord API returned 403 Forbidden for POST /guilds/{guild_id}/channels (code 50013): Missing Permissions"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestConfigValidate(t *testing.T) {
	if err := (&Config{}).Validate(); err == nil {
		t.Error("missing tokens should error")
	}
	if err := (&Config{BotToken: "x"}).Validate(); err != nil {
		t.Errorf("bot token only should be valid: %v", err)
	}
	if err := (&Config{BotToken: "x", APIBaseURL: "ftp://nope"}).Validate(); err == nil {
		t.Error("bad scheme should error")
	}
	if err := (&Config{BotToken: "x", ClientID: "a"}).Validate(); err == nil {
		t.Error("client_id without secret should error")
	}
}

// ensure strconv import is used even if tests above change.
var _ = strconv.Itoa
