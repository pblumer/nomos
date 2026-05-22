package server

import (
	"encoding/json"
	"net/http"

	"github.com/nomos/nomos/internal/app"
)

// allowedDMNExt gates the DMN file endpoint to .dmn files.
var allowedDMNExt = map[string]bool{".dmn": true}

// apiDMNFile reads (GET) or writes (PUT) a .dmn file by workspace path, mirroring
// the .frm/.erd file editors. On PUT it normalizes input identifiers and syncs
// the decision I/O contract into a sibling decision.yaml when present.
func (h *handler) apiDMNFile(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("path")
	path, ok := h.resolveWorkspacePath(raw, allowedDMNExt)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or unsupported path"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		content, err := app.ReadDMNFile(path)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"path": raw, "content": content})
	case http.MethodPut:
		var body struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		res, err := app.SaveDMNFile(path, body.Content)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	default:
		w.Header().Set("Allow", "GET, PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
