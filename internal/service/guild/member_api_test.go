package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestMemberRoleContract(t *testing.T) {
	var put, del bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/members/2":
			_, _ = w.Write([]byte(`{"roles":["3","9"]}`))
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/1/members/2/roles/3":
			put = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/members/2/roles/3":
			del = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	if err := AddMemberRole(ctx, c, "1", "2", "3", ""); err != nil || !put {
		t.Fatalf("add: %v put=%v", err, put)
	}
	m, err := GetMember(ctx, c, "1", "2")
	if err != nil {
		t.Fatalf("get member: %v", err)
	}
	if !m.MemberHasRole("3") || m.MemberHasRole("99") {
		t.Errorf("roles = %v", m.Roles)
	}
	if err := RemoveMemberRole(ctx, c, "1", "2", "3", ""); err != nil || !del {
		t.Fatalf("remove: %v del=%v", err, del)
	}
}
