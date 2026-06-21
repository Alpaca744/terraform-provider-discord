package channel

import (
	"context"
	"net/http"
	"testing"
)

func TestListChannelsContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/guilds/1/channels" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"10","type":0,"name":"general","position":0},{"id":"11","type":2,"name":"Voice","position":1}]`))
	})
	defer srv.Close()

	channels, err := ListChannels(context.Background(), c, "1")
	if err != nil || len(channels) != 2 {
		t.Fatalf("ListChannels: %v %+v", err, channels)
	}
	if channels[0].Name != "general" || channels[1].Type != 2 {
		t.Errorf("channels = %+v", channels)
	}
}
