package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveLoadAndPermissions(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "config.json")
	c := Defaults()
	c.APIKey = "secret"
	c.BrightnessPercent = 150
	if err := Save(p, c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.APIKey != "secret" || got.BrightnessPercent != 100 {
		t.Fatalf("unexpected config: %#v", got)
	}
	if runtime.GOOS != "windows" {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0600 {
			t.Fatalf("mode=%o", st.Mode().Perm())
		}
	}
}

func TestDefaultsEnableScheduledWake(t *testing.T) {
	if !Defaults().WakeForRefresh {
		t.Fatal("scheduled wake should be enabled by default")
	}
}

func TestRejectsBadURL(t *testing.T) {
	c := Defaults()
	c.BaseURL = "example.test"
	if err := c.Normalize(); err == nil {
		t.Fatal("expected invalid URL")
	}
}

func TestRejectsInsecureRemoteURL(t *testing.T) {
	c := Defaults()
	c.BaseURL = "http://192.168.1.20:3000"
	if err := c.Normalize(); err == nil {
		t.Fatal("expected remote HTTP URL to be rejected")
	}
	c.BaseURL = "http://127.0.0.1:9988"
	if err := c.Normalize(); err != nil {
		t.Fatalf("loopback mock URL should remain available: %v", err)
	}
}

func TestRejectsUnsafeServerURLFeatures(t *testing.T) {
	for _, raw := range []string{
		"https://user:secret@example.test",
		"https://example.test/trmnl",
		"https://example.test?token=secret",
		"https://example.test/#fragment",
	} {
		c := Defaults()
		c.BaseURL = raw
		if err := c.Normalize(); err == nil {
			t.Fatalf("Normalize accepted unsafe server URL %q", raw)
		}
	}
}

func TestValidateRemoteURLAllowsHTTPSImagesAndLoopbackMock(t *testing.T) {
	for _, raw := range []string{
		"https://cdn.example.test/screens/a.png?token=opaque",
		"http://127.0.0.1:9988/image/test.png",
		"http://[::1]:9988/image/test.png",
	} {
		if err := ValidateRemoteURL(raw); err != nil {
			t.Fatalf("ValidateRemoteURL(%q): %v", raw, err)
		}
	}
	for _, raw := range []string{
		"http://192.168.1.20/image.png",
		"ftp://example.test/image.png",
		"//example.test/image.png",
	} {
		if err := ValidateRemoteURL(raw); err == nil {
			t.Fatalf("ValidateRemoteURL accepted %q", raw)
		}
	}
}
