package invite

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func testClient(t *testing.T, h http.HandlerFunc) (*conns.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := conns.NewClient(conns.Config{BotToken: "tkn", APIBaseURL: srv.URL})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	return c, srv
}

func TestInviteLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/channels/10/invites":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["max_age"] != float64(3600) || got["max_uses"] != float64(5) {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"code":"abc123","channel":{"id":"10"},"guild":{"id":"5"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/invites/abc123":
			_, _ = w.Write([]byte(`{"code":"abc123","channel":{"id":"10"},"guild":{"id":"5"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/invites/abc123":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":"abc123"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	inv, err := createOnChannel(ctx, c, "10", createBody{MaxAge: 3600, MaxUses: 5}, "")
	if err != nil || inv.Code != "abc123" || inv.Channel.ID != "10" {
		t.Fatalf("create: %v %+v", err, inv)
	}
	if _, err := get(ctx, c, "abc123"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := deleteByCode(ctx, c, "abc123", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestInviteApplyURL(t *testing.T) {
	r := &inviteResource{}
	m := &inviteModel{}
	r.apply(m, &Invite{Code: "xyz", Channel: inviteEntity{ID: "10"}, Guild: inviteEntity{ID: "5"}})
	if m.URL.ValueString() != "https://discord.gg/xyz" {
		t.Errorf("url = %q", m.URL.ValueString())
	}
	if m.ChannelID.ValueString() != "10" || m.GuildID.ValueString() != "5" {
		t.Errorf("model = %+v", m)
	}
}
