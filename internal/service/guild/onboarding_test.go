package guild

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	d := r.apply(context.Background(), m, &Onboarding{Enabled: true, Mode: 1, Prompts: json.RawMessage(`[{"id":"5","title":"Pick"}]`)})
	if d.HasError() {
		t.Fatalf("apply: %v", d)
	}
	if !m.Enabled.ValueBool() || m.Prompts.IsNull() {
		t.Errorf("model = %+v", m)
	}
	// A populated config cleared on the server is reflected as null (drift).
	_ = r.apply(context.Background(), m, &Onboarding{Prompts: json.RawMessage(`[]`)})
	if !m.Prompts.IsNull() {
		t.Error("server-cleared prompts should become null")
	}
}

// Repro A: a configured empty list must round-trip as "[]", not collapse to null.
func TestOnboardingApplyEmptyListPreserved(t *testing.T) {
	r := &onboardingResource{}
	m := &onboardingModel{Prompts: jsontypes.NewNormalizedValue("[]")}
	_ = r.apply(context.Background(), m, &Onboarding{Prompts: json.RawMessage(`[]`)})
	if m.Prompts.IsNull() || m.Prompts.ValueString() != "[]" {
		t.Errorf("configured [] should be preserved, got %q (null=%v)", m.Prompts.ValueString(), m.Prompts.IsNull())
	}
	// A null configuration stays null when the server has no prompts.
	mn := &onboardingModel{Prompts: jsontypes.NewNormalizedNull()}
	_ = r.apply(context.Background(), mn, &Onboarding{Prompts: json.RawMessage(`[]`)})
	if !mn.Prompts.IsNull() {
		t.Error("configured null should stay null")
	}
}

// Repro B: Discord reassigns ids, reorders keys, and enriches options; the
// configured prompts must remain authoritative (no inconsistency, no drift).
func TestOnboardingApplyPopulatedPreserved(t *testing.T) {
	r := &onboardingResource{}
	cfg := `[{"id":"0","type":0,"title":"Roles","options":[{"id":"0","title":"Dev"}]}]`
	m := &onboardingModel{Prompts: jsontypes.NewNormalizedValue(cfg)}
	api := json.RawMessage(`[{"id":"9876543210","title":"Roles","type":0,"options":[` +
		`{"id":"1234567890","title":"Dev","description":null,"emoji":{"id":null,"name":null,"animated":false},"role_ids":[],"channel_ids":[]}` +
		`]}]`)
	_ = r.apply(context.Background(), m, &Onboarding{Prompts: api})
	if m.Prompts.ValueString() != cfg {
		t.Errorf("configured prompts should be preserved, got %q", m.Prompts.ValueString())
	}

	// A genuine change (different title) is adopted from the API.
	m2 := &onboardingModel{Prompts: jsontypes.NewNormalizedValue(cfg)}
	changed := json.RawMessage(`[{"id":"9876543210","title":"Pick a role","type":0,"options":[{"id":"1","title":"Dev"}]}]`)
	_ = r.apply(context.Background(), m2, &Onboarding{Prompts: changed})
	if m2.Prompts.ValueString() != string(changed) {
		t.Errorf("genuine drift should adopt API value, got %q", m2.Prompts.ValueString())
	}
}

func TestJSONSubset(t *testing.T) {
	parse := func(s string) any {
		var v any
		_ = json.Unmarshal([]byte(s), &v)
		return v
	}
	// Superset with reassigned id and extra fields agrees.
	if !jsonSubset(parse(`{"type":0,"title":"x"}`), parse(`{"id":"9","type":0,"title":"x","emoji":null}`)) {
		t.Error("expected subset match ignoring id and extra keys")
	}
	// A differing scalar does not agree.
	if jsonSubset(parse(`{"title":"x"}`), parse(`{"title":"y"}`)) {
		t.Error("expected mismatch on differing title")
	}
	// Arrays of differing length do not agree.
	if jsonSubset(parse(`[1,2]`), parse(`[1]`)) {
		t.Error("expected mismatch on array length")
	}
}
