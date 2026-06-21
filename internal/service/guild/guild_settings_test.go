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

func TestModifyGuildContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/guilds/1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		// Only set fields should be present; afk_channel_id was nil so omitted.
		if got["name"] != "Renamed" {
			t.Errorf("name = %v", got["name"])
		}
		if _, present := got["afk_channel_id"]; present {
			t.Error("unset afk_channel_id should be omitted")
		}
		_, _ = w.Write([]byte(`{"id":"1","name":"Renamed","verification_level":2,"premium_progress_bar_enabled":true}`))
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	name := "Renamed"
	g, err := ModifyGuild(context.Background(), c, "1", GuildSettingsBody{Name: &name}, "")
	if err != nil {
		t.Fatalf("ModifyGuild: %v", err)
	}
	if g.Name != "Renamed" || g.VerificationLevel != 2 || !g.PremiumProgressBarEnabled {
		t.Errorf("guild = %+v", g)
	}
}

func TestGuildSettingsApplyNullsEmpty(t *testing.T) {
	r := &guildSettingsResource{}
	var m guildSettingsModel
	r.apply(&m, &Guild{ID: "1", Name: "G", Description: "", RulesChannelID: ""})
	if !m.Description.IsNull() || !m.RulesChannelID.IsNull() {
		t.Error("empty optional fields should map to null")
	}
	if m.Name.ValueString() != "G" {
		t.Errorf("name = %q", m.Name.ValueString())
	}
}

func TestGuildSettingsApplyCommunity(t *testing.T) {
	r := &guildSettingsResource{}
	var on guildSettingsModel
	r.apply(&on, &Guild{ID: "1", Name: "G", Features: []string{"NEWS", "COMMUNITY"}})
	if !on.Community.ValueBool() {
		t.Error("expected community true when COMMUNITY present")
	}
	var off guildSettingsModel
	r.apply(&off, &Guild{ID: "1", Name: "G", Features: []string{"NEWS"}})
	if off.Community.ValueBool() {
		t.Error("expected community false when COMMUNITY absent")
	}
}

func TestSetFeature(t *testing.T) {
	// Enabling preserves existing features and appends when missing.
	got := setFeature([]string{"NEWS", "VERIFIED"}, featureCommunity, true)
	if len(got) != 3 || got[2] != featureCommunity {
		t.Errorf("enable append = %v", got)
	}
	// Enabling when already present is a no-op on membership and order.
	got = setFeature([]string{"COMMUNITY", "NEWS"}, featureCommunity, true)
	if len(got) != 2 || got[0] != "COMMUNITY" || got[1] != "NEWS" {
		t.Errorf("enable existing = %v", got)
	}
	// Disabling removes only the named feature, preserving the rest.
	got = setFeature([]string{"NEWS", "COMMUNITY", "VERIFIED"}, featureCommunity, false)
	if len(got) != 2 || containsFeature(got, featureCommunity) {
		t.Errorf("disable = %v", got)
	}
}
