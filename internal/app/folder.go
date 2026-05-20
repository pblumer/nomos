package app

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/namespace"
	"github.com/nomos/nomos/internal/storage"
)

// FolderDTO is a namespace folder (ADR-0027): a git-tracked directory whose
// address is derived from its path.
type FolderDTO struct {
	Canonical string `json:"canonical"`
	Label     string `json:"label,omitempty"`
	TreePath  string `json:"treePath"`
	Path      string `json:"path"`
}

var folderSegment = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// folderDir resolves a canonical namespace to its directory and tree path.
// Unlike namespace.Identity it accepts single-label (TLD-level) folders.
func folderDir(path, canonical string) (dir, canon, treePath string, err error) {
	canon = namespace.Canonical(strings.TrimSpace(canonical))
	if canon == "" {
		return "", "", "", Error(CodeInvalidNamespace, "empty namespace", http.StatusBadRequest, nil)
	}
	treePath = namespace.TreePath(canon)
	return domainDirFromTreePath(path, treePath), canon, treePath, nil
}

// CreateFolder creates a namespace folder under an (optional) parent and writes
// folder.yaml so the directory is materialized in git (ADR-0027).
func CreateFolder(path, parentCanonical, label string) (FolderDTO, error) {
	label = strings.TrimSpace(label)
	if !folderSegment.MatchString(label) {
		return FolderDTO{}, Error(CodeInvalidNamespace, "Invalid folder label (use lowercase letters, digits, hyphens): "+label, http.StatusBadRequest, nil)
	}
	parent := namespace.Canonical(strings.TrimSpace(parentCanonical))
	canonical := label
	if parent != "" {
		canonical = label + "." + parent
	}
	dir, canon, tp, err := folderDir(path, canonical)
	if err != nil {
		return FolderDTO{}, err
	}
	if _, err := os.Stat(dir); err == nil {
		return FolderDTO{}, Error(CodeInvalidNamespace, "Folder already exists: "+canon, http.StatusConflict, nil)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return FolderDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	if err := fsx.WriteYAML(storage.FolderFile(dir), model.Folder{Label: label}); err != nil {
		return FolderDTO{}, Error(CodeInternalError, "write folder.yaml: "+err.Error(), http.StatusInternalServerError, err)
	}
	return FolderDTO{Canonical: canon, Label: label, TreePath: tp, Path: dir}, nil
}

// GetFolder resolves a folder by its canonical address.
func GetFolder(path, canonical string) (FolderDTO, error) {
	dir, canon, tp, err := folderDir(path, canonical)
	if err != nil {
		return FolderDTO{}, err
	}
	if _, err := os.Stat(dir); err != nil {
		return FolderDTO{}, Error(CodeDomainNotFound, "Folder not found: "+canon, http.StatusNotFound, nil)
	}
	label := namespace.Label(canon)
	var f model.Folder
	if err := fsx.ReadYAML(storage.FolderFile(dir), &f); err == nil && f.Label != "" {
		label = f.Label
	}
	return FolderDTO{Canonical: canon, Label: label, TreePath: tp, Path: dir}, nil
}

// RenameFolder changes a folder's leaf label, keeping its parent.
func RenameFolder(path, canonical, newLabel string) (FolderDTO, error) {
	parent := namespace.ParentCanonical(namespace.Canonical(canonical))
	return moveFolderTo(path, canonical, parent, newLabel)
}

// MoveFolder reparents a folder (and its whole subtree) under a new parent.
// Because the address is derived from the path, a single directory move
// reparents all descendants — no per-node rewrite (ADR-0027/0028).
func MoveFolder(path, canonical, newParentCanonical string) (FolderDTO, error) {
	return moveFolderTo(path, canonical, namespace.Canonical(newParentCanonical), namespace.Label(namespace.Canonical(canonical)))
}

func moveFolderTo(path, canonical, newParent, newLabel string) (FolderDTO, error) {
	newLabel = strings.TrimSpace(newLabel)
	if !folderSegment.MatchString(newLabel) {
		return FolderDTO{}, Error(CodeInvalidNamespace, "Invalid folder label: "+newLabel, http.StatusBadRequest, nil)
	}
	oldCanon := namespace.Canonical(canonical)
	newCanon := newLabel
	if newParent != "" {
		newCanon = newLabel + "." + newParent
	}
	if newCanon == oldCanon {
		return FolderDTO{}, Error(CodeInvalidInput, "folder already at this location", http.StatusBadRequest, nil)
	}
	if newParent == oldCanon || strings.HasSuffix(newParent, "."+oldCanon) {
		return FolderDTO{}, Error(CodeInvalidInput, "cannot move a folder under its own descendant", http.StatusBadRequest, nil)
	}
	srcDir, _, _, err := folderDir(path, oldCanon)
	if err != nil {
		return FolderDTO{}, err
	}
	if _, err := os.Stat(srcDir); err != nil {
		return FolderDTO{}, Error(CodeDomainNotFound, "Folder not found: "+oldCanon, http.StatusNotFound, nil)
	}
	dstDir, dstCanon, dstTp, err := folderDir(path, newCanon)
	if err != nil {
		return FolderDTO{}, err
	}
	if _, err := os.Stat(dstDir); err == nil {
		return FolderDTO{}, Error(CodeInvalidNamespace, "Target already exists: "+dstCanon, http.StatusConflict, nil)
	}
	if err := os.MkdirAll(filepath.Dir(dstDir), 0o755); err != nil {
		return FolderDTO{}, Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	if err := os.Rename(srcDir, dstDir); err != nil {
		return FolderDTO{}, Error(CodeInternalError, "move failed: "+err.Error(), http.StatusInternalServerError, err)
	}
	// Keep the optional label in sync with the (possibly new) leaf label.
	var f model.Folder
	_ = fsx.ReadYAML(storage.FolderFile(dstDir), &f)
	f.Label = newLabel
	_ = fsx.WriteYAML(storage.FolderFile(dstDir), f)
	return FolderDTO{Canonical: dstCanon, Label: newLabel, TreePath: dstTp, Path: dstDir}, nil
}

// DeleteFolder removes an empty folder (only its folder.yaml may remain).
func DeleteFolder(path, canonical string) error {
	dir, canon, _, err := folderDir(path, canonical)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); err != nil {
		return Error(CodeDomainNotFound, "Folder not found: "+canon, http.StatusNotFound, nil)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	for _, e := range entries {
		if e.Name() != "folder.yaml" {
			return Error(CodeInvalidInput, "Folder is not empty: "+canon, http.StatusConflict, nil)
		}
	}
	if err := os.RemoveAll(dir); err != nil {
		return Error(CodeInternalError, err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}
