package server

import (
	"encoding/json"
	"net/http"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

// handleDataObjectRoutes serves repository-scoped CRUD for data objects under
// .../data-objects[/{id}]. seg is the path remainder after data-objects.
func (h *handler) handleDataObjectRoutes(w http.ResponseWriter, r *http.Request, loc string, seg []string) {
	switch len(seg) {
	case 0:
		h.dataObjectCollection(w, r, loc)
	case 1:
		h.dataObjectItem(w, r, loc, seg[0])
	default:
		http.NotFound(w, r)
	}
}

// dataObjectCollection: GET lists data objects, POST creates one.
func (h *handler) dataObjectCollection(w http.ResponseWriter, r *http.Request, loc string) {
	switch r.Method {
	case http.MethodGet:
		objs, err := app.ListDataObjects(loc)
		h.writeOrErr(w, map[string]any{"data_objects": objs}, err)
	case http.MethodPost:
		var obj model.DataObject
		if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		saved, err := app.SaveDataObject(loc, obj)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, saved)
	default:
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// dataObjectItem: GET/PUT/DELETE a single data object.
func (h *handler) dataObjectItem(w http.ResponseWriter, r *http.Request, loc, id string) {
	switch r.Method {
	case http.MethodGet:
		obj, err := app.GetDataObject(loc, id)
		h.writeOrErr(w, obj, err)
	case http.MethodPut:
		var obj model.DataObject
		if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		obj.ID = id
		saved, err := app.SaveDataObject(loc, obj)
		h.writeOrErr(w, saved, err)
	case http.MethodDelete:
		if err := app.DeleteDataObject(loc, id); err != nil {
			h.apiErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
