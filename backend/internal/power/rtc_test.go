package power

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscoverSetAndClear(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "rtc0")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	alarm := filepath.Join(dir, "wakealarm")
	if err := os.WriteFile(alarm, nil, 0600); err != nil {
		t.Fatal(err)
	}
	rtc, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	wake := time.Now().Add(10 * time.Minute).Truncate(time.Second)
	if err := rtc.Set(wake); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(alarm); err != nil || string(b) != fmtEpoch(wake) {
		t.Fatalf("alarm=%q err=%v", string(b), err)
	}
	if err := rtc.Clear(); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(alarm); string(b) != "0\n" {
		t.Fatalf("cleared alarm=%q", string(b))
	}
}

func fmtEpoch(at time.Time) string {
	return fmt.Sprintf("%d\n", at.Unix())
}
