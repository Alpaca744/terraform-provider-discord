package conns

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoMultipart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot tkn" {
			t.Errorf("auth = %q", r.Header.Get("Authorization"))
		}
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "multipart/form-data" {
			t.Fatalf("content-type = %q (%v)", r.Header.Get("Content-Type"), err)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		var sawName, sawFile bool
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("part: %v", err)
			}
			switch part.FormName() {
			case "name":
				b, _ := io.ReadAll(part)
				if string(b) != "wave" {
					t.Errorf("name = %q", b)
				}
				sawName = true
			case "file":
				if part.FileName() != "sticker.png" {
					t.Errorf("filename = %q", part.FileName())
				}
				b, _ := io.ReadAll(part)
				if string(b) != "PNGDATA" {
					t.Errorf("file content = %q", b)
				}
				if ct := part.Header.Get("Content-Type"); ct != "image/png" {
					t.Errorf("part content-type = %q", ct)
				}
				sawFile = true
			}
		}
		if !sawName || !sawFile {
			t.Errorf("missing parts: name=%v file=%v", sawName, sawFile)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"99"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BotToken: "tkn", APIBaseURL: srv.URL})
	var out struct {
		ID string `json:"id"`
	}
	err := c.DoMultipart(context.Background(), "creating sticker", http.MethodPost, "/guilds/1/stickers",
		map[string]string{"name": "wave"},
		MultipartFile{FieldName: "file", FileName: "sticker.png", ContentType: "image/png", Content: []byte("PNGDATA")},
		RequestOptions{Out: &out})
	if err != nil {
		t.Fatalf("DoMultipart: %v", err)
	}
	if out.ID != "99" {
		t.Errorf("out.ID = %q", out.ID)
	}
}

func TestDoMultipartError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":50013,"message":"Missing Permissions"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BotToken: "tkn", APIBaseURL: srv.URL})
	err := c.DoMultipart(context.Background(), "op", http.MethodPost, "/x",
		map[string]string{"a": "b"}, MultipartFile{}, RequestOptions{})
	if !IsForbidden(err) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if !strings.Contains(err.Error(), "Missing Permissions") {
		t.Errorf("error = %v", err)
	}
}
