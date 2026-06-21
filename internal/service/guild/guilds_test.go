package guild

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestListCurrentUserGuildsContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/users/@me/guilds" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"1","name":"Alpha","owner":true},{"id":"2","name":"Beta","owner":false}]`))
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	guilds, err := ListCurrentUserGuilds(context.Background(), c)
	if err != nil || len(guilds) != 2 {
		t.Fatalf("ListCurrentUserGuilds: %v %+v", err, guilds)
	}
	if guilds[0].Name != "Alpha" || !guilds[0].Owner || guilds[1].Owner {
		t.Errorf("guilds = %+v", guilds)
	}
}
