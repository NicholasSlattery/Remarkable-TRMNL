package trmnl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trmnl-remarkable/backend/internal/config"
)

func TestDisplayHeadersAndFlexibleRate(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("access-token") != "sekret" || r.Header.Get("ID") != "AA:BB" {
			t.Errorf("missing auth headers")
		}
		if r.Header.Get("battery-voltage") != "3.925" || r.Header.Get("firmware-version") != "1.1.0" {
			t.Errorf("incorrect device metadata headers: %#v", r.Header)
		}
		fmt.Fprint(w, `{"status":0,"image_url":"https://example.test/a.png","refresh_rate":"1800","special_function":"sleep"}`)
	}))
	defer s.Close()
	c := New()
	c.Version = "1.1.0"
	c.Battery = func() string { return "3.925" }
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

func TestDownloadUsesConditionalValidators(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != `"screen-v1"` || r.Header.Get("If-Modified-Since") == "" {
			t.Errorf("conditional validators missing: %#v", r.Header)
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer s.Close()
	c := New()
	_, _, err := c.Download(context.Background(), s.URL+"/screen.png", time.Second, `"screen-v1"`, time.Now().Add(-time.Hour).Format(http.TimeFormat))
	if !errors.Is(err, ErrNotModified) {
		t.Fatalf("expected ErrNotModified, got %v", err)
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

func TestClientRefusesCrossOriginRedirectWithoutLeakingToken(t *testing.T) {
	leaked := false
	destination := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = r.Header.Get("access-token") != ""
		w.WriteHeader(http.StatusOK)
	}))
	defer destination.Close()
	source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	defer source.Close()

	client := New()
	client.HTTP.Transport = source.Client().Transport
	cfg := config.Defaults()
	cfg.BaseURL = source.URL
	cfg.APIKey = "secret-token"
	_, err := client.Display(context.Background(), cfg, true)
	if err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("expected redirect refusal, got %v", err)
	}
	if leaked {
		t.Fatal("access token leaked to redirected origin")
	}
}

func TestDownloadRejectsRemoteHTTP(t *testing.T) {
	client := New()
	_, _, err := client.Download(context.Background(), "http://192.0.2.1/image.png", time.Second, "", "")
	if err == nil {
		t.Fatal("remote HTTP image URL was accepted")
	}
}

func TestDownloadAllowsHTTPSCDNRedirectWithoutCredentials(t *testing.T) {
	leaked := false
	destination := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = r.Header.Get("access-token") != ""
		fmt.Fprint(w, "image")
	}))
	defer destination.Close()
	source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL+"/image.png", http.StatusFound)
	}))
	defer source.Close()
	client := New()
	client.HTTP.Transport = source.Client().Transport
	body, _, err := client.Download(context.Background(), source.URL+"/redirect", time.Second, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	if leaked {
		t.Fatal("image redirect received an API credential")
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	if got := parseRetryAfter("300", now); got != 5*time.Minute {
		t.Fatalf("seconds retry-after = %v", got)
	}
	if got := parseRetryAfter(now.Add(10*time.Minute).Format(http.TimeFormat), now); got != 10*time.Minute {
		t.Fatalf("date retry-after = %v", got)
	}
}
