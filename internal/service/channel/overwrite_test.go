package channel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestOverwritePutAndFind(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/channels/10/permissions/3":
			body, _ := io.ReadAll(r.Body)
			var got map[string]any
			_ = json.Unmarshal(body, &got)
			if got["type"] != float64(0) || got["allow"] != "1024" {
				t.Errorf("payload = %v", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/channels/10":
			_, _ = w.Write([]byte(`{"id":"10","type":0,"name":"general","permission_overwrites":[{"id":"3","type":0,"allow":"1024","deny":"0"}]}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/10/permissions/3":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	if err := putOverwrite(ctx, c, "10", "3", overwriteBody{Type: 0, Allow: "1024", Deny: "0"}, ""); err != nil {
		t.Fatalf("put: %v", err)
	}
	ow, found, err := findOverwrite(ctx, c, "10", "3")
	if err != nil || !found {
		t.Fatalf("find: %v found=%v", err, found)
	}
	if ow.Allow != "1024" {
		t.Errorf("allow = %q", ow.Allow)
	}

	// A missing overwrite is reported as not-found, not an error.
	_, found, err = findOverwrite(ctx, c, "10", "999")
	if err != nil || found {
		t.Errorf("missing overwrite: err=%v found=%v", err, found)
	}
	if err := deleteOverwrite(ctx, c, "10", "3", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestOverwriteTypeMapping(t *testing.T) {
	if typeNameToInt("member") != overwriteTypeMember || typeNameToInt("role") != overwriteTypeRole {
		t.Error("name->int mapping wrong")
	}
	if typeIntToName(overwriteTypeMember) != "member" || typeIntToName(overwriteTypeRole) != "role" {
		t.Error("int->name mapping wrong")
	}
}
