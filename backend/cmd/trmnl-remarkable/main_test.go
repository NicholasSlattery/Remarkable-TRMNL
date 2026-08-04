package main

import (
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
