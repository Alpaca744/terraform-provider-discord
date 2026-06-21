package guild

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

func TestCreateRoleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/guilds/100/roles" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Audit-Log-Reason") != "tf" {
			t.Errorf("missing audit reason")
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["name"] != "mods" || got["permissions"] != "3072" {
			t.Errorf("payload = %v", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"7","name":"mods","permissions":"3072","position":2}`))
	})
	defer srv.Close()

	name := "mods"
	perms := "3072"
	role, err := CreateRole(context.Background(), c, "100", roleWriteBody{Name: &name, Permissions: &perms}, "tf")
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.ID != "7" || role.Position != 2 {
		t.Errorf("role = %+v", role)
	}
}

func TestGetModifyDeleteRoleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/roles/2":
			_, _ = w.Write([]byte(`{"id":"2","name":"r"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/roles/2":
			_, _ = w.Write([]byte(`{"id":"2","name":"renamed"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/roles/2":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	if _, err := GetRole(ctx, c, "1", "2"); err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	newName := "renamed"
	got, err := ModifyRole(ctx, c, "1", "2", roleWriteBody{Name: &newName}, "")
	if err != nil || got.Name != "renamed" {
		t.Fatalf("ModifyRole: %v %+v", err, got)
	}
	if err := DeleteRole(ctx, c, "1", "2", ""); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
}

func TestGetGuildContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/guilds/999" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"999","name":"Test","owner_id":"1","premium_tier":2}`))
	})
	defer srv.Close()

	g, err := GetGuild(context.Background(), c, "999")
	if err != nil {
		t.Fatalf("GetGuild: %v", err)
	}
	if g.Name != "Test" || g.PremiumTier != 2 {
		t.Errorf("guild = %+v", g)
	}
}
