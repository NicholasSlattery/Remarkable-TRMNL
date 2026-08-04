package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
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
