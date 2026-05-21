package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestLocalRegistryReturnsDefaultRepository(t *testing.T) {
	ws := t.TempDir()
	repos := NewLocalRegistry(ws).List()
	if len(repos) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(repos))
	}
	r := repos[0]
	if r.ID != LocalDefaultID {
		t.Errorf("ID = %q, want %q", r.ID, LocalDefaultID)
	}
	if r.Kind != KindFilesystem {
		t.Errorf("Kind = %q, want %q", r.Kind, KindFilesystem)
	}
	if r.Location != ws {
		t.Errorf("Location = %q, want %q", r.Location, ws)
	}
	if r.Status != "unknown" {
		t.Errorf("Status = %q, want unknown for a non-git workspace", r.Status)
	}
}

func TestCreateFilesystemDefaultLocationIsWorkspaceRelative(t *testing.T) {
	ws := t.TempDir()
	r, err := NewLocalRegistry(ws).CreateFilesystem("acme", "Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	want := filepath.Join(ws, ".nomos", "repos", "acme")
	if r.Location != want {
		t.Errorf("Location = %q, want %q", r.Location, want)
	}
	// The repo must reappear (resolved) on a fresh List from the same workspace.
	if !containsID(NewLocalRegistry(ws).List(), "acme") {
		t.Errorf("created repository not found in List")
	}
	// New repositories are git-first: a git database is initialized on creation.
	if _, err := os.Stat(filepath.Join(want, ".git")); err != nil {
		t.Errorf("expected git repository to be initialized: %v", err)
	}
}

func TestCreateFilesystemHonorsReposDirOverride(t *testing.T) {
	ws := t.TempDir()
	base := t.TempDir() // a predefined, writable area outside the workspace
	t.Setenv(storage.ReposDirEnv, base)

	r, err := NewLocalRegistry(ws).CreateFilesystem("acme", "Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	want := filepath.Join(base, "acme")
	if r.Location != want {
		t.Errorf("Location = %q, want %q", r.Location, want)
	}
	if _, err := os.Stat(filepath.Join(want, ".nomos", "cosmos.yaml")); err != nil {
		t.Errorf("cosmos.yaml not created under override dir: %v", err)
	}
	// Stored as absolute, so List from the workspace still resolves it.
	if !containsID(NewLocalRegistry(ws).List(), "acme") {
		t.Errorf("override repository not found in List")
	}
}

func containsID(repos []Repository, id string) bool {
	for _, r := range repos {
		if r.ID == id {
			return true
		}
	}
	return false
}
