package monetization

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

func TestMonetizationContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/applications/9/skus":
			_, _ = w.Write([]byte(`[{"id":"1","type":5,"name":"Premium","slug":"premium","flags":0}]`))
		case "/applications/9/entitlements":
			if r.URL.Query().Get("user_id") != "42" {
				t.Errorf("entitlements user_id = %q", r.URL.Query().Get("user_id"))
			}
			_, _ = w.Write([]byte(`[{"id":"2","sku_id":"1","application_id":"9","user_id":"42","type":8,"deleted":false}]`))
		case "/skus/1/subscriptions":
			if r.URL.Query().Get("user_id") != "42" {
				t.Errorf("subscriptions user_id = %q", r.URL.Query().Get("user_id"))
			}
			_, _ = w.Write([]byte(`[{"id":"3","user_id":"42","sku_ids":["1"],"status":0}]`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	skus, err := ListSKUs(ctx, c, "9")
	if err != nil || len(skus) != 1 || skus[0].Name != "Premium" {
		t.Fatalf("ListSKUs: %v %+v", err, skus)
	}
	ents, err := ListEntitlements(ctx, c, "9", "42", "")
	if err != nil || len(ents) != 1 || ents[0].UserID == nil || *ents[0].UserID != "42" {
		t.Fatalf("ListEntitlements: %v %+v", err, ents)
	}
	subs, err := ListSubscriptions(ctx, c, "1", "42")
	if err != nil || len(subs) != 1 || len(subs[0].SKUIDs) != 1 {
		t.Fatalf("ListSubscriptions: %v %+v", err, subs)
	}
}
