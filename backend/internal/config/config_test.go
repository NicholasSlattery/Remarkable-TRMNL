package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
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

func TestQuietHoursWrapsPastMidnight(t *testing.T) {
	c := Defaults()
	c.QuietHoursEnabled = true
	c.QuietHoursStart, c.QuietHoursEnd = "23:00", "07:00"
	at := func(hour, minute int) time.Time {
		return time.Date(2026, 8, 5, hour, minute, 0, 0, time.UTC)
	}
	for _, quiet := range []time.Time{at(23, 0), at(23, 30), at(0, 15), at(6, 59)} {
		if !c.InQuietHours(quiet) {
			t.Fatalf("%s should be inside the quiet window", quiet.Format("15:04"))
		}
	}
	for _, active := range []time.Time{at(7, 0), at(12, 0), at(22, 59)} {
		if c.InQuietHours(active) {
			t.Fatalf("%s should be outside the quiet window", active.Format("15:04"))
		}
	}
	// A refresh landing at 01:00 is deferred to the 07:00 resume on the same day.
	if got := c.NextActiveTime(at(1, 0)); !got.Equal(at(7, 0)) {
		t.Fatalf("NextActiveTime(01:00) = %s, want 07:00", got.Format("15:04"))
	}
	// A refresh landing at 23:30 is deferred to 07:00 the following morning.
	want := at(7, 0).AddDate(0, 0, 1)
	if got := c.NextActiveTime(at(23, 30)); !got.Equal(want) {
		t.Fatalf("NextActiveTime(23:30) = %s, want %s", got, want)
	}
	// Times outside the window are untouched.
	if got := c.NextActiveTime(at(12, 0)); !got.Equal(at(12, 0)) {
		t.Fatalf("NextActiveTime(12:00) = %s, want it unchanged", got.Format("15:04"))
	}
}

func TestQuietHoursDisabledAndDaytimeWindow(t *testing.T) {
	c := Defaults()
	at := time.Date(2026, 8, 5, 23, 30, 0, 0, time.UTC)
	if c.InQuietHours(at) {
		t.Fatal("quiet hours applied while disabled")
	}
	c.QuietHoursEnabled = true
	c.QuietHoursStart, c.QuietHoursEnd = "09:00", "17:00"
	if !c.InQuietHours(time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)) {
		t.Fatal("a same-day window should contain noon")
	}
	if c.InQuietHours(at) {
		t.Fatal("a same-day window should not contain 23:30")
	}
}

func TestNormalizeValidatesNewSettings(t *testing.T) {
	c := Defaults()
	c.QuietHoursStart = "25:00"
	if err := c.Normalize(); err == nil {
		t.Fatal("an invalid quiet-hours time was accepted")
	}
	c = Defaults()
	c.DitherPalette = []string{"#00ff00", "not-a-colour"}
	if err := c.Normalize(); err == nil {
		t.Fatal("an invalid palette colour was accepted")
	}
	c = Defaults()
	c.Dither = "nonsense"
	c.BatterySaverPercent = 99
	if err := c.Normalize(); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if c.Dither != "off" || c.BatterySaverPercent != 0 {
		t.Fatalf("out-of-range values were not clamped: dither=%q saver=%d", c.Dither, c.BatterySaverPercent)
	}
	if len(c.Palette()) == 0 {
		t.Fatal("Palette() returned nothing for a default config")
	}
}
