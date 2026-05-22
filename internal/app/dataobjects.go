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

// ListDataObjects returns the data objects under .nomos/data, sorted by ID.
func ListDataObjects(loc string) ([]model.DataObject, error) {
	dir := storage.DataDir(loc)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.DataObject{}, nil
		}
		return nil, Error(CodeInternalError, "failed to read data directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	out := []model.DataObject{}
	for _, e := range entries {
		if e.IsDir() || !isYAMLName(e.Name()) {
			continue
		}
		obj, ok := readDataObject(filepath.Join(dir, e.Name()))
		if !ok {
			continue
		}
		if obj.ID == "" {
			obj.ID = strings.TrimSuffix(strings.TrimSuffix(e.Name(), ".yaml"), ".yml")
		}
		out = append(out, obj)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// GetDataObject returns a single data object by id, or a 404 error when no
// .nomos/data/<id>.yaml exists.
func GetDataObject(loc, id string) (model.DataObject, error) {
	id = strings.TrimSpace(id)
	if !typeIDPattern.MatchString(id) {
		return model.DataObject{}, Error(CodeInvalidInput, "data object id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	path := filepath.Join(storage.DataDir(loc), id+".yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return model.DataObject{}, Error(CodeDataObjectNotFound, "data object not found: "+id, http.StatusNotFound, err)
		}
		return model.DataObject{}, Error(CodeInternalError, "failed to read data object: "+err.Error(), http.StatusInternalServerError, err)
	}
	obj, ok := readDataObject(path)
	if !ok {
		return model.DataObject{}, Error(CodeInternalError, "data object is not valid YAML: "+id, http.StatusInternalServerError, nil)
	}
	if obj.ID == "" {
		obj.ID = id
	}
	return obj, nil
}

// SaveDataObject writes (creating .nomos/data if needed) a data object to
// .nomos/data/<id>.yaml. The id must be a lowercase slug.
func SaveDataObject(loc string, obj model.DataObject) (model.DataObject, error) {
	obj.ID = strings.TrimSpace(obj.ID)
	if !typeIDPattern.MatchString(obj.ID) {
		return model.DataObject{}, Error(CodeInvalidInput, "data object id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	dir := storage.DataDir(loc)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return model.DataObject{}, Error(CodeInternalError, "failed to create data directory: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := fsx.WriteYAML(filepath.Join(dir, obj.ID+".yaml"), obj); err != nil {
		return model.DataObject{}, Error(CodeInternalError, "failed to write data object: "+err.Error(), http.StatusInternalServerError, err)
	}
	return obj, nil
}

// DeleteDataObject removes a data object (.nomos/data/<id>.yaml).
func DeleteDataObject(loc, id string) error {
	id = strings.TrimSpace(id)
	if !typeIDPattern.MatchString(id) {
		return Error(CodeInvalidInput, "data object id must be a lowercase slug (a-z, 0-9, -, _)", http.StatusBadRequest, nil)
	}
	path := filepath.Join(storage.DataDir(loc), id+".yaml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Error(CodeDataObjectNotFound, "data object not found: "+id, http.StatusNotFound, err)
		}
		return Error(CodeInternalError, "failed to read data object: "+err.Error(), http.StatusInternalServerError, err)
	}
	if err := os.Remove(path); err != nil {
		return Error(CodeInternalError, "failed to delete data object: "+err.Error(), http.StatusInternalServerError, err)
	}
	return nil
}

func readDataObject(path string) (model.DataObject, bool) {
	var obj model.DataObject
	if err := fsx.ReadYAML(path, &obj); err != nil {
		return model.DataObject{}, false
	}
	return obj, true
}
