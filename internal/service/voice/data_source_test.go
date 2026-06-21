package voice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestVoiceRegionsContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/voice/regions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"us-east","name":"US East","optimal":true,"deprecated":false,"custom":false}]`))
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	regions, err := getRegions(context.Background(), c)
	if err != nil || len(regions) != 1 {
		t.Fatalf("getRegions: %v %+v", err, regions)
	}
	if regions[0].ID != "us-east" || !regions[0].Optimal {
		t.Errorf("region = %+v", regions[0])
	}
}
