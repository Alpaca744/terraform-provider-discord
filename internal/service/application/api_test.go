package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alpaca744/terraform-provider-discord/internal/conns"
)

func TestGetCurrentApplicationContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/applications/@me" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"42","name":"My App","description":"d","tags":["util"],"flags":565248,"bot_public":true}`))
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	app, err := GetCurrentApplication(context.Background(), c)
	if err != nil {
		t.Fatalf("GetCurrentApplication: %v", err)
	}
	if app.ID != "42" || app.Name != "My App" || !app.BotPublic {
		t.Errorf("app = %+v", app)
	}
	if len(app.Tags) != 1 || app.Tags[0] != "util" {
		t.Errorf("tags = %v", app.Tags)
	}
}

func TestModifyCurrentApplicationContract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/applications/@me" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"42","name":"My App","description":"updated"}`))
	}))
	defer srv.Close()

	c, _ := conns.NewClient(conns.Config{BotToken: "t", APIBaseURL: srv.URL})
	desc := "updated"
	app, err := ModifyCurrentApplication(context.Background(), c, ApplicationSettingsBody{Description: &desc})
	if err != nil || app.Description != "updated" {
		t.Fatalf("modify: %v %+v", err, app)
	}
}
