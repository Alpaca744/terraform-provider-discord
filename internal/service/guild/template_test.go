package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestTemplateLifecycleContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/1/templates":
			_, _ = w.Write([]byte(`{"code":"abc","name":"base","source_guild_id":"1","usage_count":0}`))
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/templates/abc":
			_, _ = w.Write([]byte(`{"code":"abc","name":"base","source_guild_id":"1","usage_count":3}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/templates/abc":
			_, _ = w.Write([]byte(`{"code":"abc","name":"renamed","source_guild_id":"1","usage_count":3}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/templates/abc":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":"abc"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()
	name := "base"

	tmpl, err := CreateTemplate(ctx, c, "1", templateWriteBody{Name: &name}, "")
	if err != nil || tmpl.Code != "abc" {
		t.Fatalf("create: %v %+v", err, tmpl)
	}
	got, err := GetTemplate(ctx, c, "abc")
	if err != nil || got.UsageCount != 3 {
		t.Fatalf("get: %v %+v", err, got)
	}
	renamed := "renamed"
	upd, err := ModifyTemplate(ctx, c, "1", "abc", templateWriteBody{Name: &renamed}, "")
	if err != nil || upd.Name != "renamed" {
		t.Fatalf("modify: %v %+v", err, upd)
	}
	if err := DeleteTemplate(ctx, c, "1", "abc", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
