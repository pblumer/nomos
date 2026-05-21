package app

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/storage"
)

// BuildRepoTree mirrors a repository's physical working directory 1:1 as a tree
// of folder and file nodes (ADR: below the repository, the git/filesystem layout
// is the source of truth; the REST API stays flat). Directories that hold a
// recognized artifact (service.yaml, decision.yaml) collapse into a single typed
// leaf so their internal layout is not exposed, and recognized catalog files
// (blueprints) become typed file nodes so their detail views keep working.
// Everything else is shown verbatim: directories as folders, files as files.
//
// Pure plumbing that is not authored content is hidden: the git database (.git)
// and the derived .nomos/{repos,cache,index} directories.
func BuildRepoTree(loc string) []NamespaceTreeNodeDTO {
	loc = filepath.Clean(loc)
	tree, err := cosmosfs.LoadTree(loc)
	if err != nil {
		// No cosmos workspace yet (freshly created repo): mirror whatever is on
		// disk without artifact recognition.
		tree = cosmosfs.Tree{}
	}

	services := map[string]cosmosfs.ServiceNode{}
	for _, s := range tree.Services {
		services[filepath.Clean(s.Path)] = s
	}
	decisions := map[string]cosmosfs.DecisionNode{}
	for _, d := range tree.Decisions {
		decisions[filepath.Clean(d.Path)] = d
	}
	blueprints := map[string]cosmosfs.BlueprintNode{}
	for _, b := range tree.Blueprints {
		blueprints[filepath.Clean(b.Path)] = b
	}

	skip := map[string]bool{
		filepath.Clean(storage.ReposDir(loc)): true,
		filepath.Clean(storage.CacheDir(loc)): true,
		filepath.Clean(storage.IndexDir(loc)): true,
	}

	m := &repoMirror{services: services, decisions: decisions, blueprints: blueprints, skip: skip}
	return m.children(loc, "")
}

type repoMirror struct {
	services   map[string]cosmosfs.ServiceNode
	decisions  map[string]cosmosfs.DecisionNode
	blueprints map[string]cosmosfs.BlueprintNode
	skip       map[string]bool
}

func (m *repoMirror) children(absDir, relDir string) []NamespaceTreeNodeDTO {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil
	}
	// Folders first, then files; alphabetical within each group.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	out := []NamespaceTreeNodeDTO{}
	for _, e := range entries {
		name := e.Name()
		abs := filepath.Clean(filepath.Join(absDir, name))
		rel := filepath.Join(relDir, name)

		if e.IsDir() {
			if name == ".git" || m.skip[abs] {
				continue
			}
			if sn, ok := m.services[abs]; ok {
				out = append(out, m.serviceNode(sn, rel))
				continue
			}
			if dn, ok := m.decisions[abs]; ok {
				out = append(out, m.decisionNode(dn, rel))
				continue
			}
			out = append(out, NamespaceTreeNodeDTO{
				Label:          name,
				Kind:           "folder",
				IsFolder:       true,
				GitPath:        rel,
				TreePath:       rel,
				DisplayPath:    rel,
				Persisted:      true,
				CanOpenDetails: true,
				Children:       m.children(abs, rel),
			})
			continue
		}

		if bn, ok := m.blueprints[abs]; ok {
			out = append(out, NamespaceTreeNodeDTO{
				Label:          firstNonEmpty(bn.Metadata.Name, bn.Metadata.ID, name),
				Kind:           "blueprint",
				Canonical:      bn.Metadata.ID,
				GitPath:        rel,
				TreePath:       rel,
				DisplayPath:    rel,
				Persisted:      true,
				CanOpenDetails: true,
			})
			continue
		}
		out = append(out, NamespaceTreeNodeDTO{
			Label:          name,
			Kind:           "file",
			GitPath:        rel,
			TreePath:       rel,
			DisplayPath:    rel,
			Persisted:      true,
			CanOpenDetails: true,
		})
	}
	return out
}

// safeRepoPath resolves a repository-relative path against the workspace root,
// rejecting empty input and anything that escapes the workspace (path
// traversal) or targets git internals.
func safeRepoPath(loc, rel string) (string, error) {
	rel = strings.TrimSpace(filepath.ToSlash(rel))
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." {
		return "", Error(CodeInvalidInput, "path is required", http.StatusBadRequest, nil)
	}
	for _, seg := range strings.Split(rel, "/") {
		switch {
		case seg == "" || seg == ".":
			return "", Error(CodeInvalidInput, "invalid path: "+rel, http.StatusBadRequest, nil)
		case seg == "..":
			return "", Error(CodeInvalidInput, "path traversal is not allowed", http.StatusBadRequest, nil)
		case seg == ".git":
			return "", Error(CodeInvalidInput, "path targets git internals", http.StatusBadRequest, nil)
		}
	}
	return filepath.Join(loc, filepath.FromSlash(rel)), nil
}

// resolveArtifactDir validates that the workspace-relative folder lies within
// the artifact root (e.g. .nomos/services) and returns the absolute creation
// directory. An empty folder means the root itself. This lets predefined
// artifacts be created inside a freely-organized subfolder of their root while
// staying discoverable by the root-scoped scanners.
func resolveArtifactDir(workspace, root, folderRel string) (string, error) {
	rootClean := filepath.Clean(root)
	if strings.TrimSpace(folderRel) == "" {
		return rootClean, nil
	}
	abs, err := safeRepoPath(workspace, folderRel)
	if err != nil {
		return "", err
	}
	if abs != rootClean && !strings.HasPrefix(abs, rootClean+string(filepath.Separator)) {
		return "", Error(CodeInvalidInput, "target folder is outside the artifact root", http.StatusBadRequest, nil)
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		return "", Error(CodeInvalidInput, "target folder does not exist: "+folderRel, http.StatusBadRequest, err)
	}
	return abs, nil
}

// CreateRepoFolder creates a directory at the given repository-relative path,
// mirroring the create-folder gesture in the Explorer tree onto the filesystem.
func CreateRepoFolder(loc, rel string) error {
	abs, err := safeRepoPath(loc, rel)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err == nil {
		return Error(CodeInvalidInput, "a file or folder already exists at "+rel, http.StatusConflict, nil)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return Error(CodeInternalError, "failed to create folder: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

// MoveRepoNode moves the file or folder at fromRel into the folder at toDirRel,
// keeping its base name. It is the filesystem realization of a drag-and-drop in
// the Explorer tree (ADR: folders are a human/git-facing organization layer;
// artifact identity is the stable ID resolved via the index, not the path).
func MoveRepoNode(loc, fromRel, toDirRel string) error {
	fromAbs, err := safeRepoPath(loc, fromRel)
	if err != nil {
		return err
	}
	toDirAbs, err := safeRepoPath(loc, toDirRel)
	if err != nil {
		return err
	}
	fi, err := os.Stat(fromAbs)
	if err != nil {
		return Error(CodeInvalidInput, "source does not exist: "+fromRel, http.StatusNotFound, err)
	}
	if td, err := os.Stat(toDirAbs); err != nil || !td.IsDir() {
		return Error(CodeInvalidInput, "destination is not a folder: "+toDirRel, http.StatusBadRequest, err)
	}
	// Disallow moving a folder into itself or its own subtree.
	if fi.IsDir() {
		rel, err := filepath.Rel(fromAbs, toDirAbs)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return Error(CodeInvalidInput, "cannot move a folder into itself", http.StatusBadRequest, nil)
		}
	}
	dest := filepath.Join(toDirAbs, filepath.Base(fromAbs))
	if dest == fromAbs {
		return nil
	}
	if _, err := os.Stat(dest); err == nil {
		return Error(CodeInvalidInput, "destination already contains "+filepath.Base(fromAbs), http.StatusConflict, nil)
	}
	if err := os.Rename(fromAbs, dest); err != nil {
		return Error(CodeInternalError, "failed to move: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func (m *repoMirror) serviceNode(sn cosmosfs.ServiceNode, rel string) NamespaceTreeNodeDTO {
	s := serviceDTO(sn)
	sd := s
	return NamespaceTreeNodeDTO{
		Label:          s.Name,
		Kind:           "service",
		Canonical:      s.Name,
		Service:        &sd,
		GitPath:        rel,
		TreePath:       rel,
		DisplayPath:    rel,
		Persisted:      true,
		CanOpenDetails: true,
	}
}

func (m *repoMirror) decisionNode(dn cosmosfs.DecisionNode, rel string) NamespaceTreeNodeDTO {
	dec := decisionDTO(dn)
	dd := dec
	return NamespaceTreeNodeDTO{
		Label:          firstNonEmpty(dec.Name, dec.ID),
		Kind:           "decision",
		Canonical:      dec.ID,
		Decision:       &dd,
		GitPath:        rel,
		TreePath:       rel,
		DisplayPath:    rel,
		Persisted:      true,
		CanOpenDetails: true,
	}
}
