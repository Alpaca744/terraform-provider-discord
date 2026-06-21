package sticker

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
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

func TestStickerLifecycleContract(t *testing.T) {
	c, srv := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/guilds/1/stickers":
			_, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			mr := multipart.NewReader(r.Body, params["boundary"])
			var sawTags bool
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if part.FormName() == "tags" {
					b, _ := io.ReadAll(part)
					if string(b) != "wave" {
						t.Errorf("tags = %q", b)
					}
					sawTags = true
				}
			}
			if !sawTags {
				t.Error("tags field missing")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"8","guild_id":"1","name":"wave","tags":"wave","format_type":1}`))
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/1/stickers/8":
			_, _ = w.Write([]byte(`{"id":"8","guild_id":"1","name":"wave","tags":"wave","format_type":1}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/guilds/1/stickers/8":
			_, _ = w.Write([]byte(`{"id":"8","guild_id":"1","name":"renamed","tags":"wave","format_type":1}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/1/stickers/8":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	ctx := context.Background()

	st, err := create(ctx, c, "1", "wave", "a wave", "wave", "png", []byte("PNGDATA"), "")
	if err != nil || st.ID != "8" || st.FormatType != 1 {
		t.Fatalf("create: %v %+v", err, st)
	}
	if _, err := get(ctx, c, "1", "8"); err != nil {
		t.Fatalf("get: %v", err)
	}
	name := "renamed"
	upd, err := modify(ctx, c, "1", "8", modifyBody{Name: &name}, "")
	if err != nil || upd.Name != "renamed" {
		t.Fatalf("modify: %v %+v", err, upd)
	}
	if err := deleteSticker(ctx, c, "1", "8", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestStickerUnsupportedFormat(t *testing.T) {
	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: "https://example.com"})
	if _, err := create(context.Background(), c, "1", "n", "", "t", "webp", []byte("x"), ""); err == nil {
		t.Error("unsupported format should error")
	}
	if IsFormat("webp") || !IsFormat("lottie") {
		t.Error("IsFormat mapping wrong")
	}
}
