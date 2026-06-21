package emoji

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

func TestEmojiLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/1/emojis":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["name"] != "blob" || got["image"] != "data:image/png;base64,AAAA" {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"id":"9","name":"blob","animated":false,"roles":[]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/emojis/9":
			_, _ = w.Write([]byte(`{"id":"9","name":"blob","animated":false,"roles":["3"]}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/emojis/9":
			_, _ = w.Write([]byte(`{"id":"9","name":"renamed","animated":false,"roles":[]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/emojis/9":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	em, err := create(ctx, c, "1", createBody{Name: "blob", Image: "data:image/png;base64,AAAA"}, "")
	if err != nil || em.ID != "9" {
		t.Fatalf("create: %v %+v", err, em)
	}
	got, err := get(ctx, c, "1", "9")
	if err != nil || len(got.Roles) != 1 {
		t.Fatalf("get: %v %+v", err, got)
	}
	name := "renamed"
	upd, err := modify(ctx, c, "1", "9", modifyBody{Name: &name}, "")
	if err != nil || upd.Name != "renamed" {
		t.Fatalf("modify: %v %+v", err, upd)
	}
	if err := deleteEmoji(ctx, c, "1", "9", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
