// Package repo enumerates and manages the git-first repositories a Nomos server
// exposes (ADR-0022 §2/§3). The local server always exposes a default
// repository (the active workspace) plus any repositories listed in
// .nomos/repositories.yaml. Filesystem repositories can be created; remote
// (github/gitbucket) registration is a later step.
package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nomos/nomos/internal/storage"
	"gopkg.in/yaml.v3"
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

var repoID = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Registry enumerates and mutates the repositories a server manages.
type Registry struct {
	workspace string
}

// NewLocalRegistry returns a registry for the local server backed by workspace.
func NewLocalRegistry(workspace string) *Registry { return &Registry{workspace: workspace} }

type configEntry struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name,omitempty"`
	Kind     Kind   `yaml:"kind"`
	Location string `yaml:"location"`
}

type reposConfig struct {
	Repositories []configEntry `yaml:"repositories"`
}

func (r *Registry) load() (reposConfig, error) {
	var cfg reposConfig
	data, err := os.ReadFile(storage.RepositoriesFile(r.workspace))
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	return cfg, yaml.Unmarshal(data, &cfg)
}

func (r *Registry) save(cfg reposConfig) error {
	if err := os.MkdirAll(storage.NomosDir(r.workspace), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(storage.RepositoriesFile(r.workspace), data, 0o644)
}

// List returns the default repository (the active workspace) followed by the
// configured additional repositories.
func (r *Registry) List() []Repository {
	out := []Repository{r.localDefault()}
	cfg, err := r.load()
	if err != nil {
		return out
	}
	for _, e := range cfg.Repositories {
		repo := Repository{ID: e.ID, Name: e.Name, Kind: e.Kind, Location: r.resolve(e.Location), Status: "unknown"}
		probeGit(&repo)
		out = append(out, repo)
	}
	return out
}

// CreateFilesystem creates a new local git-first repository workspace and
// records it in the config. The id is derived from the name when not given.
func (r *Registry) CreateFilesystem(id, name string) (Repository, error) {
	if id == "" {
		id = slug(name)
	}
	if !repoID.MatchString(id) {
		return Repository{}, fmt.Errorf("invalid repository id: %q", id)
	}
	if id == LocalDefaultID {
		return Repository{}, fmt.Errorf("id %q is reserved", LocalDefaultID)
	}
	cfg, err := r.load()
	if err != nil {
		return Repository{}, err
	}
	for _, e := range cfg.Repositories {
		if e.ID == id {
			return Repository{}, fmt.Errorf("repository %q already exists", id)
		}
	}
	loc := filepath.Join(storage.ReposDir(r.workspace), id)
	if _, err := os.Stat(loc); err == nil {
		return Repository{}, fmt.Errorf("repository directory already exists: %s", loc)
	}
	if err := os.MkdirAll(filepath.Join(storage.NomosDir(loc), "domains"), 0o755); err != nil {
		return Repository{}, err
	}
	displayName := name
	if displayName == "" {
		displayName = id
	}
	cosmosYAML := fmt.Sprintf("id: %s\nname: %s\nversion: 0.1.0\nstatus: draft\nowner: unknown\n", id, displayName)
	if err := os.WriteFile(storage.CosmosFile(loc), []byte(cosmosYAML), 0o644); err != nil {
		return Repository{}, err
	}
	// Store a workspace-relative location so repositories.yaml stays portable.
	rel := filepath.ToSlash(filepath.Join(".nomos", "repos", id))
	cfg.Repositories = append(cfg.Repositories, configEntry{ID: id, Name: displayName, Kind: KindFilesystem, Location: rel})
	if err := r.save(cfg); err != nil {
		return Repository{}, err
	}
	repo := Repository{ID: id, Name: displayName, Kind: KindFilesystem, Location: loc, Status: "unknown"}
	probeGit(&repo)
	return repo, nil
}

// resolve turns a possibly-relative configured location into an absolute path.
func (r *Registry) resolve(location string) string {
	if filepath.IsAbs(location) {
		return location
	}
	return filepath.Join(r.workspace, location)
}

func (r *Registry) localDefault() Repository {
	repo := Repository{ID: LocalDefaultID, Kind: KindFilesystem, Location: r.workspace, Status: "unknown"}
	probeGit(&repo)
	return repo
}

func slug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
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
