package soundboard

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

func TestSoundLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/1/soundboard-sounds":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["name"] != "airhorn" || got["volume"] != 0.5 {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write([]byte(`{"sound_id":"7","guild_id":"1","name":"airhorn","volume":0.5,"emoji_id":null,"emoji_name":null}`))
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/soundboard-sounds/7":
			_, _ = w.Write([]byte(`{"sound_id":"7","guild_id":"1","name":"airhorn","volume":0.5}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/soundboard-sounds/7":
			_, _ = w.Write([]byte(`{"sound_id":"7","guild_id":"1","name":"renamed","volume":1}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/soundboard-sounds/7":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	vol := 0.5
	s, err := create(ctx, c, "1", createBody{Name: "airhorn", Sound: "data:audio/mpeg;base64,AAAA", Volume: &vol}, "")
	if err != nil || s.SoundID != "7" {
		t.Fatalf("create: %v %+v", err, s)
	}
	if _, err := get(ctx, c, "1", "7"); err != nil {
		t.Fatalf("get: %v", err)
	}
	name := "renamed"
	upd, err := modify(ctx, c, "1", "7", modifyBody{Name: &name}, "")
	if err != nil || upd.Name != "renamed" {
		t.Fatalf("modify: %v %+v", err, upd)
	}
	if err := deleteSound(ctx, c, "1", "7", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestSoundApply(t *testing.T) {
	r := &soundResource{}
	m := &soundModel{}
	r.apply(m, &Sound{SoundID: "7", GuildID: "1", Name: "x", Volume: 0.5})
	if m.SoundID.ValueString() != "7" || m.Volume.ValueFloat64() != 0.5 {
		t.Errorf("model = %+v", m)
	}
	if !m.EmojiID.IsNull() {
		t.Error("nil emoji_id should map to null")
	}
}
