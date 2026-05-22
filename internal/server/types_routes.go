package server

import (
	"encoding/json"
	"net/http"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

// handleTypeRoutes serves the repository-scoped, RESTful CRUD for user-defined
// types and their instances. seg is the path remainder after
// /api/v1/repositories/{repo}/types, i.e.:
//
//	[]                       -> collection of type definitions
//	[id]                     -> a single type definition
//	[id, "instances"]        -> instances of a type (list/create; item via ?path=)
func (h *handler) handleTypeRoutes(w http.ResponseWriter, r *http.Request, loc string, seg []string) {
	switch len(seg) {
	case 0:
		h.typeDefCollection(w, r, loc)
	case 1:
		h.typeDefItem(w, r, loc, seg[0])
	case 2:
		switch seg[1] {
		case "instances":
			h.typeInstances(w, r, loc, seg[0])
		case "new-id":
			h.typeNewID(w, r, loc, seg[0])
		default:
			http.NotFound(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

// typeDefCollection: GET lists type definitions, POST creates one.
func (h *handler) typeDefCollection(w http.ResponseWriter, r *http.Request, loc string) {
	switch r.Method {
	case http.MethodGet:
		defs, err := app.ListTypeDefs(loc)
		h.writeOrErr(w, map[string]any{"types": defs}, err)
	case http.MethodPost:
		var def model.TypeDef
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		saved, err := app.SaveTypeDef(loc, def)
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

// typeDefItem: GET/PUT/DELETE a single type definition.
func (h *handler) typeDefItem(w http.ResponseWriter, r *http.Request, loc, id string) {
	switch r.Method {
	case http.MethodGet:
		def, err := app.GetTypeDef(loc, id)
		h.writeOrErr(w, def, err)
	case http.MethodPut:
		var def model.TypeDef
		if err := json.NewDecoder(r.Body).Decode(&def); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		def.ID = id
		saved, err := app.SaveTypeDef(loc, def)
		h.writeOrErr(w, saved, err)
	case http.MethodDelete:
		if err := app.DeleteTypeDef(loc, id); err != nil {
			h.apiErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// typeNewID mints a fresh instance id from the type's IDPrefix.
func (h *handler) typeNewID(w http.ResponseWriter, r *http.Request, loc, typeID string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, err := app.NewTypeInstanceID(loc, typeID)
	h.writeOrErr(w, map[string]string{"id": id}, err)
}

// typeInstances serves instances of a type. The collection lives at
// .../types/{id}/instances; a single instance is addressed with ?path=<dir>.
//
//	GET    (no ?path)  -> list instances
//	GET    ?path=X     -> read one instance
//	POST   {path,data} -> create an instance
//	PUT    ?path=X      -> overwrite an instance's data
//	DELETE ?path=X      -> remove an instance
func (h *handler) typeInstances(w http.ResponseWriter, r *http.Request, loc, typeID string) {
	path := r.URL.Query().Get("path")
	switch r.Method {
	case http.MethodGet:
		if path == "" {
			items, err := app.ListTypeInstances(loc, typeID)
			h.writeOrErr(w, map[string]any{"type_id": typeID, "instances": items}, err)
			return
		}
		inst, err := app.GetTypeInstance(loc, typeID, path)
		h.writeOrErr(w, inst, err)
	case http.MethodPost:
		var body struct {
			Path string         `json:"path"`
			Data map[string]any `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		inst, err := app.CreateTypeInstance(loc, typeID, body.Path, body.Data)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, inst)
	case http.MethodPut:
		var body struct {
			Data map[string]any `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		inst, err := app.UpdateTypeInstance(loc, typeID, path, body.Data)
		h.writeOrErr(w, inst, err)
	case http.MethodDelete:
		if err := app.DeleteTypeInstance(loc, typeID, path); err != nil {
			h.apiErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
