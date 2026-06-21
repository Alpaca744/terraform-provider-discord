package stage

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

func TestStageInstanceLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/stage-instances":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["channel_id"] != "10" || got["topic"] != "Q&A" {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"id":"1","guild_id":"5","channel_id":"10","topic":"Q&A","privacy_level":2}`))
		case r.Method == http.MethodGet && r.URL.Path == "/stage-instances/10":
			_, _ = w.Write([]byte(`{"id":"1","guild_id":"5","channel_id":"10","topic":"Q&A","privacy_level":2}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/stage-instances/10":
			_, _ = w.Write([]byte(`{"id":"1","guild_id":"5","channel_id":"10","topic":"Updated","privacy_level":2}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/stage-instances/10":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	si, err := create(ctx, c, createBody{ChannelID: "10", Topic: "Q&A"}, "")
	if err != nil || si.ID != "1" || si.GuildID != "5" {
		t.Fatalf("create: %v %+v", err, si)
	}
	if _, err := get(ctx, c, "10"); err != nil {
		t.Fatalf("get: %v", err)
	}
	topic := "Updated"
	upd, err := modify(ctx, c, "10", modifyBody{Topic: &topic}, "")
	if err != nil || upd.Topic != "Updated" {
		t.Fatalf("modify: %v %+v", err, upd)
	}
	if err := deleteInstance(ctx, c, "10", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
