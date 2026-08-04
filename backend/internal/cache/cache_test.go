package cache

import (
	"bytes"
	"os"
	"testing"
)

func TestPutIsAtomicAndPrunes(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	for i := 0; i < 4; i++ {
		if _, err := s.Put("x.png", bytes.NewReader([]byte{1, 2, 3, byte(i)}), Entry{ImageName: "x"}, 100, 3); err != nil {
			t.Fatal(err)
		}
	}
	es, err := s.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 3 {
		t.Fatalf("entries=%d", len(es))
	}
	for _, e := range es {
		if _, err := os.Stat(e.Path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRejectsOversize(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	if _, err := s.Put("x.png", bytes.NewReader(make([]byte, 11)), Entry{}, 10, 2); err == nil {
		t.Fatal("expected error")
	}
}
