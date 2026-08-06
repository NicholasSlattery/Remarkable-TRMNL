package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// knownHosts remembers the SSH host key each tablet presented, so a key that
// changes between installer runs is caught rather than silently re-trusted.
// Without it the installer trusts on first use every single time it starts.
type knownHosts struct {
	mu    sync.Mutex
	path  string
	Hosts map[string]string `json:"hosts"`
}

func loadKnownHosts(path string) *knownHosts {
	k := &knownHosts{path: path, Hosts: map[string]string{}}
	b, err := os.ReadFile(path)
	if err != nil {
		return k
	}
	var stored struct {
		Hosts map[string]string `json:"hosts"`
	}
	if err := json.Unmarshal(b, &stored); err == nil && stored.Hosts != nil {
		k.Hosts = stored.Hosts
	}
	return k
}

// defaultKnownHostsPath keeps the record beside the user's other application
// data rather than next to the executable, which may sit on read-only media.
func defaultKnownHostsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "trmnl-remarkable", "known-tablets.json")
}

func hostKey(host string) string {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "" {
		h = "10.11.99.1"
	}
	if !strings.Contains(h, ":") {
		h += ":22"
	}
	return h
}

// Lookup returns the fingerprint recorded for host, if any.
func (k *knownHosts) Lookup(host string) (string, bool) {
	if k == nil {
		return "", false
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	value, ok := k.Hosts[hostKey(host)]
	return value, ok && value != ""
}

// Remember stores a fingerprint the user has explicitly approved.
func (k *knownHosts) Remember(host, fingerprint string) {
	if k == nil || fingerprint == "" {
		return
	}
	k.mu.Lock()
	k.Hosts[hostKey(host)] = fingerprint
	snapshot := make(map[string]string, len(k.Hosts))
	for key, value := range k.Hosts {
		snapshot[key] = value
	}
	path := k.path
	k.mu.Unlock()
	if path == "" {
		return
	}
	b, err := json.MarshalIndent(struct {
		Hosts map[string]string `json:"hosts"`
	}{Hosts: snapshot}, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".known-*.tmp")
	if err != nil {
		return
	}
	name := tmp.Name()
	defer os.Remove(name)
	_ = tmp.Chmod(0600)
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}
	_ = os.Rename(name, path)
}
