package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

// sourceLanguage maps a file extension to the editor language the frontend uses.
func sourceLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	default:
		return "text"
	}
}

var allowedSourceExt = map[string]bool{
	".yaml": true, ".yml": true, ".json": true, ".md": true, ".markdown": true,
}

// allowedViewExt gates the view endpoint to .frm form files (ADR-0024).
var allowedViewExt = map[string]bool{".frm": true}

// allowedErdExt gates the ERD endpoint to .erd relationship-diagram files.
var allowedErdExt = map[string]bool{".erd": true}

// resolveSourcePath maps a client-supplied path to an absolute source file
// inside the named repository (empty/"default" = workspace), rejecting
// traversal outside the root and disallowed types.
func (h *handler) resolveSourcePath(repoID, raw string) (string, bool) {
	return h.resolveRepoPath(repoID, raw, allowedSourceExt)
}

// repoRoot returns the working-directory root for a repository id. An empty id
// (or the default repository) resolves to the local workspace; any other id is
// looked up in the registry so file editors can reach files inside attached
// repositories, whose paths are repo-relative (ADR-0022).
func (h *handler) repoRoot(repoID string) (string, bool) {
	if repoID == "" || repoID == app.LocalDefaultRepoID {
		return h.cosmosPath, true
	}
	dto, err := app.GetRepository(h.cosmosPath, repoID)
	if err != nil || dto.Location == "" {
		return "", false
	}
	return dto.Location, true
}

// resolveRepoPath maps a client-supplied, repository-relative path to an
// absolute file inside the named repository's working directory, applying the
// same traversal/extension guards as resolveWorkspacePath.
func (h *handler) resolveRepoPath(repoID, raw string, allowed map[string]bool) (string, bool) {
	root, ok := h.repoRoot(repoID)
	if !ok {
		return "", false
	}
	return resolvePathUnder(root, raw, allowed)
}

// resolveWorkspacePath maps a client-supplied path to an absolute file inside
// the cosmos workspace given an extension allow-set, rejecting traversal
// outside the root and disallowed types.
func (h *handler) resolveWorkspacePath(raw string, allowed map[string]bool) (string, bool) {
	return resolvePathUnder(h.cosmosPath, raw, allowed)
}

// resolvePathUnder resolves raw against root, rejecting disallowed extensions
// and any target that escapes root. Two bases are tried: the working directory
// (artifact DTO paths carry the same base as root) and root itself (Explorer
// file nodes carry repo-relative paths). The first candidate that stays inside
// root wins.
func resolvePathUnder(rootPath, raw string, allowed map[string]bool) (string, bool) {
	if raw == "" {
		return "", false
	}
	if !allowed[strings.ToLower(filepath.Ext(raw))] {
		return "", false
	}
	root, err := filepath.Abs(rootPath)
	if err != nil {
		return "", false
	}
	var candidates []string
	if abs, err := filepath.Abs(raw); err == nil {
		candidates = append(candidates, abs)
	}
	if !filepath.IsAbs(raw) {
		candidates = append(candidates, filepath.Join(root, raw))
	}
	for _, target := range candidates {
		rel, err := filepath.Rel(root, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		return target, true
	}
	return "", false
}

// apiSource reads (GET) or writes (PUT) a single artifact source file. GET with
// ?render=1 returns sanitized HTML for Markdown sources. PUT validates syntax,
// writes the file, then runs cosmos validation so the backend stays the
// authority on artifact correctness.
func (h *handler) apiSource(w http.ResponseWriter, r *http.Request) {
	path, ok := h.resolveSourcePath(r.URL.Query().Get("repo"), r.URL.Query().Get("path"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or unsupported path"})
		return
	}
	lang := sourceLanguage(path)
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if r.URL.Query().Get("render") == "1" && lang == "markdown" {
			writeJSON(w, http.StatusOK, map[string]any{
				"path": r.URL.Query().Get("path"), "language": lang, "html": renderMarkdown(data),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"path": r.URL.Query().Get("path"), "language": lang, "content": string(data),
		})
	case http.MethodPut:
		var body struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if msg := checkSourceSyntax(lang, body.Content); msg != "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "syntaxError": msg})
			return
		}
		info, statErr := os.Stat(path)
		mode := os.FileMode(0o644)
		if statErr == nil {
			mode = info.Mode().Perm()
		}
		if err := os.WriteFile(path, []byte(body.Content), mode); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		resp := map[string]any{"ok": true, "language": lang}
		if val, err := app.ValidateCosmos(h.cosmosPath); err == nil {
			resp["validation"] = val
		}
		writeJSON(w, http.StatusOK, resp)
	default:
		w.Header().Set("Allow", "GET, PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiSourceScaffold returns a snippet of the required fields the posted YAML
// content is still missing, for the file at ?path. It never writes; the client
// appends the snippet to the editor for the user to fill in and review.
func (h *handler) apiSourceScaffold(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.resolveSourcePath(r.URL.Query().Get("repo"), r.URL.Query().Get("path")); !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or unsupported path"})
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	snippet, missing, err := app.RequiredFieldScaffold(body.Content)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	if missing == nil {
		missing = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"snippet": snippet, "missing": missing})
}

// apiView reads (GET) or writes (PUT) a .frm form file by workspace path,
// translating between its YAML on disk and the form-js schema the editor wants.
// It complements the type-keyed type-form endpoint by letting the Explorer open
// any .frm directly in the form-js editor (ADR-0024).
func (h *handler) apiView(w http.ResponseWriter, r *http.Request) {
	path, ok := h.resolveRepoPath(r.URL.Query().Get("repo"), r.URL.Query().Get("path"), allowedViewExt)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or unsupported path"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		var v model.View
		if err := yaml.Unmarshal(data, &v); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "form is not valid YAML: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"path": r.URL.Query().Get("path"), "view": v})
	case http.MethodPut:
		var v model.View
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if v.Engine == "" {
			v.Engine = "form-js"
		}
		out, err := yaml.Marshal(v)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		mode := os.FileMode(0o644)
		if info, statErr := os.Stat(path); statErr == nil {
			mode = info.Mode().Perm()
		}
		if err := os.WriteFile(path, out, mode); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		w.Header().Set("Allow", "GET, PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiERD reads (GET) or writes (PUT) a .erd relationship-diagram file by
// workspace path, translating between its YAML on disk and JSON for the
// lightweight SVG editor. Only the layout is stored here; relationships live in
// the type definitions' dependencies.
func (h *handler) apiERD(w http.ResponseWriter, r *http.Request) {
	path, ok := h.resolveRepoPath(r.URL.Query().Get("repo"), r.URL.Query().Get("path"), allowedErdExt)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or unsupported path"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		var e model.ERD
		if err := yaml.Unmarshal(data, &e); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "ERD is not valid YAML: " + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"path": r.URL.Query().Get("path"), "erd": e})
	case http.MethodPut:
		var e model.ERD
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if e.Engine == "" {
			e.Engine = "nomos-erd"
		}
		out, err := yaml.Marshal(e)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		mode := os.FileMode(0o644)
		if info, statErr := os.Stat(path); statErr == nil {
			mode = info.Mode().Perm()
		}
		if err := os.WriteFile(path, out, mode); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		w.Header().Set("Allow", "GET, PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiRenderMarkdown renders an arbitrary Markdown body to sanitized HTML, used
// by the editor preview without persisting anything.
func (h *handler) apiRenderMarkdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"html": renderMarkdown([]byte(body.Content))})
}

// checkSourceSyntax returns a human-readable error if the content is not
// well-formed for its language, or "" when it parses.
func checkSourceSyntax(lang, content string) string {
	switch lang {
	case "yaml":
		var v any
		if err := yaml.Unmarshal([]byte(content), &v); err != nil {
			return err.Error()
		}
	case "json":
		var v any
		if err := json.Unmarshal([]byte(content), &v); err != nil {
			return err.Error()
		}
	}
	return ""
}

var (
	mdRenderer = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
	)
	mdSanitizer = bluemonday.UGCPolicy()
)

// renderMarkdown converts Markdown to HTML and sanitizes it against XSS.
func renderMarkdown(src []byte) string {
	var buf strings.Builder
	if err := mdRenderer.Convert(src, &buf); err != nil {
		return ""
	}
	return mdSanitizer.Sanitize(buf.String())
}
