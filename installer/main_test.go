package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseKeyValues(t *testing.T) {
	got := parseKeyValues("model=reMarkable Ferrari\nos=3.27.3.0\ninstalled=yes\nactive=no\n")
	if got["model"] != "reMarkable Ferrari" || got["installed"] != "yes" || got["active"] != "no" {
		t.Fatalf("unexpected values: %#v", got)
	}
}

func TestVerifyPayload(t *testing.T) {
	dir := t.TempDir()
	data := []byte("safe payload")
	if err := os.WriteFile(filepath.Join(dir, "file.bin"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	m := payloadManifest{Version: "1.0.0", Files: map[string]string{"file.bin": hex.EncodeToString(sum[:])}}
	b, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyPayload(dir); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}
	m.Files["../escape"] = hex.EncodeToString(sum[:])
	b, _ = json.Marshal(m)
	_ = os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0o600)
	if _, err := verifyPayload(dir); err == nil {
		t.Fatal("unsafe payload path accepted")
	}
}

func TestShellQuote(t *testing.T) {
	if got := shellQuote("a'b"); got != `'a'"'"'b'` {
		t.Fatalf("unexpected shell quote: %s", got)
	}
}

func TestInstallerPageRenders(t *testing.T) {
	s := &server{csrf: "test-token"}
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	s.index(recorder, req)
	if recorder.Code != 200 {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, expected := range []string{"TRMNL for reMarkable", "Find my tablet", "Developer Mode", "Reactivate after reboot", "Uninstall and erase data", "test-token"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("installer page missing %q", expected)
		}
	}
	if strings.Contains(body, "{{CSRF}}") || strings.Contains(body, "{{VERSION}}") {
		t.Fatal("installer template placeholders were not replaced")
	}
}
