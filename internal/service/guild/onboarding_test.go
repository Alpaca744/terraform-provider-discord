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

func TestOnboardingContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/onboarding":
			_, _ = w.Write([]byte(`{"guild_id":"1","enabled":true,"mode":0,"default_channel_ids":["10"],"prompts":[{"id":"5","title":"Pick"}]}`))
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/1/onboarding":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["enabled"] != true || got["mode"] != float64(1) {
				t.Errorf("payload = %v", got)
			}
			if _, ok := got["prompts"].([]any); !ok {
				t.Errorf("prompts not an array: %v", got["prompts"])
			}
			_, _ = w.Write([]byte(`{"guild_id":"1","enabled":true,"mode":1,"default_channel_ids":["10"],"prompts":[{"id":"5","title":"Pick"}]}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	ob, err := GetOnboarding(ctx, c, "1")
	if err != nil || !ob.Enabled || len(ob.DefaultChannelIDs) != 1 {
		t.Fatalf("get: %v %+v", err, ob)
	}
	upd, err := PutOnboarding(ctx, c, "1", OnboardingBody{
		Prompts:           json.RawMessage(`[{"id":"5","title":"Pick"}]`),
		DefaultChannelIDs: []string{"10"},
		Enabled:           true,
		Mode:              1,
	}, "")
	if err != nil || upd.Mode != 1 {
		t.Fatalf("put: %v %+v", err, upd)
	}
}

func TestOnboardingApply(t *testing.T) {
	r := &onboardingResource{}
	m := &onboardingModel{}
	d := r.apply(context.Background(), m, &Onboarding{Enabled: true, Mode: 1, Prompts: json.RawMessage(`[{"id":"5"}]`)})
	if d.HasError() {
		t.Fatalf("apply: %v", d)
	}
	if !m.Enabled.ValueBool() || m.Prompts.IsNull() {
		t.Errorf("model = %+v", m)
	}
	// Empty prompts array becomes null.
	_ = r.apply(context.Background(), m, &Onboarding{Prompts: json.RawMessage(`[]`)})
	if !m.Prompts.IsNull() {
		t.Error("empty prompts should be null")
	}
}
