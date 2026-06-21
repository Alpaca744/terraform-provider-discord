package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestUserContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/@me":
			_, _ = w.Write([]byte(`{"id":"1","username":"bot","global_name":"Bot","discriminator":"0","bot":true}`))
		case "/users/42":
			_, _ = w.Write([]byte(`{"id":"42","username":"alice","discriminator":"0"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	me, err := GetCurrentUser(ctx, c)
	if err != nil || me.ID != "1" || !me.Bot {
		t.Fatalf("current user: %v %+v", err, me)
	}
	u, err := GetUser(ctx, c, "42")
	if err != nil || u.Username != "alice" {
		t.Fatalf("user: %v %+v", err, u)
	}
}

func TestApplyUserNullAvatar(t *testing.T) {
	var m userModel
	applyUser(&m, &User{ID: "1", Username: "x", Avatar: ""})
	if !m.Avatar.IsNull() {
		t.Error("empty avatar should be null")
	}
}
