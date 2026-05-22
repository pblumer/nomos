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
	"github.com/nomos/nomos/internal/viewgen"
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
	// Ensure the type ships with its four standard view forms (ADR-0024). Only
	// variants that do not yet exist are written, so editing a type never
	// clobbers forms the user has already authored.
	if err := SeedTypeForms(loc, def); err != nil {
		return model.TypeDef{}, err
	}
	return def, nil
}

// SeedTypeForms writes the four standard view forms (new/edit/list/short) for a
// type under .nomos/views, generating each from the type's properties. Existing
// form files are left untouched, so it is safe to call on every save.
func SeedTypeForms(loc string, def model.TypeDef) error {
	dir := storage.ViewsDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Error(CodeInternalError, "failed to create views directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	props := typePropsToViewgen(def.Properties)
	for _, v := range viewgen.Variants {
		path := filepath.Join(dir, viewgen.FormName(def.ID, v))
		if _, err := os.Stat(path); err == nil {
			continue // keep an authored form
		} else if !os.IsNotExist(err) {
			return Error(CodeInternalError, "failed to inspect form: "+err.Error(), http.StatusInternalServerError, err)
		}
		view := model.View{Engine: "form-js", EngineVersion: "1", Schema: viewgen.Schema(def.ID, def.Label, props, v)}
		if err := fsx.WriteYAML(path, view); err != nil {
			return Error(CodeInternalError, "failed to write form: "+err.Error(), http.StatusInternalServerError, err)
		}
	}
	return nil
}

func typePropsToViewgen(props []model.TypeProperty) []viewgen.Prop {
	out := make([]viewgen.Prop, 0, len(props))
	for _, p := range props {
		out = append(out, viewgen.Prop{Name: p.Name, Type: p.Type, Label: p.Label, Required: p.Required, Description: p.Description})
	}
	return out
}

// normalizeFormVariant maps a request-supplied variant to a known one,
// defaulting to "new" when empty.
func normalizeFormVariant(variant string) (viewgen.Variant, error) {
	variant = strings.TrimSpace(variant)
	if variant == "" {
		return viewgen.VariantNew, nil
	}
	if !viewgen.IsVariant(variant) {
		return "", Error(CodeInvalidInput, "unknown form variant: "+variant, http.StatusBadRequest, nil)
	}
	return viewgen.Variant(variant), nil
}

// TypeFormName returns the .frm filename that holds the data-entry form for new
// instances of a type (e.g. "task" -> "task_new.frm").
func TypeFormName(typeID string) string { return viewgen.FormName(typeID, viewgen.VariantNew) }

// LoadTypeForm reads the "new" data-entry form (.frm) for a type. ok is false
// when no form has been authored yet.
func LoadTypeForm(loc, typeID string) (model.View, bool, error) {
	return LoadTypeFormVariant(loc, typeID, "")
}

// LoadTypeFormVariant reads one of a type's standard view forms (new/edit/list/
// short) from .nomos/views. An empty variant defaults to "new". ok is false
// when that form has not been authored yet.
func LoadTypeFormVariant(loc, typeID, variant string) (model.View, bool, error) {
	typeID = strings.TrimSpace(typeID)
	if !typeIDPattern.MatchString(typeID) {
		return model.View{}, false, Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	v, err := normalizeFormVariant(variant)
	if err != nil {
		return model.View{}, false, err
	}
	path := filepath.Join(storage.ViewsDir(loc), viewgen.FormName(typeID, v))
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return model.View{}, false, nil
		}
		return model.View{}, false, Error(CodeInternalError, "failed to read form: "+err.Error(), http.StatusInternalServerError, err)
	}
	var view model.View
	if err := fsx.ReadYAML(path, &view); err != nil {
		return model.View{}, false, Error(CodeInternalError, "failed to parse form: "+err.Error(), http.StatusInternalServerError, err)
	}
	return view, true, nil
}

// SaveTypeForm writes the "new" data-entry form for a type.
func SaveTypeForm(loc, typeID string, v model.View) (model.View, error) {
	return SaveTypeFormVariant(loc, typeID, "", v)
}

// SaveTypeFormVariant writes (creating .nomos/views if needed) one of a type's
// standard view forms to .nomos/views/<id>_<variant>.frm. An empty variant
// defaults to "new". Engine defaults to "form-js".
func SaveTypeFormVariant(loc, typeID, variant string, v model.View) (model.View, error) {
	typeID = strings.TrimSpace(typeID)
	if !typeIDPattern.MatchString(typeID) {
		return model.View{}, Error(CodeInvalidInput, "type id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	variantKind, err := normalizeFormVariant(variant)
	if err != nil {
		return model.View{}, err
	}
	if v.Engine == "" {
		v.Engine = "form-js"
	}
	dir := storage.ViewsDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return model.View{}, Error(CodeInternalError, "failed to create views directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := fsx.WriteYAML(filepath.Join(dir, viewgen.FormName(typeID, variantKind)), v); err != nil {
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

// DeleteTypeDef removes a type definition (.nomos/types/<id>.yaml) and its four
// standard view forms (.nomos/views/<id>_{new,edit,list,short}.frm) when present.
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
	for _, v := range viewgen.Variants {
		formPath := filepath.Join(storage.ViewsDir(loc), viewgen.FormName(id, v))
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
