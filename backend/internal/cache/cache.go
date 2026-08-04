package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Entry struct {
	Path         string    `json:"path"`
	ImageURL     string    `json:"image_url"`
	ImageName    string    `json:"image_name"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"last_modified,omitempty"`
	SavedAt      time.Time `json:"saved_at"`
}

type Store struct{ Dir string }

func (s Store) ensure() error { return os.MkdirAll(s.Dir, 0700) }

func (s Store) Entries() ([]Entry, error) {
	b, err := os.ReadFile(filepath.Join(s.Dir, "index.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, err
	}
	valid := entries[:0]
	for _, e := range entries {
		if _, err := os.Stat(e.Path); err == nil {
			valid = append(valid, e)
		}
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].SavedAt.After(valid[j].SavedAt) })
	return valid, nil
}

func (s Store) Latest() (Entry, bool) {
	es, err := s.Entries()
	if err != nil || len(es) == 0 {
		return Entry{}, false
	}
	return es[0], true
}

func (s Store) Put(name string, src io.Reader, entry Entry, maxBytes int64, keep int) (Entry, error) {
	if err := s.ensure(); err != nil {
		return Entry{}, err
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".bmp" {
		ext = ".img"
	}
	tmp, err := os.CreateTemp(s.Dir, ".download-*.tmp")
	if err != nil {
		return Entry{}, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	n, err := io.Copy(tmp, io.LimitReader(src, maxBytes+1))
	if err != nil {
		tmp.Close()
		return Entry{}, err
	}
	if n == 0 || n > maxBytes {
		tmp.Close()
		return Entry{}, fmt.Errorf("invalid image size: %d", n)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return Entry{}, err
	}
	if err := tmp.Close(); err != nil {
		return Entry{}, err
	}
	final := filepath.Join(s.Dir, fmt.Sprintf("screen-%d%s", time.Now().UnixNano(), ext))
	if err := os.Rename(tmpName, final); err != nil {
		return Entry{}, err
	}
	entry.Path, entry.SavedAt = final, time.Now().UTC()
	es, _ := s.Entries()
	es = append([]Entry{entry}, es...)
	if keep < 2 {
		keep = 2
	}
	if len(es) > keep {
		for _, old := range es[keep:] {
			_ = os.Remove(old.Path)
		}
		es = es[:keep]
	}
	if err := s.writeIndex(es); err != nil {
		_ = os.Remove(final)
		return Entry{}, err
	}
	return entry, nil
}

func (s Store) writeIndex(entries []Entry) error {
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Dir, ".index-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(s.Dir, "index.json"))
}

func (s Store) Clear() error {
	es, _ := s.Entries()
	for _, e := range es {
		_ = os.Remove(e.Path)
	}
	_ = os.Remove(filepath.Join(s.Dir, "index.json"))
	return nil
}
