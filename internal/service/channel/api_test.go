package channel

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

func TestCreateChannelContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/guilds/10/channels" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["name"] != "general" || got["type"] != float64(0) {
			t.Errorf("payload = %v", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"55","type":0,"name":"general","guild_id":"10","position":3}`))
	})
	defer srv.Close()

	ch, err := create(context.Background(), c, "10", createBody{Name: "general", Type: 0}, "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if ch.ID != "55" || ch.GuildID != "10" || ch.Position != 3 {
		t.Errorf("channel = %+v", ch)
	}
}

func TestGetModifyDeleteChannelContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/channels/55":
			_, _ = w.Write([]byte(`{"id":"55","type":0,"name":"general"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/channels/55":
			_, _ = w.Write([]byte(`{"id":"55","type":0,"name":"renamed"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/55":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"55"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	if _, err := get(ctx, c, "55"); err != nil {
		t.Fatalf("get: %v", err)
	}
	name := "renamed"
	got, err := modify(ctx, c, "55", modifyBody{Name: &name}, "")
	if err != nil || got.Name != "renamed" {
		t.Fatalf("modify: %v %+v", err, got)
	}
	if err := deleteChannel(ctx, c, "55", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestChannelApply(t *testing.T) {
	r := &channelResource{}
	m := &channelResourceModel{}
	r.apply(m, &Channel{ID: "55", Type: 2, Name: "voice", GuildID: "10", Position: 1, Topic: "hi"})
	if m.ID.ValueString() != "55" || m.Type.ValueInt64() != 2 {
		t.Errorf("id/type = %s/%d", m.ID.ValueString(), m.Type.ValueInt64())
	}
	if m.Topic.ValueString() != "hi" {
		t.Errorf("topic = %q", m.Topic.ValueString())
	}
	// ParentID was empty -> should be null, not "".
	if !m.ParentID.IsNull() {
		t.Errorf("empty parent_id should be null, got %q", m.ParentID.ValueString())
	}
}
