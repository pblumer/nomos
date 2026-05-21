package app

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

var typeIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// ListTypeDefs returns the user-defined types under .nomos/types, sorted by ID.
func ListTypeDefs(loc string) ([]model.TypeDef, error) {
	dir := storage.TypesDir(loc)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.TypeDef{}, nil
		}
		return nil, Error(CodeInternalError, "failed to read types directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	out := []model.TypeDef{}
	for _, e := range entries {
		if e.IsDir() || !isYAMLName(e.Name()) {
			continue
		}
		def, ok := readTypeDef(filepath.Join(dir, e.Name()))
		if !ok {
			continue
		}
		if def.ID == "" {
			def.ID = strings.TrimSuffix(strings.TrimSuffix(e.Name(), ".yaml"), ".yml")
		}
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// SaveTypeDef writes (creating .nomos/types if needed) a type definition to
// .nomos/types/<id>.yaml. The id must be a lowercase slug.
func SaveTypeDef(loc string, def model.TypeDef) (model.TypeDef, error) {
	def.ID = strings.TrimSpace(def.ID)
	if !typeIDPattern.MatchString(def.ID) {
		return model.TypeDef{}, Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	dir := storage.TypesDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return model.TypeDef{}, Error(CodeInternalError, "failed to create types directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := fsx.WriteYAML(filepath.Join(dir, def.ID+".yaml"), def); err != nil {
		return model.TypeDef{}, Error(CodeInternalError, "failed to write type definition: "+err.Error(), http.StatusInternalServerError, err)
	}
	return def, nil
}

// DeleteTypeDef removes a type definition file from .nomos/types.
func DeleteTypeDef(loc, id string) error {
	id = strings.TrimSpace(id)
	if !typeIDPattern.MatchString(id) {
		return Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	if err := os.Remove(filepath.Join(storage.TypesDir(loc), id+".yaml")); err != nil {
		if os.IsNotExist(err) {
			return Error(CodeTypeNotFound, "type definition not found: "+id, http.StatusNotFound, err)
		}
		return Error(CodeInternalError, "failed to delete type definition: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

// readTypeDef parses a type definition file; ok is false when the file is
// missing or not valid YAML.
func readTypeDef(path string) (model.TypeDef, bool) {
	var def model.TypeDef
	if err := fsx.ReadYAML(path, &def); err != nil {
		return model.TypeDef{}, false
	}
	return def, true
}

func isYAMLName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
