package automod

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

func TestRuleLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/1/auto-moderation/rules":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["trigger_type"] != float64(1) || got["name"] != "block-bad" {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"id":"50","guild_id":"1","name":"block-bad","trigger_type":1,"event_type":1,"enabled":true,"actions":[{"type":1}],"trigger_metadata":{"keyword_filter":["bad"]}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/auto-moderation/rules/50":
			_, _ = w.Write([]byte(`{"id":"50","guild_id":"1","name":"block-bad","trigger_type":1,"event_type":1,"actions":[{"type":1}]}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/auto-moderation/rules/50":
			body, _ := io.ReadAll(r.Body)
			if _, hasTrigger := jsonField(body, "trigger_type"); hasTrigger {
				t.Error("modify body must not include trigger_type")
			}
			_, _ = w.Write([]byte(`{"id":"50","guild_id":"1","name":"renamed","trigger_type":1,"event_type":1,"actions":[{"type":1}]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/auto-moderation/rules/50":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	rule, err := create(ctx, c, "1", createBody{
		Name: "block-bad", EventType: 1, TriggerType: 1,
		Actions:         []Action{{Type: 1}},
		TriggerMetadata: &TriggerMetadata{KeywordFilter: []string{"bad"}},
		Enabled:         true,
	}, "")
	if err != nil || rule.ID != "50" {
		t.Fatalf("create: %v %+v", err, rule)
	}
	if _, err := get(ctx, c, "1", "50"); err != nil {
		t.Fatalf("get: %v", err)
	}
	got, err := modify(ctx, c, "1", "50", modifyBody{Name: "renamed", EventType: 1, Actions: []Action{{Type: 1}}}, "")
	if err != nil || got.Name != "renamed" {
		t.Fatalf("modify: %v %+v", err, got)
	}
	if err := deleteRule(ctx, c, "1", "50", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func jsonField(body []byte, key string) (any, bool) {
	var m map[string]any
	_ = json.Unmarshal(body, &m)
	v, ok := m[key]
	return v, ok
}

func TestRuleFlatten(t *testing.T) {
	r := &ruleResource{}
	m := &ruleModel{}
	rule := &Rule{
		ID: "50", GuildID: "1", Name: "n", EventType: 1, TriggerType: 1, Enabled: true,
		ExemptRoles:     []string{"9"},
		Actions:         []Action{{Type: 2, Metadata: &ActionMetadata{ChannelID: "7"}}},
		TriggerMetadata: &TriggerMetadata{KeywordFilter: []string{"x"}, MentionTotalLimit: 5},
	}
	if d := r.flatten(context.Background(), m, rule); d.HasError() {
		t.Fatalf("flatten: %v", d)
	}
	if m.ID.ValueString() != "50" || len(m.Actions) != 1 {
		t.Fatalf("model = %+v", m)
	}
	if m.Actions[0].ChannelID.ValueString() != "7" || m.Actions[0].Type.ValueInt64() != 2 {
		t.Errorf("action = %+v", m.Actions[0])
	}
	if m.TriggerMetadata == nil || m.TriggerMetadata.MentionTotalLimit.ValueInt64() != 5 {
		t.Errorf("trigger_metadata = %+v", m.TriggerMetadata)
	}
	// Action without timeout duration must stay null, not zero.
	if !m.Actions[0].DurationSeconds.IsNull() {
		t.Errorf("duration should be null, got %v", m.Actions[0].DurationSeconds)
	}
}
