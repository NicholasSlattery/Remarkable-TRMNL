package trmnl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"trmnl-remarkable/backend/internal/config"
)

func TestDisplayHeadersAndFlexibleRate(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access-token") != "sekret" || r.Header.Get("ID") != "AA:BB" {
			t.Errorf("missing auth headers")
		}
		fmt.Fprint(w, `{"status":0,"image_url":"https://example.test/a.png","refresh_rate":"1800","special_function":"sleep"}`)
	}))
	defer s.Close()
	c := New()
	cfg := config.Defaults()
	cfg.BaseURL = s.URL
	cfg.APIKey = "sekret"
	cfg.DeviceID = "AA:BB"
	r, err := c.Display(context.Background(), cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if int(r.RefreshRate) != 1800 || r.SpecialFunction != "sleep" {
		t.Fatalf("unexpected response: %#v", r)
	}
}

func TestCurrentEndpointFallback(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/display/current" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"image_url":"https://example.test/a.bmp","refresh_rate":60}`)
	}))
	defer s.Close()
	c := New()
	cfg := config.Defaults()
	cfg.BaseURL = s.URL
	if _, err := c.Display(context.Background(), cfg, false); err != nil {
		t.Fatal(err)
	}
}
