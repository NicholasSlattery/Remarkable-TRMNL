package trmnl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trmnl-remarkable/backend/internal/config"
)

type FlexibleInt int

func (f *FlexibleInt) UnmarshalJSON(b []byte) error {
	var n int
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		n = v
	} else if string(b) != "null" {
		if err := json.Unmarshal(b, &n); err != nil {
			return err
		}
	}
	*f = FlexibleInt(n)
	return nil
}

type DisplayResponse struct {
	Status               int         `json:"status"`
	ImageURL             string      `json:"image_url"`
	ImageName            string      `json:"image_name"`
	Filename             string      `json:"filename"`
	RefreshRate          FlexibleInt `json:"refresh_rate"`
	ImageURLTimeout      FlexibleInt `json:"image_url_timeout"`
	SpecialFunction      string      `json:"special_function"`
	UpdateFirmware       bool        `json:"update_firmware"`
	ResetFirmware        bool        `json:"reset_firmware"`
	MaximumCompatibility bool        `json:"maximum_compatibility"`
}

type Client struct {
	HTTP    *http.Client
	Version string
	Battery func() string
	RSSI    func() string
}

func New() *Client { return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}, Version: "dev"} }

func (c *Client) Display(ctx context.Context, cfg config.Config, advance bool) (DisplayResponse, error) {
	paths := []string{"/api/display"}
	if !advance {
		// The first two paths are non-advancing. Older BYOS implementations
		// expose neither, so /api/display is the compatibility fallback.
		paths = []string{"/api/display/current", "/api/current_screen", "/api/display"}
	}
	var last error
	for _, p := range paths {
		resp, err := c.get(ctx, cfg, p)
		if err == nil {
			return resp, nil
		}
		last = err
		var he *HTTPError
		if advance || !errors.As(err, &he) || he.Status != http.StatusNotFound {
			break
		}
	}
	return DisplayResponse{}, last
}

type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string { return fmt.Sprintf("TRMNL server returned HTTP %d", e.Status) }

func (c *Client) get(ctx context.Context, cfg config.Config, path string) (DisplayResponse, error) {
	base, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil {
		return DisplayResponse{}, err
	}
	ref, _ := url.Parse(path)
	endpoint := base.ResolveReference(ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return DisplayResponse{}, err
	}
	if cfg.APIKey != "" {
		req.Header.Set("access-token", cfg.APIKey)
	}
	if cfg.DeviceID != "" {
		req.Header.Set("ID", cfg.DeviceID)
	}
	req.Header.Set("User-Agent", "trmnl-remarkable/"+c.Version)
	req.Header.Set("model", "reMarkable Paper Pro")
	req.Header.Set("firmware-version", c.Version)
	if c.Battery != nil {
		if v := c.Battery(); v != "" {
			req.Header.Set("battery-voltage", v)
		}
	}
	if c.RSSI != nil {
		if v := c.RSSI(); v != "" {
			req.Header.Set("rssi", v)
		}
	}
	r, err := c.HTTP.Do(req)
	if err != nil {
		return DisplayResponse{}, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(r.Body, 1024))
		return DisplayResponse{}, &HTTPError{Status: r.StatusCode, Body: string(b)}
	}
	var out DisplayResponse
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&out); err != nil {
		return out, fmt.Errorf("decode display response: %w", err)
	}
	if out.ImageURL == "" {
		return out, errors.New("display response contained no image_url")
	}
	if _, err := url.ParseRequestURI(out.ImageURL); err != nil {
		return out, fmt.Errorf("invalid image_url: %w", err)
	}
	return out, nil
}

func (c *Client) Download(ctx context.Context, imageURL string, timeout time.Duration, etag, lastModified string) (io.ReadCloser, http.Header, error) {
	if timeout <= 0 || timeout > 2*time.Minute {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	_ = cancel
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	r, err := c.HTTP.Do(req)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	if r.StatusCode == http.StatusNotModified {
		_ = r.Body.Close()
		cancel()
		return nil, r.Header, ErrNotModified
	}
	if r.StatusCode != http.StatusOK {
		r.Body.Close()
		cancel()
		return nil, nil, fmt.Errorf("image server returned HTTP %d", r.StatusCode)
	}
	return &cancelReadCloser{ReadCloser: r.Body, cancel: cancel}, r.Header, nil
}

var ErrNotModified = errors.New("image not modified")

type cancelReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelReadCloser) Close() error { err := c.ReadCloser.Close(); c.cancel(); return err }
