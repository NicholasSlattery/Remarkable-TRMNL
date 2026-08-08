package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"time"
	"trmnl-remarkable/backend/internal/config"
)

func TestReadBatteryPercentPrefersSystemBattery(t *testing.T) {
	root := t.TempDir()
	writePowerSupply(t, root, "elants-marker-battery", "Wireless", "0")
	writePowerSupply(t, root, "max1726x_battery", "Battery", "71")
	if got := readBatteryPercentAt(root); got != 71 {
		t.Fatalf("readBatteryPercentAt() = %d, want 71", got)
	}
}

func TestReadBatteryVoltageUsesSystemBatteryMicrovolts(t *testing.T) {
	root := t.TempDir()
	writePowerSupply(t, root, "max1726x_battery", "Battery", "71")
	path := filepath.Join(root, "max1726x_battery", "voltage_now")
	if err := os.WriteFile(path, []byte("3925000\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := readBatteryVoltageAt(root); got != "3.925" {
		t.Fatalf("readBatteryVoltageAt() = %q, want 3.925", got)
	}
}

func TestValidateImagePayload(t *testing.T) {
	var encoded bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 1620, 2160))
	img.Set(0, 0, color.NRGBA{R: 255, A: 255})
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	if err := validateImagePayload(encoded.Bytes()); err != nil {
		t.Fatalf("valid panel image rejected: %v", err)
	}
	if err := validateImagePayload([]byte("not an image")); err == nil {
		t.Fatal("invalid image accepted")
	}
	encoded.Reset()
	if err := png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 10001, 1))); err != nil {
		t.Fatal(err)
	}
	if err := validateImagePayload(encoded.Bytes()); err == nil {
		t.Fatal("oversized image dimensions accepted")
	}
}

func TestReadDeviceIDPrefersWiFiMAC(t *testing.T) {
	root := t.TempDir()
	writeInterfaceAddress(t, root, "usb0", "02:00:00:00:00:01")
	writeInterfaceAddress(t, root, "wlan0", "aa:bb:cc:dd:ee:ff")
	if got := readDeviceIDAt(root); got != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("readDeviceIDAt() = %q, want Wi-Fi MAC", got)
	}
}

func TestUntilNextBrightnessSlot(t *testing.T) {
	at := time.Date(2026, 8, 3, 8, 12, 30, 0, time.Local)
	if got, want := untilNextBrightnessSlot(at), 17*time.Minute+30*time.Second; got != want {
		t.Fatalf("untilNextBrightnessSlot() = %v, want %v", got, want)
	}
	at = time.Date(2026, 8, 3, 8, 30, 0, 0, time.Local)
	if got := untilNextBrightnessSlot(at); got != 30*time.Minute {
		t.Fatalf("boundary duration = %v, want 30m", got)
	}
}

// QML's Image ignores a source assignment that repeats the current URL, so
// every distinct cached screen must render to a distinct path.
func TestRenderedViewPathTracksItsSource(t *testing.T) {
	a := &app{dataDir: t.TempDir()}
	cfg := config.Defaults()
	cfg.Invert = true
	first := writeTestScreen(t, a.dataDir, "screen-1000.png")
	second := writeTestScreen(t, a.dataDir, "screen-2000.png")

	firstRender, err := a.renderedView(first, cfg)
	if err != nil {
		t.Fatalf("invert first screen: %v", err)
	}
	secondRender, err := a.renderedView(second, cfg)
	if err != nil {
		t.Fatalf("invert second screen: %v", err)
	}
	if firstRender == secondRender {
		t.Fatalf("both screens rendered to %s; the dashboard would never reload", firstRender)
	}
	repeat, err := a.renderedView(first, cfg)
	if err != nil {
		t.Fatalf("re-invert first screen: %v", err)
	}
	if repeat != firstRender {
		t.Fatalf("re-rendering %s produced %s, want the stable %s", first, repeat, firstRender)
	}
	for _, path := range []string{firstRender, secondRender} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("render %s is missing: %v", path, err)
		}
	}
}

// With no transform enabled the cached file is shown directly, so no render is
// written and nothing has to be pruned later.
func TestRenderedViewReturnsSourceWhenNothingApplies(t *testing.T) {
	a := &app{dataDir: t.TempDir()}
	source := writeTestScreen(t, a.dataDir, "screen-3000.png")
	got, err := a.renderedView(source, config.Defaults())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != source {
		t.Fatalf("render = %s, want the untouched source %s", got, source)
	}
}

// Dithering and inversion must produce distinct filenames so switching either
// setting reloads the panel.
func TestRenderedViewNamesEncodeTheTransforms(t *testing.T) {
	a := &app{dataDir: t.TempDir()}
	source := writeGradientScreen(t, a.dataDir, "screen-4000.png")

	inverted := config.Defaults()
	inverted.Invert = true
	dithered := config.Defaults()
	dithered.Dither = "auto"
	both := config.Defaults()
	both.Invert, both.Dither = true, "auto"

	seen := map[string]string{}
	for name, cfg := range map[string]config.Config{"invert": inverted, "dither": dithered, "both": both} {
		path, err := a.renderedView(source, cfg)
		if err != nil {
			t.Fatalf("%s render: %v", name, err)
		}
		if previous, clash := seen[path]; clash {
			t.Fatalf("%s and %s both rendered to %s", name, previous, path)
		}
		seen[path] = name
	}
}

func TestBatterySavingRespectsChargingAndThreshold(t *testing.T) {
	a := &app{}
	off := config.Defaults()
	off.BatterySaverPercent = 0
	if saving, _ := a.batterySaving(off); saving {
		t.Fatal("battery saving engaged while disabled")
	}
}

func TestPruneInvertedKeepsNewestRenders(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "inverted.png")
	if err := os.WriteFile(legacy, []byte("stale"), 0600); err != nil {
		t.Fatal(err)
	}
	var paths []string
	for i := range 5 {
		p := filepath.Join(dir, fmt.Sprintf("inverted-screen-%d.png", i))
		if err := os.WriteFile(p, []byte("render"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, time.Unix(int64(1000+i), 0), time.Unix(int64(1000+i), 0)); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	pruneInverted(dir, 2)
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("the pre-2.1 fixed-name render was not removed")
	}
	for i, p := range paths {
		_, err := os.Stat(p)
		if i >= 3 && err != nil {
			t.Fatalf("newest render %s was pruned: %v", p, err)
		}
		if i < 3 && err == nil {
			t.Fatalf("stale render %s survived pruning", p)
		}
	}
}

// Errors reach the tablet screen, the refresh history and the log file, so a
// signed image URL must not survive them.
func TestSafeErrorRedactsCredentialsAndSignedURLs(t *testing.T) {
	signed := errors.New(`Get "https://trmnl-assets.s3.amazonaws.com/screens/abc.png?X-Amz-Signature=deadbeef&X-Amz-Credential=AKIA": dial tcp: i/o timeout`)
	got := safeError(signed)
	for _, secret := range []string{"X-Amz-Signature", "deadbeef", "AKIA"} {
		if strings.Contains(got, secret) {
			t.Fatalf("safeError leaked %q: %s", secret, got)
		}
	}
	if !strings.Contains(got, "trmnl-assets.s3.amazonaws.com") {
		t.Fatalf("safeError dropped the origin operators need: %s", got)
	}

	if got := safeError(errors.New(`request failed with access-token abc123`)); strings.Contains(got, "abc123") {
		t.Fatalf("safeError leaked the access token: %s", got)
	}

	plain := errors.New(`Get "https://trmnl.com/api/display": connection refused`)
	if got := safeError(plain); !strings.Contains(got, "https://trmnl.com/api/display") {
		t.Fatalf("safeError mangled a URL that carries no secrets: %s", got)
	}
	if safeError(nil) != "" {
		t.Fatal("safeError(nil) should be empty")
	}
}

func writeGradientScreen(t *testing.T, dir, name string) string {
	t.Helper()
	var encoded bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 10), G: uint8(y * 10), B: 120, A: 255})
		}
	}
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeTestScreen(t *testing.T, dir, name string) string {
	t.Helper()
	var encoded bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeInterfaceAddress(t *testing.T, root, name, address string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "address"), []byte(address+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func writePowerSupply(t *testing.T, root, name, typ, capacity string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "type"), []byte(typ+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "capacity"), []byte(capacity+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
}
