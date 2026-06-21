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

func TestWelcomeScreenContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/welcome-screen":
			_, _ = w.Write([]byte(`{"description":"Welcome!","welcome_channels":[{"channel_id":"10","description":"Start here","emoji_name":"👋"}]}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/welcome-screen":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["enabled"] != true {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"description":"Welcome!","welcome_channels":[{"channel_id":"10","description":"Start here","emoji_name":"👋"}]}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	ws, err := GetWelcomeScreen(ctx, c, "1")
	if err != nil || ws.Description == nil || *ws.Description != "Welcome!" || len(ws.WelcomeChannels) != 1 {
		t.Fatalf("get: %v %+v", err, ws)
	}
	enabled := true
	upd, err := ModifyWelcomeScreen(ctx, c, "1", WelcomeScreenBody{Enabled: &enabled, WelcomeChannels: ws.WelcomeChannels}, "")
	if err != nil || len(upd.WelcomeChannels) != 1 {
		t.Fatalf("modify: %v %+v", err, upd)
	}
}

func TestWelcomeScreenApply(t *testing.T) {
	r := &welcomeScreenResource{}
	m := &welcomeScreenModel{}
	desc := "hi"
	r.apply(m, &WelcomeScreen{Description: &desc, WelcomeChannels: []WelcomeChannel{{ChannelID: "10", Description: "go"}}})
	if m.Description.ValueString() != "hi" || len(m.WelcomeChannels) != 1 {
		t.Errorf("model = %+v", m)
	}
	r.apply(m, &WelcomeScreen{Description: nil, WelcomeChannels: nil})
	if !m.Description.IsNull() || m.WelcomeChannels != nil {
		t.Error("empty screen should null description and clear channels")
	}
}
