// Package update provides an opt-in check for newer Nomos releases.
//
// The check is disabled by default and only runs when NOMOS_UPDATE_CHECK is set
// to a truthy value, keeping Nomos offline-friendly and avoiding surprise
// network traffic. Results are cached and refreshed in the background so that
// page renders and API calls never block on the network.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	versionpkg "github.com/nomos/nomos/internal/version"
)

const (
	releasesURL = "https://api.github.com/repos/pblumer/nomos/releases/latest"
	cacheTTL    = time.Hour
)

// Status is a snapshot of the update-check result, safe to expose via the API
// and HTML templates.
type Status struct {
	Enabled         bool      `json:"enabled"`
	Current         string    `json:"current"`
	Latest          string    `json:"latest,omitempty"`
	UpdateAvailable bool      `json:"updateAvailable"`
	CheckedAt       time.Time `json:"checkedAt,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// Enabled reports whether the online update check has been opted into via the
// NOMOS_UPDATE_CHECK environment variable.
func Enabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NOMOS_UPDATE_CHECK"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// Checker performs cached, opt-in update checks against GitHub releases.
type Checker struct {
	client   *http.Client
	mu       sync.Mutex
	cached   Status
	fetched  time.Time
	inflight bool
}

func NewChecker() *Checker {
	return &Checker{client: &http.Client{Timeout: 5 * time.Second}}
}

// Status returns the current cached status. When the check is enabled and the
// cache is stale, it triggers a background refresh and returns the previously
// cached value immediately, so callers never block on the network.
func (c *Checker) Status() Status {
	current := versionpkg.Get().Version
	if !Enabled() {
		return Status{Enabled: false, Current: current}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cached.Enabled = true
	c.cached.Current = current
	if !c.inflight && (c.fetched.IsZero() || time.Since(c.fetched) > cacheTTL) {
		c.inflight = true
		go c.refresh(current)
	}
	return c.cached
}

func (c *Checker) refresh(current string) {
	latest, err := c.fetchLatest()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.inflight = false
	c.fetched = time.Now()
	c.cached.CheckedAt = c.fetched
	if err != nil {
		c.cached.Error = err.Error()
		return
	}
	c.cached.Error = ""
	c.cached.Latest = latest
	c.cached.UpdateAvailable = isNewer(latest, current)
}

func (c *Checker) fetchLatest() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github returned status %d", resp.StatusCode)
	}
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("release response had no tag_name")
	}
	return payload.TagName, nil
}

// isNewer reports whether latest is a strictly higher semver than current.
// Non-semver versions (e.g. the "dev" default) never count as updatable.
func isNewer(latest, current string) bool {
	lv, ok1 := parseSemver(latest)
	cv, ok2 := parseSemver(current)
	if !ok1 || !ok2 {
		return false
	}
	for i := 0; i < 3; i++ {
		if lv[i] != cv[i] {
			return lv[i] > cv[i]
		}
	}
	return false
}

func parseSemver(s string) ([3]int, bool) {
	var out [3]int
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return out, false
	}
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
