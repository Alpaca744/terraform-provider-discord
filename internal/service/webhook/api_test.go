package webhook

import (
	"context"
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

func TestWebhookLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/channels/10/webhooks":
			_, _ = w.Write([]byte(`{"id":"77","channel_id":"10","name":"deploys","guild_id":"5","token":"abc"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/webhooks/77":
			_, _ = w.Write([]byte(`{"id":"77","channel_id":"10","name":"deploys"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/webhooks/77":
			_, _ = w.Write([]byte(`{"id":"77","channel_id":"20","name":"deploys"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/webhooks/77":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	wh, err := create(ctx, c, "10", createBody{Name: "deploys"}, "")
	if err != nil || wh.ID != "77" || wh.Token != "abc" {
		t.Fatalf("create: %v %+v", err, wh)
	}
	if _, err := get(ctx, c, "77"); err != nil {
		t.Fatalf("get: %v", err)
	}
	newChannel := "20"
	moved, err := modify(ctx, c, "77", modifyBody{ChannelID: &newChannel}, "")
	if err != nil || moved.ChannelID != "20" {
		t.Fatalf("modify: %v %+v", err, moved)
	}
	if err := deleteWebhook(ctx, c, "77", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestWebhookApplyURL(t *testing.T) {
	r := &webhookResource{}
	m := &webhookResourceModel{}
	r.apply(m, &Webhook{ID: "77", ChannelID: "10", Name: "x", Token: "abc"})
	want := "https://discord.com/api/webhooks/77/abc"
	if m.URL.ValueString() != want {
		t.Errorf("url = %q, want %q", m.URL.ValueString(), want)
	}
}
