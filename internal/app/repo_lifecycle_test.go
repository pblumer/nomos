package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDeleteRepoNode(t *testing.T) {
	p := createAppTestCosmos(t)
	rel := ".nomos/services/user-account"
	if err := DeleteRepoNode(p, rel); err != nil {
		t.Fatalf("DeleteRepoNode: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Errorf("expected folder removed, stat err = %v", err)
	}
	if err := DeleteRepoNode(p, "does/not/exist"); err == nil {
		t.Error("expected error deleting a missing path")
	}
	if err := DeleteRepoNode(p, "../escape"); err == nil {
		t.Error("expected path-traversal to be rejected")
	}
}

func TestRenameRepoNode(t *testing.T) {
	p := createAppTestCosmos(t)
	if err := RenameRepoNode(p, ".nomos/services/user-account", "user-acct"); err != nil {
		t.Fatalf("RenameRepoNode: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, ".nomos", "services", "user-acct")); err != nil {
		t.Errorf("renamed folder not found: %v", err)
	}
	if err := RenameRepoNode(p, ".nomos/services/user-acct", "a/b"); err == nil {
		t.Error("expected error for a name containing a path separator")
	}
	// Renaming onto an existing sibling is a conflict.
	if err := RenameRepoNode(p, ".nomos/services/user-acct", "privileged-account"); err == nil {
		t.Error("expected conflict renaming onto an existing sibling")
	}
}

// gitInit pins a signing-free identity so commits work regardless of host config.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.name", "Test"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgsign", "false"},
	} {
		if err := exec.Command("git", append([]string{"-C", dir}, args...)...).Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
}

func TestGitStatusAndCommitOnDefaultRepo(t *testing.T) {
	p := createAppTestCosmos(t)
	gitInit(t, p)

	st, err := GitStatus(p, "default")
	if err != nil {
		t.Fatalf("GitStatus: %v", err)
	}
	if !st.Initialized || !st.Dirty {
		t.Fatalf("expected an initialized, dirty repo, got %+v", st)
	}

	st, err = CommitRepo(p, "default", "initial import")
	if err != nil {
		t.Fatalf("CommitRepo: %v", err)
	}
	if st.Dirty {
		t.Errorf("expected clean tree after commit, got %+v", st)
	}

	br, err := CreateBranch(p, "default", "feature-y", true)
	if err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	if br.Current != "feature-y" {
		t.Errorf("current branch = %q, want feature-y", br.Current)
	}

	tags, err := CreateTag(p, "default", "v0.1.0", "first")
	if err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if len(tags.Tags) != 1 || tags.Tags[0].Name != "v0.1.0" {
		t.Errorf("unexpected tags: %+v", tags.Tags)
	}
}

func TestGitOpsUnknownRepo(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := GitStatus(p, "ghost"); err == nil {
		t.Error("expected error for an unknown repository")
	}
}

func TestCreateAndDeleteRepository(t *testing.T) {
	p := createAppTestCosmos(t)
	r, err := CreateRepository(p, "Team Beta")
	if err != nil {
		t.Fatalf("CreateRepository: %v", err)
	}
	if !containsRepo(t, p, r.ID) {
		t.Fatal("created repository not listed")
	}
	renamed, err := RenameRepository(p, r.ID, "Team Gamma")
	if err != nil {
		t.Fatalf("RenameRepository: %v", err)
	}
	if renamed.Name != "Team Gamma" {
		t.Errorf("Name = %q, want Team Gamma", renamed.Name)
	}
	if err := DeleteRepository(p, r.ID); err != nil {
		t.Fatalf("DeleteRepository: %v", err)
	}
	if containsRepo(t, p, r.ID) {
		t.Error("deleted repository still listed")
	}
	if err := DeleteRepository(p, "default"); err == nil {
		t.Error("expected error deleting the default repository")
	}
}

func containsRepo(t *testing.T, path, id string) bool {
	t.Helper()
	repos, err := ListRepositories(path)
	if err != nil {
		t.Fatalf("ListRepositories: %v", err)
	}
	for _, r := range repos.Repositories {
		if r.ID == id {
			return true
		}
	}
	return false
}
