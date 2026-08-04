package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
}

func Defaults() Config {
	return Config{
		BaseURL: DefaultBaseURL, RefreshMode: "server", MinimumRefreshSeconds: 60,
		FitMode: "fit", Orientation: "auto", UseSystemBrightness: true,
		RestoreBrightnessOnExit: true, BrightnessPercent: 50,
		StartWithCacheOffline: true, WakeForRefresh: true,
		LoggingLevel: "info", HistoryLimit: 30,
	}
}

func (c *Config) Normalize() error {
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Hostname() == "" {
		return errors.New("server URL is invalid")
	}
	if u.Scheme != "https" {
		isLoopback := u.Scheme == "http" && (u.Hostname() == "localhost" || net.ParseIP(u.Hostname()) != nil && net.ParseIP(u.Hostname()).IsLoopback())
		if !isLoopback {
			return errors.New("server URL must use HTTPS; HTTP is allowed only for a mock server on this device")
		}
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
