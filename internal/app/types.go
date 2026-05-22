package app

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/idgen"
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
	def.IDPrefix = strings.ToUpper(strings.TrimSpace(def.IDPrefix))
	if def.IDPrefix != "" && !idgen.IsValidPrefix(def.IDPrefix) {
		return model.TypeDef{}, Error(CodeInvalidInput, "id_prefix must be exactly three uppercase letters (e.g. RSK)", http.StatusBadRequest, nil)
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

// TypeFormName returns the .frm filename that holds the data-entry form for new
// instances of a type (e.g. "task" -> "task_new.frm").
func TypeFormName(typeID string) string { return typeID + "_new.frm" }

// LoadTypeForm reads the data-entry form (.frm) for a type from .nomos/views.
// ok is false when no form has been authored yet.
func LoadTypeForm(loc, typeID string) (model.View, bool, error) {
	typeID = strings.TrimSpace(typeID)
	if !typeIDPattern.MatchString(typeID) {
		return model.View{}, false, Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	path := filepath.Join(storage.ViewsDir(loc), TypeFormName(typeID))
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return model.View{}, false, nil
		}
		return model.View{}, false, Error(CodeInternalError, "failed to read form: "+err.Error(), http.StatusInternalServerError, err)
	}
	var v model.View
	if err := fsx.ReadYAML(path, &v); err != nil {
		return model.View{}, false, Error(CodeInternalError, "failed to parse form: "+err.Error(), http.StatusInternalServerError, err)
	}
	return v, true, nil
}

// SaveTypeForm writes (creating .nomos/views if needed) the data-entry form for
// a type to .nomos/views/<id>_new.frm. Engine defaults to "form-js".
func SaveTypeForm(loc, typeID string, v model.View) (model.View, error) {
	typeID = strings.TrimSpace(typeID)
	if !typeIDPattern.MatchString(typeID) {
		return model.View{}, Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	if v.Engine == "" {
		v.Engine = "form-js"
	}
	dir := storage.ViewsDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return model.View{}, Error(CodeInternalError, "failed to create views directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := fsx.WriteYAML(filepath.Join(dir, TypeFormName(typeID)), v); err != nil {
		return model.View{}, Error(CodeInternalError, "failed to write form: "+err.Error(), http.StatusInternalServerError, err)
	}
	return v, nil
}

// GetTypeDef returns a single type definition by id, or a 404 error when no
// .nomos/types/<id>.yaml exists.
func GetTypeDef(loc, id string) (model.TypeDef, error) {
	id = strings.TrimSpace(id)
	if !typeIDPattern.MatchString(id) {
		return model.TypeDef{}, Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	path := filepath.Join(storage.TypesDir(loc), id+".yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return model.TypeDef{}, Error(CodeTypeNotFound, "type not found: "+id, http.StatusNotFound, err)
		}
		return model.TypeDef{}, Error(CodeInternalError, "failed to read type definition: "+err.Error(), http.StatusInternalServerError, err)
	}
	def, ok := readTypeDef(path)
	if !ok {
		return model.TypeDef{}, Error(CodeInternalError, "type definition is not valid YAML: "+id, http.StatusInternalServerError, nil)
	}
	if def.ID == "" {
		def.ID = id
	}
	return def, nil
}

// DeleteTypeDef removes a type definition (.nomos/types/<id>.yaml) and its
// data-entry form (.nomos/views/<id>_new.frm) when present.
func DeleteTypeDef(loc, id string) error {
	id = strings.TrimSpace(id)
	if !typeIDPattern.MatchString(id) {
		return Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	path := filepath.Join(storage.TypesDir(loc), id+".yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Error(CodeTypeNotFound, "type not found: "+id, http.StatusNotFound, err)
		}
		return Error(CodeInternalError, "failed to read type definition: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := os.Remove(path); err != nil {
		return Error(CodeInternalError, "failed to delete type definition: "+err.Error(), http.StatusInternalServerError, err)
	}
	if formPath := filepath.Join(storage.ViewsDir(loc), TypeFormName(id)); formPath != "" {
		if err := os.Remove(formPath); err != nil && !os.IsNotExist(err) {
			return Error(CodeInternalError, "failed to delete type form: "+err.Error(), http.StatusInternalServerError, err)
		}
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
