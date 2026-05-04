package version

import (
	"strings"
	"testing"
)

func TestGetReturnsDefaultVersionInfo(t *testing.T) {
	got := Get()

	if got.Name != "nomos" {
		t.Fatalf("expected name nomos, got %q", got.Name)
	}
	if got.Version == "" || got.Commit == "" || got.Date == "" || got.Dirty == "" || got.BuiltBy == "" {
		t.Fatalf("expected non-empty metadata, got %#v", got)
	}
	if !strings.HasPrefix(got.Go, "go") {
		t.Fatalf("expected go runtime version, got %q", got.Go)
	}
	if got.OS == "" || got.Arch == "" {
		t.Fatalf("expected os/arch to be set, got os=%q arch=%q", got.OS, got.Arch)
	}
}

func TestGetReturnsConfiguredVersionInfo(t *testing.T) {
	origVersion, origCommit, origDate, origDirty, origBuiltBy := Version, Commit, Date, Dirty, BuiltBy
	defer func() {
		Version, Commit, Date, Dirty, BuiltBy = origVersion, origCommit, origDate, origDirty, origBuiltBy
	}()

	Version = "v0.1.0"
	Commit = "abc1234"
	Date = "2026-05-04T12:00:00Z"
	Dirty = "false"
	BuiltBy = "test"

	got := Get()
	if got.Version != "v0.1.0" || got.Commit != "abc1234" || got.Date != "2026-05-04T12:00:00Z" || got.Dirty != "false" || got.BuiltBy != "test" {
		t.Fatalf("unexpected metadata: %#v", got)
	}
}
