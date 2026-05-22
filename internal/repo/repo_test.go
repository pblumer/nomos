package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/idgen"
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

func TestCreateFilesystemDefaultLocationIsOutsideWorkspace(t *testing.T) {
	ws := t.TempDir()
	home := t.TempDir() // isolate ~/.nomos-repos from the developer's real home
	t.Setenv("HOME", home)
	r, err := NewLocalRegistry(ws).CreateFilesystem("Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	// The id is system-generated per ADR-0020 (REP_ prefix), not slug-derived.
	if !idgen.IsValidForType(r.ID, "repository") {
		t.Errorf("ID = %q, want a system-generated REP_ id", r.ID)
	}
	want := filepath.Join(home, storage.DefaultReposDirName, r.ID)
	if r.Location != want {
		t.Errorf("Location = %q, want %q", r.Location, want)
	}
	// The repo must reappear (resolved) on a fresh List from the same workspace.
	if !containsID(NewLocalRegistry(ws).List(), r.ID) {
		t.Errorf("created repository not found in List")
	}
	// New repositories are git-first: a git database is initialized on creation.
	if _, err := os.Stat(filepath.Join(want, ".git")); err != nil {
		t.Errorf("expected git repository to be initialized: %v", err)
	}
	// Only the Decision starter type is seeded under .nomos/types.
	if _, err := os.Stat(filepath.Join(storage.TypesDir(want), "decision.yaml")); err != nil {
		t.Errorf("expected seeded type definition %q: %v", "decision", err)
	}
	for _, id := range []string{"service", "blueprint", "product"} {
		if _, err := os.Stat(filepath.Join(storage.TypesDir(want), id+".yaml")); !os.IsNotExist(err) {
			t.Errorf("expected no seeded type definition %q, stat err = %v", id, err)
		}
	}
	// The starter "capture a type" form is seeded under .nomos/views.
	if _, err := os.Stat(filepath.Join(storage.ViewsDir(want), "type_new.frm")); err != nil {
		t.Errorf("expected seeded type-capture view: %v", err)
	}
	// The Decision starter type ships with its four standard view forms.
	for _, v := range []string{"new", "edit", "list", "short"} {
		if _, err := os.Stat(filepath.Join(storage.ViewsDir(want), "decision_"+v+".frm")); err != nil {
			t.Errorf("expected seeded decision_%s.frm: %v", v, err)
		}
	}
	// The obsolete .nomos/domains directory is no longer scaffolded.
	if _, err := os.Stat(filepath.Join(storage.NomosDir(want), "domains")); !os.IsNotExist(err) {
		t.Errorf("expected no .nomos/domains directory, stat err = %v", err)
	}
}

func TestCreateFilesystemHonorsReposDirOverride(t *testing.T) {
	ws := t.TempDir()
	base := t.TempDir() // a predefined, writable area outside the workspace
	t.Setenv(storage.ReposDirEnv, base)

	r, err := NewLocalRegistry(ws).CreateFilesystem("Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	want := filepath.Join(base, r.ID)
	if r.Location != want {
		t.Errorf("Location = %q, want %q", r.Location, want)
	}
	if _, err := os.Stat(filepath.Join(want, ".nomos", "cosmos.yaml")); err != nil {
		t.Errorf("cosmos.yaml not created under override dir: %v", err)
	}
	// Stored as absolute, so List from the workspace still resolves it.
	if !containsID(NewLocalRegistry(ws).List(), r.ID) {
		t.Errorf("override repository not found in List")
	}
}

func TestAttachRegistersExistingDirectoryWithoutScaffolding(t *testing.T) {
	ws := t.TempDir()
	existing := t.TempDir() // a pre-existing directory outside the workspace

	r, err := NewLocalRegistry(ws).Attach(existing, "Imported")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if !idgen.IsValidForType(r.ID, "repository") {
		t.Errorf("ID = %q, want a system-generated REP_ id", r.ID)
	}
	if r.Name != "Imported" {
		t.Errorf("Name = %q, want Imported", r.Name)
	}
	if r.Location != existing {
		t.Errorf("Location = %q, want %q", r.Location, existing)
	}
	// Attaching must not scaffold a fresh workspace: no cosmos.yaml is written.
	if _, err := os.Stat(filepath.Join(storage.NomosDir(existing), "cosmos.yaml")); !os.IsNotExist(err) {
		t.Errorf("Attach scaffolded cosmos.yaml; stat err = %v", err)
	}
	// The attached repository must reappear on a fresh List from the workspace.
	if !containsID(NewLocalRegistry(ws).List(), r.ID) {
		t.Errorf("attached repository not found in List")
	}
}

func TestAttachDefaultsNameToBaseAndRejectsBadInput(t *testing.T) {
	ws := t.TempDir()
	existing := filepath.Join(t.TempDir(), "team-beta")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	reg := NewLocalRegistry(ws)

	r, err := reg.Attach(existing, "")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	if r.Name != "team-beta" {
		t.Errorf("Name = %q, want directory base name team-beta", r.Name)
	}
	// Re-attaching the same directory is a duplicate and must be rejected.
	if _, err := reg.Attach(existing, ""); err == nil {
		t.Error("expected error re-attaching an already attached directory")
	}
	// The workspace itself is the default repository and cannot be attached.
	if _, err := reg.Attach(ws, ""); err == nil {
		t.Error("expected error attaching the workspace")
	}
	// A non-existent path is rejected.
	if _, err := reg.Attach(filepath.Join(existing, "missing"), ""); err == nil {
		t.Error("expected error attaching a non-existent path")
	}
	// An empty location is rejected.
	if _, err := reg.Attach("", ""); err == nil {
		t.Error("expected error attaching an empty location")
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
