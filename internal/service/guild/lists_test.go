package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestListRolesAndPreviewContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/guilds/1/roles":
			_, _ = w.Write([]byte(`[{"id":"3","name":"mods","permissions":"3072","position":1}]`))
		case "/guilds/1/preview":
			_, _ = w.Write([]byte(`{"id":"1","name":"Cool","description":"d","features":["COMMUNITY"],"approximate_member_count":42}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	roles, err := ListRoles(ctx, c, "1")
	if err != nil || len(roles) != 1 || roles[0].Permissions != "3072" {
		t.Fatalf("ListRoles: %v %+v", err, roles)
	}
	p, err := GetGuildPreview(ctx, c, "1")
	if err != nil || p.Name != "Cool" || len(p.Features) != 1 || p.ApproximateMemberCount != 42 {
		t.Fatalf("GetGuildPreview: %v %+v", err, p)
	}
}
