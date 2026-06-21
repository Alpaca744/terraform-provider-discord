package scheduledevent

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

func TestScheduledEventLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/1/scheduled-events":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["name"] != "Launch" || got["entity_type"] != float64(3) {
				t.Errorf("payload = %v", got)
			}
			if md, ok := got["entity_metadata"].(map[string]any); !ok || md["location"] != "Online" {
				t.Errorf("entity_metadata = %v", got["entity_metadata"])
			}
			_, _ = w.Write([]byte(`{"id":"50","guild_id":"1","name":"Launch","entity_type":3,"privacy_level":2,"status":1,"scheduled_start_time":"2026-07-01T10:00:00Z","entity_metadata":{"location":"Online"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/scheduled-events/50":
			_, _ = w.Write([]byte(`{"id":"50","guild_id":"1","name":"Launch","entity_type":3,"privacy_level":2,"status":1,"scheduled_start_time":"2026-07-01T10:00:00Z","entity_metadata":{"location":"Online"}}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/scheduled-events/50":
			_, _ = w.Write([]byte(`{"id":"50","guild_id":"1","name":"Renamed","entity_type":3,"privacy_level":2,"status":4,"scheduled_start_time":"2026-07-01T10:00:00Z"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/scheduled-events/50":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	loc := "Online"
	ev, err := create(ctx, c, "1", createBody{
		Name: "Launch", EntityType: 3, PrivacyLevel: 2,
		ScheduledStartTime: "2026-07-01T10:00:00Z",
		EntityMetadata:     &EntityMetadata{Location: loc},
	}, "")
	if err != nil || ev.ID != "50" {
		t.Fatalf("create: %v %+v", err, ev)
	}
	if ev.EntityMetadata == nil || ev.EntityMetadata.Location != "Online" {
		t.Errorf("metadata = %+v", ev.EntityMetadata)
	}
	if _, err := get(ctx, c, "1", "50"); err != nil {
		t.Fatalf("get: %v", err)
	}
	name := "Renamed"
	status := int64(4)
	upd, err := modify(ctx, c, "1", "50", modifyBody{Name: &name, Status: &status}, "")
	if err != nil || upd.Status != 4 {
		t.Fatalf("modify: %v %+v", err, upd)
	}
	if err := deleteEvent(ctx, c, "1", "50", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestEventApplyLocation(t *testing.T) {
	r := &eventResource{}
	m := &eventModel{}
	r.apply(m, &Event{ID: "1", GuildID: "1", Name: "x", EntityType: 3, EntityMetadata: &EntityMetadata{Location: "Online"}})
	if m.Location.ValueString() != "Online" {
		t.Errorf("location = %q", m.Location.ValueString())
	}
	r.apply(m, &Event{ID: "1", GuildID: "1", Name: "x", EntityType: 2})
	if !m.Location.IsNull() {
		t.Error("missing location should be null")
	}
}
