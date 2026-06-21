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

func TestWidgetContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/widget":
			_, _ = w.Write([]byte(`{"enabled":true,"channel_id":"42"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/widget":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["enabled"] != true {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"enabled":true,"channel_id":"42"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	ws, err := GetWidgetSettings(ctx, c, "1")
	if err != nil || !ws.Enabled || ws.ChannelID == nil || *ws.ChannelID != "42" {
		t.Fatalf("get: %v %+v", err, ws)
	}
	enabled := true
	upd, err := ModifyWidgetSettings(ctx, c, "1", WidgetSettingsBody{Enabled: &enabled}, "")
	if err != nil || !upd.Enabled {
		t.Fatalf("modify: %v %+v", err, upd)
	}
}

func TestWidgetApply(t *testing.T) {
	r := &widgetResource{}
	m := &widgetModel{}
	ch := "42"
	r.apply(m, &WidgetSettings{Enabled: true, ChannelID: &ch})
	if !m.Enabled.ValueBool() || m.ChannelID.ValueString() != "42" {
		t.Errorf("model = %+v", m)
	}
	r.apply(m, &WidgetSettings{Enabled: false, ChannelID: nil})
	if !m.ChannelID.IsNull() {
		t.Error("nil channel_id should be null")
	}
}
