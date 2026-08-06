package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://trmnl.com"

type Config struct {
	APIKey                  string `json:"api_key,omitempty"`
	BaseURL                 string `json:"base_url"`
	DeviceID                string `json:"device_id,omitempty"`
	RefreshMode             string `json:"refresh_mode"`
	MinimumRefreshSeconds   int    `json:"minimum_refresh_seconds"`
	FitMode                 string `json:"fit_mode"`
	Orientation             string `json:"orientation"`
	Invert                  bool   `json:"invert"`
	UseSystemBrightness     bool   `json:"use_system_brightness"`
	RestoreBrightnessOnExit bool   `json:"restore_brightness_on_exit"`
	BrightnessPercent       int    `json:"brightness_percent"`
	StartWithCacheOffline   bool   `json:"start_with_cache_offline"`
	AlwaysOn                bool   `json:"always_on"`
	WakeForRefresh          bool   `json:"wake_for_refresh"`
	LoggingLevel            string `json:"logging_level"`
	HistoryLimit            int    `json:"history_limit"`

	// QuietHours suppresses scheduled refreshes overnight. The window may wrap
	// past midnight. Manual refreshes are never suppressed.
	QuietHoursEnabled bool   `json:"quiet_hours_enabled"`
	QuietHoursStart   string `json:"quiet_hours_start"`
	QuietHoursEnd     string `json:"quiet_hours_end"`

	// BatterySaverPercent lengthens the refresh interval once the battery falls
	// to this level while running on battery. Zero disables it.
	BatterySaverPercent int `json:"battery_saver_percent"`

	// Dither is "off" or "auto". Auto applies Floyd-Steinberg error diffusion to
	// DitherPalette when the source image contains more colors than the panel
	// can show, which removes banding in gradients.
	Dither        string   `json:"dither"`
	DitherPalette []string `json:"dither_palette,omitempty"`

	// UpdateCheck contacts the GitHub releases API. It is off by default because
	// it is the only outbound request not directed at the configured dashboard.
	UpdateCheck bool `json:"update_check"`
}

// DefaultDitherPalette approximates the Paper Pro colour panel. Override it in
// config.json if a firmware renders a different set.
var DefaultDitherPalette = []string{"#000000", "#ffffff", "#c33124", "#e6b422", "#2a5caa", "#2f7a4a"}

const (
	BatterySaverMultiplier = 4
	BatterySaverMaxRefresh = 6 * time.Hour
)

func Defaults() Config {
	return Config{
		BaseURL: DefaultBaseURL, RefreshMode: "server", MinimumRefreshSeconds: 60,
		FitMode: "fit", Orientation: "auto", UseSystemBrightness: true,
		RestoreBrightnessOnExit: true, BrightnessPercent: 50,
		StartWithCacheOffline: true, WakeForRefresh: true,
		LoggingLevel: "info", HistoryLimit: 30,
		QuietHoursStart: "23:00", QuietHoursEnd: "07:00",
		BatterySaverPercent: 20, Dither: "off",
	}
}

// ParseClock accepts a 24-hour HH:MM value and returns minutes past midnight.
func ParseClock(value string) (int, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("expected a 24-hour HH:MM time, got %q", value)
	}
	return t.Hour()*60 + t.Minute(), nil
}

// InQuietHours reports whether at falls inside the configured window. The
// window is half-open so a refresh scheduled exactly at the end time proceeds.
func (c Config) InQuietHours(at time.Time) bool {
	if !c.QuietHoursEnabled {
		return false
	}
	start, err := ParseClock(c.QuietHoursStart)
	if err != nil {
		return false
	}
	end, err := ParseClock(c.QuietHoursEnd)
	if err != nil || start == end {
		return false
	}
	minutes := at.Hour()*60 + at.Minute()
	if start < end {
		return minutes >= start && minutes < end
	}
	// The window wraps past midnight.
	return minutes >= start || minutes < end
}

// NextActiveTime moves a scheduled refresh to the first moment after the quiet
// window closes, leaving times outside the window untouched.
func (c Config) NextActiveTime(at time.Time) time.Time {
	if !c.InQuietHours(at) {
		return at
	}
	end, err := ParseClock(c.QuietHoursEnd)
	if err != nil {
		return at
	}
	resume := time.Date(at.Year(), at.Month(), at.Day(), end/60, end%60, 0, 0, at.Location())
	if !resume.After(at) {
		resume = resume.AddDate(0, 0, 1)
	}
	return resume
}

// Palette returns the configured dither palette, falling back to the default
// when it is unset or contains no usable colours.
func (c Config) Palette() []string {
	if len(c.DitherPalette) > 0 {
		return c.DitherPalette
	}
	return DefaultDitherPalette
}

func (c *Config) Normalize() error {
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if err := ValidateRemoteURL(c.BaseURL); err != nil {
		return fmt.Errorf("server URL: %w", err)
	}
	u, _ := url.Parse(c.BaseURL)
	if u.Path != "" && u.Path != "/" || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("server URL must be an origin without a path, query, or fragment")
	}
	if c.MinimumRefreshSeconds < 60 {
		c.MinimumRefreshSeconds = 60
	}
	if c.MinimumRefreshSeconds > 86400 {
		c.MinimumRefreshSeconds = 86400
	}
	if c.BrightnessPercent < 0 {
		c.BrightnessPercent = 0
	}
	if c.BrightnessPercent > 100 {
		c.BrightnessPercent = 100
	}
	if c.HistoryLimit < 5 || c.HistoryLimit > 200 {
		c.HistoryLimit = 30
	}
	switch c.FitMode {
	case "fit", "fill", "stretch":
	default:
		c.FitMode = "fit"
	}
	switch c.Orientation {
	case "auto", "portrait", "landscape":
	default:
		c.Orientation = "auto"
	}
	switch c.Dither {
	case "off", "auto":
	default:
		c.Dither = "off"
	}
	if c.BatterySaverPercent < 0 || c.BatterySaverPercent > 90 {
		c.BatterySaverPercent = 0
	}
	if c.QuietHoursStart == "" {
		c.QuietHoursStart = "23:00"
	}
	if c.QuietHoursEnd == "" {
		c.QuietHoursEnd = "07:00"
	}
	if _, err := ParseClock(c.QuietHoursStart); err != nil {
		return fmt.Errorf("quiet hours start: %w", err)
	}
	if _, err := ParseClock(c.QuietHoursEnd); err != nil {
		return fmt.Errorf("quiet hours end: %w", err)
	}
	for _, value := range c.DitherPalette {
		if _, _, _, err := ParseHexColor(value); err != nil {
			return fmt.Errorf("dither palette: %w", err)
		}
	}
	return nil
}

// ParseHexColor accepts #rgb or #rrggbb and returns the 8-bit components.
func ParseHexColor(value string) (r, g, b uint8, err error) {
	s := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("%q is not a #rgb or #rrggbb colour", value)
	}
	n, convErr := strconv.ParseUint(s, 16, 32)
	if convErr != nil {
		return 0, 0, 0, fmt.Errorf("%q is not a #rgb or #rrggbb colour", value)
	}
	return uint8(n >> 16), uint8(n >> 8), uint8(n), nil
}

// ValidateRemoteURL applies the transport policy shared by Device API and
// image requests. Production traffic must use HTTPS; plain HTTP is limited to
// a loopback mock running on the tablet itself.
func ValidateRemoteURL(value string) error {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil || !u.IsAbs() || u.Hostname() == "" {
		return errors.New("URL is invalid")
	}
	if u.User != nil {
		return errors.New("embedded usernames or passwords are not allowed")
	}
	if strings.EqualFold(u.Scheme, "https") {
		return nil
	}
	ip := net.ParseIP(u.Hostname())
	isLoopback := u.Scheme == "http" && (strings.EqualFold(u.Hostname(), "localhost") || ip != nil && ip.IsLoopback())
	if !isLoopback {
		return errors.New("HTTPS is required; HTTP is allowed only for a loopback mock on this device")
	}
	return nil
}

func Load(path string) (Config, error) {
	c := Defaults()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return Defaults(), fmt.Errorf("parse config: %w", err)
	}
	return c, c.Normalize()
}

func Save(path string, c Config) error {
	if err := c.Normalize(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
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
	if err := os.Rename(name, path); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

func Redacted(c Config) Config {
	if c.APIKey != "" {
		c.APIKey = "********"
	}
	return c
}
