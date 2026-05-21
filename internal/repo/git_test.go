package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

// newRepoWithFile creates a fresh filesystem repository and returns its
// location. It pins a local git identity with signing disabled so commits are
// reproducible regardless of the host's global git configuration.
func newRepoWithFile(t *testing.T) string {
	t.Helper()
	t.Setenv(storage.ReposDirEnv, t.TempDir()) // keep repos out of the real ~/.nomos-repos
	ws := t.TempDir()
	r, err := NewLocalRegistry(ws).CreateFilesystem("Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	for _, kv := range [][2]string{
		{"user.name", "Test"},
		{"user.email", "test@example.com"},
		{"commit.gpgsign", "false"},
		{"tag.gpgsign", "false"},
	} {
		if err := exec.Command("git", "-C", r.Location, "config", kv[0], kv[1]).Run(); err != nil {
			t.Fatalf("git config %s: %v", kv[0], err)
		}
	}
	return r.Location
}

func TestStatusReportsDirtyThenCleanAfterCommit(t *testing.T) {
	loc := newRepoWithFile(t)

	st, err := Status(loc)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !st.Initialized {
		t.Fatal("expected initialized git repository")
	}
	if !st.Dirty || len(st.Files) == 0 {
		t.Fatalf("expected dirty working tree with files, got %+v", st)
	}

	if err := Commit(loc, "initial"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	st, err = Status(loc)
	if err != nil {
		t.Fatalf("Status after commit: %v", err)
	}
	if st.Dirty {
		t.Errorf("expected clean working tree after commit, got %+v", st)
	}
	if st.Head == "" {
		t.Error("expected a HEAD after commit")
	}
}

func TestCommitRejectsEmptyMessageAndCleanTree(t *testing.T) {
	loc := newRepoWithFile(t)
	if err := Commit(loc, "  "); err == nil {
		t.Error("expected error for empty commit message")
	}
	if err := Commit(loc, "first"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := Commit(loc, "second"); err == nil {
		t.Error("expected error committing a clean tree")
	}
}

func TestBranchLifecycle(t *testing.T) {
	loc := newRepoWithFile(t)
	if err := Commit(loc, "initial"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := CreateBranch(loc, "feature-x", true); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	current, all, err := Branches(loc)
	if err != nil {
		t.Fatalf("Branches: %v", err)
	}
	if current != "feature-x" {
		t.Errorf("current branch = %q, want feature-x", current)
	}
	if len(all) < 2 {
		t.Errorf("expected at least 2 branches, got %v", all)
	}
}

func TestCreateBranchRejectsInvalidName(t *testing.T) {
	loc := newRepoWithFile(t)
	if err := Commit(loc, "initial"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	for _, bad := range []string{"", "-x", "a b", "a..b", "--force"} {
		if err := CreateBranch(loc, bad, false); err == nil {
			t.Errorf("expected error for invalid branch name %q", bad)
		}
	}
}

func TestTagLifecycle(t *testing.T) {
	loc := newRepoWithFile(t)
	if err := CreateTag(loc, "v0.1.0", "first"); err == nil {
		t.Error("expected error tagging a repository with no commits")
	}
	if err := Commit(loc, "initial"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := CreateTag(loc, "v0.1.0", "first release"); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	tags, err := Tags(loc)
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "v0.1.0" || tags[0].Message != "first release" {
		t.Errorf("unexpected tags: %+v", tags)
	}
}

func TestDeleteRemovesRepositoryAndDirectory(t *testing.T) {
	t.Setenv(storage.ReposDirEnv, t.TempDir())
	ws := t.TempDir()
	reg := NewLocalRegistry(ws)
	r, err := reg.CreateFilesystem("Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	if err := reg.Delete(r.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(r.Location); !os.IsNotExist(err) {
		t.Errorf("expected repository directory removed, stat err = %v", err)
	}
	if containsID(NewLocalRegistry(ws).List(), r.ID) {
		t.Error("deleted repository still listed")
	}
}

func TestDeleteRefusesDefaultAndUnknown(t *testing.T) {
	ws := t.TempDir()
	reg := NewLocalRegistry(ws)
	if err := reg.Delete(LocalDefaultID); err == nil {
		t.Error("expected error deleting the default repository")
	}
	if err := reg.Delete("nope"); err == nil {
		t.Error("expected error deleting an unknown repository")
	}
}

func TestRenameUpdatesDisplayName(t *testing.T) {
	t.Setenv(storage.ReposDirEnv, t.TempDir())
	ws := t.TempDir()
	reg := NewLocalRegistry(ws)
	created, err := reg.CreateFilesystem("Acme")
	if err != nil {
		t.Fatalf("CreateFilesystem: %v", err)
	}
	r, err := reg.Rename(created.ID, "Acme Corp")
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if r.Name != "Acme Corp" {
		t.Errorf("Name = %q, want Acme Corp", r.Name)
	}
	for _, e := range NewLocalRegistry(ws).List() {
		if e.ID == created.ID && e.Name != "Acme Corp" {
			t.Errorf("persisted name = %q, want Acme Corp", e.Name)
		}
	}
	if _, err := reg.Rename(LocalDefaultID, "x"); err == nil {
		t.Error("expected error renaming the default repository")
	}
}

func TestStatusUninitializedDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := Status(dir)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Initialized {
		t.Error("expected Initialized=false for a non-git directory")
	}
}
