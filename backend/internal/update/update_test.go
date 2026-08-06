package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewerOrdersSemanticVersions(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"2.1.0", "2.0.0", true},
		{"2.0.1", "2.0.0", true},
		{"3.0.0", "2.9.9", true},
		{"2.0.0", "2.0.0", false},
		{"2.0.0", "2.1.0", false},
		{"v2.1.0", "2.0.0", true},
		{"2.1", "2.0.0", true},
		// A development build must never advertise an update.
		{"2.1.0", "dev", false},
		{"garbage", "2.0.0", false},
		{"2.1.0-rc1", "2.0.0", true},
	}
	for _, c := range cases {
		if got := Newer(c.candidate, c.current); got != c.want {
			t.Fatalf("Newer(%q, %q) = %v, want %v", c.candidate, c.current, got, c.want)
		}
	}
}

func TestCheckReportsAvailableRelease(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v2.4.0","html_url":"https://example.test/releases/v2.4.0","draft":false}`))
	}))
	defer server.Close()

	c := New()
	c.Endpoint = server.URL
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	got, err := c.Check(context.Background(), "2.1.0", now)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !got.Available || got.Latest != "2.4.0" {
		t.Fatalf("result = %+v", got)
	}

	// A second call inside the interval must be served from cache.
	if _, err := c.Check(context.Background(), "2.1.0", now.Add(time.Hour)); err != nil {
		t.Fatalf("cached check: %v", err)
	}
	if requests != 1 {
		t.Fatalf("made %d requests, want 1 within the cache interval", requests)
	}

	if _, err := c.Check(context.Background(), "2.1.0", now.Add(25*time.Hour)); err != nil {
		t.Fatalf("expired check: %v", err)
	}
	if requests != 2 {
		t.Fatalf("made %d requests, want a refresh after the interval", requests)
	}
}

func TestCheckIgnoresDraftReleases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.0.0","draft":true}`))
	}))
	defer server.Close()
	c := New()
	c.Endpoint = server.URL
	got, err := c.Check(context.Background(), "2.1.0", time.Now())
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if got.Available {
		t.Fatal("a draft release was advertised as available")
	}
}

func TestCheckSurfacesServerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer server.Close()
	c := New()
	c.Endpoint = server.URL
	if _, err := c.Check(context.Background(), "2.1.0", time.Now()); err == nil {
		t.Fatal("an HTTP 403 was reported as success")
	}
}
