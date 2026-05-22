package app

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

// TypeInstanceDTO is a single instance of a user-defined type: a directory in
// the working tree that holds the type's detection file (TypeDef.File). Data is
// the parsed YAML content of that file.
type TypeInstanceDTO struct {
	TypeID string         `json:"type_id"`
	Path   string         `json:"path"` // repo-relative directory (slash-separated)
	Name   string         `json:"name"` // base name of the directory
	File   string         `json:"file"` // detection filename inside the directory
	Data   map[string]any `json:"data"`
}

// typeWithDetectionFile loads a type definition and ensures it declares a
// detection file; without one, instances cannot be represented on disk.
func typeWithDetectionFile(loc, typeID string) (model.TypeDef, error) {
	def, err := GetTypeDef(loc, typeID)
	if err != nil {
		return model.TypeDef{}, err
	}
	if strings.TrimSpace(def.File) == "" {
		return model.TypeDef{}, Error(CodeInvalidInput, "type '"+typeID+"' has no detection file; set TypeDef.file to manage instances", http.StatusBadRequest, nil)
	}
	return def, nil
}

// instanceSkipDirs returns the absolute derived directories that never hold
// authored instances (git database and Nomos plumbing).
func instanceSkipDirs(loc string) map[string]bool {
	return map[string]bool{
		filepath.Clean(storage.NomosDir(loc)): true,
		filepath.Clean(storage.ReposDir(loc)): true,
		filepath.Clean(storage.CacheDir(loc)): true,
		filepath.Clean(storage.IndexDir(loc)): true,
	}
}

// ListTypeInstances walks the repository and returns every directory that holds
// the type's detection file, sorted by path.
func ListTypeInstances(loc, typeID string) ([]TypeInstanceDTO, error) {
	def, err := typeWithDetectionFile(loc, typeID)
	if err != nil {
		return nil, err
	}
	out := []TypeInstanceDTO{}
	skip := instanceSkipDirs(loc)
	walkErr := filepath.WalkDir(loc, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.SkipAll
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" || skip[filepath.Clean(p)] {
			return filepath.SkipDir
		}
		if _, statErr := os.Stat(filepath.Join(p, def.File)); statErr != nil {
			return nil
		}
		out = append(out, readInstance(loc, p, def))
		return nil
	})
	if walkErr != nil {
		return nil, Error(CodeInternalError, "failed to scan instances: "+walkErr.Error(), http.StatusInternalServerError, walkErr)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// readInstance builds a DTO for the instance directory at abs.
func readInstance(loc, abs string, def model.TypeDef) TypeInstanceDTO {
	rel, _ := filepath.Rel(loc, abs)
	data := map[string]any{}
	_ = fsx.ReadYAML(filepath.Join(abs, def.File), &data)
	return TypeInstanceDTO{
		TypeID: def.ID,
		Path:   filepath.ToSlash(rel),
		Name:   filepath.Base(abs),
		File:   def.File,
		Data:   data,
	}
}

// GetTypeInstance returns a single instance by its repository-relative directory
// path, or a 404 when the detection file is absent there.
func GetTypeInstance(loc, typeID, relPath string) (TypeInstanceDTO, error) {
	def, err := typeWithDetectionFile(loc, typeID)
	if err != nil {
		return TypeInstanceDTO{}, err
	}
	abs, err := safeRepoPath(loc, relPath)
	if err != nil {
		return TypeInstanceDTO{}, err
	}
	if _, err := os.Stat(filepath.Join(abs, def.File)); err != nil {
		if os.IsNotExist(err) {
			return TypeInstanceDTO{}, Error(CodeTypeInstanceNotFound, "no '"+typeID+"' instance at "+relPath, http.StatusNotFound, err)
		}
		return TypeInstanceDTO{}, Error(CodeInternalError, "failed to read instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return readInstance(loc, abs, def), nil
}

// CreateTypeInstance creates the instance directory (and any parents) at relPath
// and writes the detection file with the given data. It fails if the detection
// file already exists there.
func CreateTypeInstance(loc, typeID, relPath string, data map[string]any) (TypeInstanceDTO, error) {
	def, err := typeWithDetectionFile(loc, typeID)
	if err != nil {
		return TypeInstanceDTO{}, err
	}
	abs, err := safeRepoPath(loc, relPath)
	if err != nil {
		return TypeInstanceDTO{}, err
	}
	file := filepath.Join(abs, def.File)
	if _, err := os.Stat(file); err == nil {
		return TypeInstanceDTO{}, Error(CodeTypeInstanceExists, "a '"+typeID+"' instance already exists at "+relPath, http.StatusConflict, nil)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return TypeInstanceDTO{}, Error(CodeInternalError, "failed to create instance directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	if data == nil {
		data = map[string]any{}
	}
	if err := fsx.WriteYAML(file, data); err != nil {
		return TypeInstanceDTO{}, Error(CodeInternalError, "failed to write instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return readInstance(loc, abs, def), nil
}

// UpdateTypeInstance overwrites an existing instance's detection file with data.
func UpdateTypeInstance(loc, typeID, relPath string, data map[string]any) (TypeInstanceDTO, error) {
	def, err := typeWithDetectionFile(loc, typeID)
	if err != nil {
		return TypeInstanceDTO{}, err
	}
	abs, err := safeRepoPath(loc, relPath)
	if err != nil {
		return TypeInstanceDTO{}, err
	}
	file := filepath.Join(abs, def.File)
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return TypeInstanceDTO{}, Error(CodeTypeInstanceNotFound, "no '"+typeID+"' instance at "+relPath, http.StatusNotFound, err)
		}
		return TypeInstanceDTO{}, Error(CodeInternalError, "failed to read instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	if data == nil {
		data = map[string]any{}
	}
	if err := fsx.WriteYAML(file, data); err != nil {
		return TypeInstanceDTO{}, Error(CodeInternalError, "failed to write instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return readInstance(loc, abs, def), nil
}

// DeleteTypeInstance removes the instance directory at relPath (the directory
// must hold the type's detection file).
func DeleteTypeInstance(loc, typeID, relPath string) error {
	def, err := typeWithDetectionFile(loc, typeID)
	if err != nil {
		return err
	}
	abs, err := safeRepoPath(loc, relPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(abs, def.File)); err != nil {
		if os.IsNotExist(err) {
			return Error(CodeTypeInstanceNotFound, "no '"+typeID+"' instance at "+relPath, http.StatusNotFound, err)
		}
		return Error(CodeInternalError, "failed to read instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := os.RemoveAll(abs); err != nil {
		return Error(CodeInternalError, "failed to delete instance: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}
