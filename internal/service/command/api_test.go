package command

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

func TestGlobalCommandLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/applications/9/commands":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["name"] != "ping" {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"id":"100","application_id":"9","name":"ping","description":"pong","type":1,"version":"1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/applications/9/commands/100":
			_, _ = w.Write([]byte(`{"id":"100","application_id":"9","name":"ping","type":1,"version":"1"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/applications/9/commands/100":
			_, _ = w.Write([]byte(`{"id":"100","application_id":"9","name":"ping","description":"updated","type":1,"version":"2"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/applications/9/commands/100":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()
	base := GlobalBasePath("9")

	cmd, err := CreateCommand(ctx, c, base, WriteBody{Name: "ping", Description: "pong", Type: 1})
	if err != nil || cmd.ID != "100" {
		t.Fatalf("create: %v %+v", err, cmd)
	}
	if _, err := GetCommand(ctx, c, base, "100"); err != nil {
		t.Fatalf("get: %v", err)
	}
	upd, err := EditCommand(ctx, c, base, "100", WriteBody{Name: "ping", Description: "updated"})
	if err != nil || upd.Version != "2" {
		t.Fatalf("edit: %v %+v", err, upd)
	}
	if err := DeleteCommand(ctx, c, base, "100"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestGuildCommandPathAndOptions(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/9/guilds/5/commands" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		// Options must be passed through as a JSON array.
		if !containsJSONArray(body, "options") {
			t.Errorf("options not forwarded: %s", body)
		}
		_, _ = w.Write([]byte(`{"id":"77","application_id":"9","guild_id":"5","name":"cfg","type":1,"version":"1","options":[{"type":1,"name":"set","description":"set"}]}`))
	})
	defer srv.Close()

	base := GuildBasePath("9", "5")
	opts := json.RawMessage(`[{"type":1,"name":"set","description":"set"}]`)
	cmd, err := CreateCommand(context.Background(), c, base, WriteBody{Name: "cfg", Type: 1, Options: opts})
	if err != nil || cmd.GuildID != "5" {
		t.Fatalf("create: %v %+v", err, cmd)
	}
	if len(cmd.Options) == 0 {
		t.Error("expected options echoed back")
	}
}

func containsJSONArray(body []byte, key string) bool {
	var m map[string]json.RawMessage
	if json.Unmarshal(body, &m) != nil {
		return false
	}
	v, ok := m[key]
	return ok && len(v) > 0 && v[0] == '['
}

func TestBasePaths(t *testing.T) {
	if GlobalBasePath("9") != "/applications/9/commands" {
		t.Error("global base path wrong")
	}
	if GuildBasePath("9", "5") != "/applications/9/guilds/5/commands" {
		t.Error("guild base path wrong")
	}
}
