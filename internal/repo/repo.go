// Package repo enumerates the git-first repositories a Nomos server manages
// (ADR-0022 §2/§3). PR 1 implements the local server only: a single default
// repository pointing at the active workspace (ADR-0022 §6).
package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Kind classifies where a repository's git-first source lives (ADR-0022 §3).
type Kind string

const (
	KindFilesystem Kind = "filesystem"
	KindGitHub     Kind = "github"
	KindGitBucket  Kind = "gitbucket"
)

// Repository describes one git-first repository exposed by a server (ADR-0022 §2).
type Repository struct {
	ID            string
	Name          string
	Kind          Kind
	Location      string
	DefaultBranch string
	Status        string // clean | dirty | unreachable | unknown
	Head          string
}

// LocalDefaultID is the identifier of the local server's default repository.
const LocalDefaultID = "default"

// Registry enumerates the repositories a server manages.
type Registry struct {
	workspace string
}

// NewLocalRegistry returns a registry for the local server backed by workspace.
func NewLocalRegistry(workspace string) *Registry { return &Registry{workspace: workspace} }

// List returns the repositories managed by the local server. PR 1 returns the
// single default repository for the active workspace.
func (r *Registry) List() []Repository {
	return []Repository{r.localDefault()}
}

func (r *Registry) localDefault() Repository {
	repo := Repository{
		ID:       LocalDefaultID,
		Kind:     KindFilesystem,
		Location: r.workspace,
		Status:   "unknown",
	}
	probeGit(&repo)
	return repo
}

// probeGit fills Head/Status/DefaultBranch from git, best-effort. A workspace
// without a git repository keeps Status "unknown".
func probeGit(repo *Repository) {
	if _, err := os.Stat(filepath.Join(repo.Location, ".git")); err != nil {
		return
	}
	if head := gitOutput(repo.Location, "rev-parse", "--short", "HEAD"); head != "" {
		repo.Head = head
	}
	if branch := gitOutput(repo.Location, "rev-parse", "--abbrev-ref", "HEAD"); branch != "" {
		repo.DefaultBranch = branch
	}
	if porcelain := gitOutput(repo.Location, "status", "--porcelain"); porcelain == "" {
		repo.Status = "clean"
	} else {
		repo.Status = "dirty"
	}
}

func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
