package power

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// RTC represents a Linux real-time clock that can wake the tablet from suspend.
type RTC struct {
	Name      string `json:"name"`
	AlarmPath string `json:"alarm_path"`
}

// Discover finds the first writable RTC wakealarm below root. The Paper Pro
// normally exposes this as /sys/class/rtc/rtc0/wakealarm.
func Discover(root string) (*RTC, error) {
	paths, err := filepath.Glob(filepath.Join(root, "rtc*", "wakealarm"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, alarm := range paths {
		f, err := os.OpenFile(alarm, os.O_WRONLY, 0)
		if err != nil {
			continue
		}
		_ = f.Close()
		return &RTC{Name: filepath.Base(filepath.Dir(alarm)), AlarmPath: alarm}, nil
	}
	return nil, errors.New("no writable RTC wake alarm found")
}

// Set programs an absolute wake time. Linux requires an existing alarm to be
// cleared before a replacement is written.
func (r *RTC) Set(at time.Time) error {
	if r == nil || r.AlarmPath == "" {
		return errors.New("RTC wake alarm is unavailable")
	}
	if at.Before(time.Now().Add(5 * time.Second)) {
		return errors.New("RTC wake time must be in the future")
	}
	if err := os.WriteFile(r.AlarmPath, []byte("0\n"), 0644); err != nil {
		return fmt.Errorf("clear RTC wake alarm: %w", err)
	}
	want := at.Unix()
	if err := os.WriteFile(r.AlarmPath, []byte(strconv.FormatInt(want, 10)+"\n"), 0644); err != nil {
		return fmt.Errorf("set RTC wake alarm: %w", err)
	}
	b, err := os.ReadFile(r.AlarmPath)
	if err != nil {
		return fmt.Errorf("verify RTC wake alarm: %w", err)
	}
	got, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil || got != want {
		return fmt.Errorf("verify RTC wake alarm: wrote %d, read %q", want, strings.TrimSpace(string(b)))
	}
	return nil
}

func (r *RTC) Clear() error {
	if r == nil || r.AlarmPath == "" {
		return nil
	}
	if err := os.WriteFile(r.AlarmPath, []byte("0\n"), 0644); err != nil {
		return fmt.Errorf("clear RTC wake alarm: %w", err)
	}
	return nil
}
