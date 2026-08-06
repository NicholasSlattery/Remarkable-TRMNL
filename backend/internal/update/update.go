// Package update checks the project's GitHub releases for a newer version.
// It is opt-in: it is the only request the app makes to a host other than the
// configured dashboard server, so it stays off until the owner enables it.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const releasesURL = "https://api.github.com/repos/NicholasSlattery/Remarkable-TRMNL/releases/latest"

// Result describes the newest published release relative to the running build.
type Result struct {
	Latest    string    `json:"latest_version"`
	URL       string    `json:"release_url,omitempty"`
	Available bool      `json:"update_available"`
	CheckedAt time.Time `json:"checked_at"`
}

// Checker caches a result so a restart loop cannot turn into a request loop.
type Checker struct {
	HTTP     *http.Client
	Endpoint string
	Interval time.Duration

	last      Result
	lastError string
	fetchedAt time.Time
}

func New() *Checker {
	return &Checker{
		HTTP:     &http.Client{Timeout: 20 * time.Second},
		Endpoint: releasesURL,
		Interval: 24 * time.Hour,
	}
}

// Check returns the cached result when it is still fresh, otherwise it queries
// GitHub. The current version is compared with semantic-version ordering.
func (c *Checker) Check(ctx context.Context, current string, now time.Time) (Result, error) {
	if !c.fetchedAt.IsZero() && now.Sub(c.fetchedAt) < c.Interval {
		if c.lastError != "" {
			return c.last, fmt.Errorf("%s", c.lastError)
		}
		return c.last, nil
	}
	result, err := c.fetch(ctx, current, now)
	c.fetchedAt = now
	if err != nil {
		c.lastError = err.Error()
		return Result{}, err
	}
	c.lastError = ""
	c.last = result
	return result, nil
}

func (c *Checker) fetch(ctx context.Context, current string, now time.Time) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoint, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "trmnl-remarkable/"+current)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("release feed returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Draft   bool   `json:"draft"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return Result{}, fmt.Errorf("decode release feed: %w", err)
	}
	latest := strings.TrimPrefix(strings.TrimSpace(payload.TagName), "v")
	if latest == "" {
		return Result{}, fmt.Errorf("release feed contained no tag")
	}
	return Result{
		Latest:    latest,
		URL:       payload.HTMLURL,
		Available: !payload.Draft && Newer(latest, current),
		CheckedAt: now,
	}, nil
}

// Newer reports whether candidate sorts above current. Unparseable or
// development versions never report an update, so a dev build stays quiet.
func Newer(candidate, current string) bool {
	a, okA := parse(candidate)
	b, okB := parse(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func parse(value string) ([3]int, bool) {
	var out [3]int
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	if i := strings.IndexAny(value, "-+"); i >= 0 {
		value = value[:i]
	}
	fields := strings.Split(value, ".")
	if len(fields) < 2 || len(fields) > 3 {
		return out, false
	}
	for i, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
