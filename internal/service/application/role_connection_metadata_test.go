package application

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestRoleConnectionMetadataContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/applications/42/role-connections/metadata":
			_, _ = w.Write([]byte(`[{"type":2,"key":"level","name":"Level","description":"Account level"}]`))
		case r.Method == http.MethodPut && r.URL.Path == "/applications/42/role-connections/metadata":
			body, _ := io.ReadAll(r.Body)
			var got []map[string]any
			_ = json.Unmarshal(body, &got)
			if len(got) != 1 || got[0]["key"] != "level" {
				t.Errorf("payload = %v", got)
			}
			_, _ = w.Write(body) // echo back
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	ctx := context.Background()

	got, err := GetRoleConnectionMetadata(ctx, c, "42")
	if err != nil || len(got) != 1 || got[0].Type != 2 {
		t.Fatalf("get: %v %+v", err, got)
	}

	put, err := PutRoleConnectionMetadata(ctx, c, "42", []RoleConnectionMetadata{
		{Type: 2, Key: "level", Name: "Level", Description: "Account level"},
	})
	if err != nil || len(put) != 1 || put[0].Key != "level" {
		t.Fatalf("put: %v %+v", err, put)
	}
}

func TestRecordRoundTrip(t *testing.T) {
	in := []recordModel{}
	api := expandRecords(in)
	if len(api) != 0 {
		t.Fatalf("empty expand = %v", api)
	}
	api = []RoleConnectionMetadata{{Type: 1, Key: "k", Name: "n", Description: "d"}}
	out := flattenRecords(api)
	if out[0].Type.ValueInt64() != 1 || out[0].Key.ValueString() != "k" {
		t.Errorf("flatten = %+v", out[0])
	}
}
