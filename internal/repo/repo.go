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
	"strings"

	"github.com/nomos/nomos/internal/idgen"
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
// records it in the config. The id is system-generated (ADR-0020, REP_ prefix);
// callers supply only the display name.
func (r *Registry) CreateFilesystem(name string) (Repository, error) {
	cfg, err := r.load()
	if err != nil {
		return Repository{}, err
	}
	id, err := r.newRepoID(cfg)
	if err != nil {
		return Repository{}, err
	}
	loc := filepath.Join(storage.ReposDir(r.workspace), id)
	if _, err := os.Stat(loc); err == nil {
		return Repository{}, fmt.Errorf("repository directory already exists: %s", loc)
	}
	// Seed the views directory with the starter "capture a type" form (ADR-0024)
	// so a freshly created repository ships a data-entry view. This MkdirAll also
	// materialises .nomos itself, which cosmos.yaml and the type seeds rely on.
	if err := SeedDefaultViews(loc); err != nil {
		return Repository{}, err
	}
	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = id
	}
	// The repository's cosmos artifact carries the repo-local opaque cosmos id
	// (ADR-0020 §Cosmos-Identität), not the topology-level REP_ id.
	cosmosYAML := fmt.Sprintf("id: %s\nname: %s\nversion: 0.1.0\nstatus: draft\nowner: unknown\n", idgen.CosmosRootID, displayName)
	if err := os.WriteFile(storage.CosmosFile(loc), []byte(cosmosYAML), 0o644); err != nil {
		return Repository{}, err
	}
	// Seed starter type definitions so .nomos/types is populated on creation
	// (best-effort: the user may delete or adapt them). The directory is
	// system-relevant, so a failure here does not block repository creation.
	_ = seedDefaultTypes(loc)
	// Initialize git so the workspace is git-first from creation (best-effort:
	// a missing git binary must not block filesystem repository creation).
	_ = exec.Command("git", "-C", loc, "init", "-q").Run()
	// Prefer a workspace-relative location so repositories.yaml stays portable;
	// fall back to an absolute path when the repos base lives outside the
	// workspace (e.g. NOMOS_REPOS_DIR points elsewhere).
	location := loc
	if rel, err := filepath.Rel(r.workspace, loc); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		location = filepath.ToSlash(rel)
	}
	cfg.Repositories = append(cfg.Repositories, configEntry{ID: id, Name: displayName, Kind: KindFilesystem, Location: location})
	if err := r.save(cfg); err != nil {
		return Repository{}, err
	}
	repo := Repository{ID: id, Name: displayName, Kind: KindFilesystem, Location: loc, Status: "unknown"}
	probeGit(&repo)
	return repo, nil
}

// Attach registers an existing directory as a repository this server manages,
// without scaffolding any files (no cosmos.yaml, types, views, or git init). It
// is the "open an existing repository" counterpart to CreateFilesystem: the
// caller supplies a path that already holds a git-first workspace and an
// optional display name. The id is system-generated (ADR-0020, REP_ prefix).
func (r *Registry) Attach(location, name string) (Repository, error) {
	location = strings.TrimSpace(location)
	if location == "" {
		return Repository{}, fmt.Errorf("location is required")
	}
	abs := r.resolve(location)
	info, err := os.Stat(abs)
	if err != nil {
		return Repository{}, fmt.Errorf("repository location not found: %s", abs)
	}
	if !info.IsDir() {
		return Repository{}, fmt.Errorf("repository location is not a directory: %s", abs)
	}
	cfg, err := r.load()
	if err != nil {
		return Repository{}, err
	}
	// The workspace is already exposed as the default repository; attaching it
	// again would create a duplicate, so reject it.
	if filepath.Clean(abs) == filepath.Clean(r.workspace) {
		return Repository{}, fmt.Errorf("this directory is already the default repository")
	}
	for _, e := range cfg.Repositories {
		if filepath.Clean(r.resolve(e.Location)) == filepath.Clean(abs) {
			return Repository{}, fmt.Errorf("repository already attached: %s", abs)
		}
	}
	id, err := r.newRepoID(cfg)
	if err != nil {
		return Repository{}, err
	}
	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = filepath.Base(abs)
	}
	// Prefer a workspace-relative location so repositories.yaml stays portable;
	// fall back to the absolute path when the directory lives outside the
	// workspace (mirrors CreateFilesystem).
	stored := abs
	if rel, err := filepath.Rel(r.workspace, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		stored = filepath.ToSlash(rel)
	}
	cfg.Repositories = append(cfg.Repositories, configEntry{ID: id, Name: displayName, Kind: KindFilesystem, Location: stored})
	if err := r.save(cfg); err != nil {
		return Repository{}, err
	}
	repo := Repository{ID: id, Name: displayName, Kind: KindFilesystem, Location: abs, Status: "unknown"}
	probeGit(&repo)
	return repo, nil
}

// Delete removes a configured filesystem repository: it drops the entry from
// repositories.yaml and deletes its working directory. The default repository
// (the active workspace) cannot be deleted.
func (r *Registry) Delete(id string) error {
	if id == LocalDefaultID {
		return fmt.Errorf("the default repository cannot be deleted")
	}
	cfg, err := r.load()
	if err != nil {
		return err
	}
	idx := -1
	for i, e := range cfg.Repositories {
		if e.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("repository %q not found", id)
	}
	loc := r.resolve(cfg.Repositories[idx].Location)
	// Never delete the workspace itself, even if it was somehow registered.
	if filepath.Clean(loc) != filepath.Clean(r.workspace) {
		if err := os.RemoveAll(loc); err != nil {
			return fmt.Errorf("failed to remove repository directory: %w", err)
		}
	}
	cfg.Repositories = append(cfg.Repositories[:idx], cfg.Repositories[idx+1:]...)
	return r.save(cfg)
}

// Rename updates a repository's display name. The id and on-disk location stay
// stable so existing references keep resolving. The default repository's name
// lives in its cosmos.yaml and is not editable here.
func (r *Registry) Rename(id, name string) (Repository, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Repository{}, fmt.Errorf("name is required")
	}
	if id == LocalDefaultID {
		return Repository{}, fmt.Errorf("the default repository is renamed via its cosmos.yaml")
	}
	cfg, err := r.load()
	if err != nil {
		return Repository{}, err
	}
	for i, e := range cfg.Repositories {
		if e.ID == id {
			cfg.Repositories[i].Name = name
			if err := r.save(cfg); err != nil {
				return Repository{}, err
			}
			repo := Repository{ID: id, Name: name, Kind: e.Kind, Location: r.resolve(e.Location), Status: "unknown"}
			probeGit(&repo)
			return repo, nil
		}
	}
	return Repository{}, fmt.Errorf("repository %q not found", id)
}

// seedDefaultTypes writes the built-in artifact types as editable definitions
// under .nomos/types so a freshly created repository ships a starter set the
// user can extend, adapt, or delete.
func seedDefaultTypes(loc string) error {
	dir := storage.TypesDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, def := range defaultTypeDefs() {
		b, err := yaml.Marshal(def)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, def.ID+".yaml"), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// SeedDefaultViews writes the starter "capture a type" form under .nomos/views
// so a freshly created repository ships a data-entry view (ADR-0024). The file
// follows the <typeID>_new.frm convention; here the captured type is "type",
// i.e. the form that records a new type definition. Exported so the cosmos-init
// CLI path seeds the same view as API-created repositories.
func SeedDefaultViews(loc string) error {
	dir := storage.ViewsDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(defaultTypeCaptureView())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "type_new.frm"), b, 0o644)
}

// viewSeed mirrors model.View's YAML shape; defined locally to keep the repo
// package free of an app/model dependency for a one-off scaffold.
type viewSeed struct {
	Engine        string         `yaml:"engine,omitempty"`
	EngineVersion string         `yaml:"engine_version,omitempty"`
	Schema        map[string]any `yaml:"schema,omitempty"`
}

// defaultTypeCaptureView is the form-js view that records a new type definition.
// It is laid out for the form-js renderer (ADR-0024): an intro text, an identity
// group (id/label/description) and a rendering group (icon/viewer/editor), with
// the id constrained to the lowercase-slug shape type ids must follow.
func defaultTypeCaptureView() viewSeed {
	return viewSeed{
		Engine:        "form-js",
		EngineVersion: "1",
		Schema: map[string]any{
			"type": "default",
			"components": []map[string]any{
				{
					"type": "text",
					"text": "## Neuen Typ erfassen\n\nLegt eine neue Typdefinition unter `.nomos/types` an.",
				},
				{
					"type":        "group",
					"label":       "Identität",
					"showOutline": true,
					"components": []map[string]any{
						{
							"type":        "textfield",
							"key":         "id",
							"label":       "ID",
							"description": "Kleinbuchstaben-Slug: a–z, 0–9, - und _ (z. B. \"data-object\").",
							"validate":    map[string]any{"required": true, "pattern": "^[a-z0-9_-]+$"},
						},
						{"type": "textfield", "key": "label", "label": "Label", "description": "Anzeigename im Explorer."},
						{"type": "textfield", "key": "file", "label": "Erkennungsdatei", "description": "Markiert ein Verzeichnis als Instanz dieses Typs (z. B. \"requirement.yaml\")."},
						{"type": "textfield", "key": "id_prefix", "label": "ID-Präfix", "description": "Optional: drei Großbuchstaben für automatische Instanz-IDs (z. B. \"RSK\" → RSK_4F7K2Q).", "validate": map[string]any{"pattern": "^[A-Z]{0,3}$"}},
						{"type": "textarea", "key": "description", "label": "Beschreibung"},
					},
				},
				{
					"type":        "group",
					"label":       "Darstellung",
					"showOutline": true,
					"components": []map[string]any{
						{"type": "textfield", "key": "icon", "label": "Icon", "description": "Material-Icon-Name, z. B. \"settings\"."},
						{
							"type":  "select",
							"key":   "viewer",
							"label": "Viewer",
							"values": []map[string]any{
								{"label": "Formular", "value": "form"},
								{"label": "BPMN", "value": "bpmn"},
								{"label": "DMN", "value": "dmn"},
							},
						},
						{
							"type":  "select",
							"key":   "editor",
							"label": "Editor",
							"values": []map[string]any{
								{"label": "Formular", "value": "form"},
								{"label": "DMN", "value": "dmn"},
							},
						},
					},
				},
				{
					"type":               "dynamiclist",
					"path":               "properties",
					"label":              "Properties",
					"showOutline":        true,
					"isRepeating":        true,
					"allowAddRemove":     true,
					"defaultRepetitions": 0,
					"components": []map[string]any{
						{"type": "textfield", "key": "name", "label": "Feldname", "validate": map[string]any{"required": true}},
						{
							"type":  "select",
							"key":   "type",
							"label": "Typ",
							"values": []map[string]any{
								{"label": "string", "value": "string"},
								{"label": "number", "value": "number"},
								{"label": "boolean", "value": "boolean"},
								{"label": "text", "value": "text"},
							},
						},
						{"type": "checkbox", "key": "required", "label": "Pflicht"},
					},
				},
			},
		},
	}
}

// typeDefSeed mirrors model.TypeDef's YAML shape; defined locally to keep the
// repo package free of an app/model dependency for a one-off scaffold.
type typeDefSeed struct {
	ID           string           `yaml:"id"`
	Label        string           `yaml:"label,omitempty"`
	Description  string           `yaml:"description,omitempty"`
	Icon         string           `yaml:"icon,omitempty"`
	Viewer       string           `yaml:"viewer,omitempty"`
	Editor       string           `yaml:"editor,omitempty"`
	File         string           `yaml:"file,omitempty"`
	Properties   []typeProp       `yaml:"properties,omitempty"`
	Dependencies []typeDependency `yaml:"dependencies,omitempty"`
}
type typeProp struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}
type typeDependency struct {
	Type     string `yaml:"type"`
	Relation string `yaml:"relation,omitempty"`
}

func defaultTypeDefs() []typeDefSeed {
	return []typeDefSeed{
		{
			ID: "service", Label: "Service", Icon: "settings", Viewer: "bpmn", Editor: "form", File: "service.yaml",
			Description: "A capability-providing service.",
			Properties: []typeProp{
				{Name: "name", Type: "string", Required: true},
				{Name: "owner", Type: "string"},
				{Name: "capabilities", Type: "list"},
			},
			Dependencies: []typeDependency{{Type: "decision", Relation: "uses"}},
		},
		{
			ID: "decision", Label: "Decision", Icon: "gavel", Viewer: "dmn", Editor: "dmn", File: "decision.yaml",
			Description: "Decision logic, optionally backed by a DMN table.",
			Properties: []typeProp{
				{Name: "name", Type: "string", Required: true},
				{Name: "inputs", Type: "list"},
				{Name: "outputs", Type: "list"},
			},
		},
		{
			ID: "blueprint", Label: "Blueprint", Icon: "assignment", Viewer: "form", Editor: "form",
			Description: "An abstract product or service definition in the catalog.",
			Properties: []typeProp{
				{Name: "name", Type: "string", Required: true},
				{Name: "fulfillment", Type: "object"},
			},
		},
		{
			ID: "product", Label: "Product", Icon: "inventory_2", Viewer: "form", Editor: "form",
			Description: "A concrete product offering composed of services.",
			Properties: []typeProp{
				{Name: "name", Type: "string", Required: true},
			},
			Dependencies: []typeDependency{{Type: "service", Relation: "fulfilled_by"}},
		},
	}
}

// newRepoID returns a fresh system-assigned repository id (ADR-0020, REP_
// prefix). It regenerates on the astronomically rare collision with an existing
// configured repository.
func (r *Registry) newRepoID(cfg reposConfig) (string, error) {
	existing := make(map[string]bool, len(cfg.Repositories))
	for _, e := range cfg.Repositories {
		existing[e.ID] = true
	}
	for attempt := 0; attempt < 5; attempt++ {
		id, err := idgen.NewForType("repository")
		if err != nil {
			return "", err
		}
		if !existing[id] {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not allocate a unique repository id")
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
