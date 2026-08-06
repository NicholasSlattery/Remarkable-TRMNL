package brightness

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Device is shared between the AppLoad receive loop, the refresh scheduler and
// the shutdown path, so every access to the mutable Current field is guarded.
// Path, Name, Min, Max, Original and Writable are fixed at discovery time.
type Device struct {
	mu       sync.Mutex
	writeMu  sync.Mutex
	Path     string `json:"path"`
	Name     string `json:"name"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Current  int    `json:"current"`
	Original int    `json:"original"`
	Writable bool   `json:"writable"`
}

// Info is a lock-free snapshot suitable for JSON encoding on another goroutine.
type Info struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Min      int    `json:"min"`
	Max      int    `json:"max"`
	Current  int    `json:"current"`
	Original int    `json:"original"`
	Writable bool   `json:"writable"`
}

// Snapshot copies the device state under the lock. Callers must use it instead
// of marshaling *Device directly.
func (d *Device) Snapshot() Info {
	d.mu.Lock()
	defer d.mu.Unlock()
	return Info{Path: d.Path, Name: d.Name, Min: d.Min, Max: d.Max, Current: d.Current, Original: d.Original, Writable: d.Writable}
}

func Discover(root string) (*Device, error) {
	paths, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		return nil, err
	}
	sort.SliceStable(paths, func(i, j int) bool {
		return strings.Contains(filepath.Base(paths[i]), "rm_frontlight") && !strings.Contains(filepath.Base(paths[j]), "rm_frontlight")
	})
	for _, p := range paths {
		max, e1 := readInt(filepath.Join(p, "max_brightness"))
		cur, e2 := readInt(filepath.Join(p, "brightness"))
		if e1 != nil || e2 != nil || max <= 0 {
			continue
		}
		min := 0
		if v, err := readInt(filepath.Join(p, "min_brightness")); err == nil {
			min = v
		}
		if min < 0 || min >= max {
			min = 0
		}
		return &Device{Path: p, Name: filepath.Base(p), Min: min, Max: max, Current: cur, Original: cur, Writable: isWritable(filepath.Join(p, "brightness"))}, nil
	}
	return nil, errors.New("no usable backlight device found")
}

func (d *Device) Read() (int, error) {
	v, err := readInt(filepath.Join(d.Path, "brightness"))
	if err == nil {
		d.mu.Lock()
		d.Current = v
		d.mu.Unlock()
	}
	return v, err
}

func (d *Device) Percent() int {
	d.mu.Lock()
	current := d.Current
	d.mu.Unlock()
	if d.Max <= d.Min {
		return 0
	}
	p := (current - d.Min) * 100 / (d.Max - d.Min)
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

func (d *Device) ValueForPercent(percent int) int {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return d.Min + ((d.Max-d.Min)*percent+50)/100
}

func (d *Device) SetPercent(percent int) error { return d.SetValue(d.ValueForPercent(percent)) }

func (d *Device) SetValue(value int) error {
	if value < d.Min || value > d.Max {
		return fmt.Errorf("brightness %d outside [%d,%d]", value, d.Min, d.Max)
	}
	// Serialize write-then-verify so a concurrent change cannot make a
	// successful write look like a hardware failure.
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	if err := os.WriteFile(filepath.Join(d.Path, "brightness"), []byte(strconv.Itoa(value)+"\n"), 0644); err != nil {
		return err
	}
	read, err := d.Read()
	if err != nil {
		return err
	}
	if read != value {
		return fmt.Errorf("brightness verification failed: wrote %d, read %d", value, read)
	}
	return nil
}

func (d *Device) Restore() error { return d.SetValue(d.Original) }

func readInt(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

func isWritable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	return f.Close() == nil
}
