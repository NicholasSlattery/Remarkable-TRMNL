package brightness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoveryAndMapping(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "rm_frontlight")
	if err := os.Mkdir(p, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "max_brightness"), []byte("2047\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "brightness"), []byte("1024\n"), 0644); err != nil {
		t.Fatal(err)
	}
	d, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	for percent, want := range map[int]int{0: 0, 25: 512, 50: 1024, 75: 1535, 100: 2047} {
		if got := d.ValueForPercent(percent); got != want {
			t.Fatalf("%d%%=%d want %d", percent, got, want)
		}
	}
	if err := d.SetPercent(25); err != nil {
		t.Fatal(err)
	}
	if d.Current != 512 {
		t.Fatalf("current=%d", d.Current)
	}
}
