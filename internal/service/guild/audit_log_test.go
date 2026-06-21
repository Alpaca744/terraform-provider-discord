package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestGetAuditLogContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/1/audit-logs" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("action_type"); got != "10" {
			t.Errorf("action_type query = %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("limit query = %q", got)
		}
		_, _ = w.Write([]byte(`{"audit_log_entries":[{"id":"7","user_id":"2","target_id":"3","action_type":10,"reason":"cleanup"}]}`))
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	at := int64(10)
	limit := int64(5)
	log, err := GetAuditLog(context.Background(), c, "1", AuditLogQuery{ActionType: &at, Limit: &limit})
	if err != nil || len(log.Entries) != 1 {
		t.Fatalf("GetAuditLog: %v %+v", err, log)
	}
	e := log.Entries[0]
	if e.ID != "7" || e.UserID == nil || *e.UserID != "2" || e.Reason != "cleanup" {
		t.Errorf("entry = %+v", e)
	}
}
