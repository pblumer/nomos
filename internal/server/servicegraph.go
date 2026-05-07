package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

func (h *handler) apiServicegraphs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/servicegraphs" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.ListServicegraphs(h.cosmosPath)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
	case http.MethodPost:
		var sg model.Servicegraph
		if err := json.NewDecoder(r.Body).Decode(&sg); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}
		if err := app.CreateServicegraph(h.cosmosPath, sg); err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"created": sg.ID})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *handler) apiServicegraphRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/servicegraphs/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			dto, err := app.GetServicegraph(h.cosmosPath, id)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case http.MethodDelete:
			if err := app.DeleteServicegraph(h.cosmosPath, id); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, map[string]string{"deleted": id})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "mermaid":
			dto, err := app.GetServicegraphMermaid(h.cosmosPath, id)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case "execution":
			dto, err := app.GetServicegraphExecutionOrder(h.cosmosPath, id)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		default:
			http.NotFound(w, r)
		}
		return
	}
	http.NotFound(w, r)
}
