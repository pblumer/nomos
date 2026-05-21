package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mcphttp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/mcpserver"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/update"
	versionpkg "github.com/nomos/nomos/internal/version"
)

//go:embed web/templates/* web/static/*
var webFS embed.FS

type handler struct {
	cosmosPath string
	tmpl       *template.Template
	updater    *update.Checker
}

func NewHandler(cosmosPath string) http.Handler {
	t := template.Must(template.New("web").ParseFS(webFS, "web/templates/*.html"))
	staticFS := must(fs.Sub(webFS, "web/static"))
	h := &handler{cosmosPath: cosmosPath, tmpl: t, updater: update.NewChecker()}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/v1/version", h.apiVersion)
	mux.HandleFunc("/openapi.json", h.openAPIJSON)
	mux.HandleFunc("/swagger", h.swaggerUI)
	mux.HandleFunc("/swagger/", h.swaggerUI)
	mux.HandleFunc("/api/docs", h.swaggerUI)
	mux.HandleFunc("/api/v1/cosmos", h.apiCosmos)
	mux.HandleFunc("/api/v1/repositories", h.apiRepositories)
	mux.HandleFunc("/api/v1/repositories/", h.apiRepositoryRoutes)
	mux.HandleFunc("/api/v1/mounts", h.apiMounts)
	mux.HandleFunc("/api/v1/mounts/", h.apiMountRoutes)
	mux.HandleFunc("/api/v1/ping", h.apiPing)
	mux.HandleFunc("/api/v1/discover", h.apiDiscover)
	mux.HandleFunc("/api/v1/index", h.apiIndex)
	mux.HandleFunc("/api/v1/index/", h.apiIndexResolve)
	mux.HandleFunc("/api/v1/services", h.apiServices)
	mux.HandleFunc("/api/v1/services/refs", h.apiServiceRefs)
	mux.HandleFunc("/api/v1/services/", h.apiServiceRoutes)
	mux.HandleFunc("/api/v1/decisions", h.apiDecisions)
	mux.HandleFunc("/api/v1/decisions/", h.apiDecisionRoutes)
	mux.HandleFunc("/api/v1/namespaces", h.apiNamespaces)
	mux.HandleFunc("/api/v1/graph", h.apiGraph)
	mux.HandleFunc("/api/v1/validate", h.apiValidate)
	mux.HandleFunc("/api/v1/source", h.apiSource)
	mux.HandleFunc("/api/v1/render/markdown", h.apiRenderMarkdown)
	mux.HandleFunc("/api/v1/blueprints", h.apiBlueprints)
	mux.HandleFunc("/api/v1/blueprints/", h.apiBlueprintRoutes)
	mux.HandleFunc("/api/v1/products/", h.apiProductRoutes)
	mux.HandleFunc("/api/v1/processes/", h.apiProcessRoutes)
	mux.HandleFunc("/api/v1/instances", h.apiInstances)
	mux.HandleFunc("/api/v1/instances/", h.apiInstanceRoutes)
	mux.HandleFunc("/api/v1/product-instances/", h.apiProvisionServiceInstance)
	mux.HandleFunc("/api/v1/servicegraphs", h.apiServicegraphs)
	mux.HandleFunc("/api/v1/servicegraphs/", h.apiServicegraphRoutes)
	mux.HandleFunc("/api/blueprints", h.apiBlueprints)
	mux.HandleFunc("/api/blueprints/", h.apiBlueprintRoutes)
	mux.HandleFunc("/api/instances", h.apiInstances)
	mux.HandleFunc("/api/instances/", h.apiInstanceRoutes)
	mux.Handle("/mcp", mcphttp.NewStreamableHTTPHandler(func(_ *http.Request) *mcphttp.Server {
		return mcpserver.New(cosmosPath)
	}, nil))
	mux.HandleFunc("/", h.routes)
	return apiKeyAuth(cosmosPath, mux)
}
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"service": "nomos", "status": "ok", "version": versionpkg.Get().Version})
}

// apiVersion returns the running server's full build version and, when the
// opt-in update check is enabled, whether a newer release is available.
func (h *handler) apiVersion(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/version" {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, 200, map[string]any{
		"version": versionpkg.Get(),
		"update":  h.updater.Status(),
	})
}
func (h *handler) apiCosmos(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/cosmos" {
		http.NotFound(w, r)
		return
	}
	dto, err := app.GetCosmos(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}
func (h *handler) apiRepositories(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/repositories" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.ListRepositories(h.cosmosPath)
		h.writeOrErr(w, dto, err)
	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		dto, err := app.CreateRepository(h.cosmosPath, body.Name)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, dto)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiRepositoryRoutes serves repository-scoped reads (ADR-0022 §2). It resolves
// {repo} to its workspace and delegates to the same app read functions the
// non-scoped aliases use, so /api/v1/repositories/default/cosmos and
// /api/v1/cosmos return identical payloads.
func (h *handler) apiRepositoryRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/repositories/")
	parts := strings.Split(rest, "/")
	repoID := parts[0]
	if repoID == "" {
		http.NotFound(w, r)
		return
	}
	repoDTO, err := app.GetRepository(h.cosmosPath, repoID)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	loc := repoDTO.Location
	resource := strings.Join(parts[1:], "/")

	// RESTful CRUD for user-defined types and their instances lives under
	// .../types[/{id}[/instances]] and supports GET/POST/PUT/DELETE. The legacy
	// POST .../fs/type and .../fs/type-form gestures stay handled below.
	if len(parts) > 1 && parts[1] == "types" {
		h.handleTypeRoutes(w, r, loc, parts[2:])
		return
	}

	if resource == "" && r.Method == http.MethodDelete {
		if err := app.DeleteRepository(h.cosmosPath, repoID); err != nil {
			h.apiErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if resource == "" && (r.Method == http.MethodPatch || r.Method == http.MethodPut) {
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		dto, err := app.RenameRepository(h.cosmosPath, repoID, body.Name)
		h.writeOrErr(w, dto, err)
		return
	}

	if r.Method == http.MethodPost {
		switch resource {
		case "fs/delete":
			var body struct {
				Path string `json:"path"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			if err := app.DeleteRepoNode(loc, body.Path); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"path": body.Path})
		case "fs/rename":
			var body struct {
				Path string `json:"path"`
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			if err := app.RenameRepoNode(loc, body.Path, body.Name); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"path": body.Path, "name": body.Name})
		case "git/commit":
			var body struct {
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			dto, err := app.CommitRepo(h.cosmosPath, repoID, body.Message)
			h.writeOrErr(w, dto, err)
		case "git/branches":
			var body struct {
				Name     string `json:"name"`
				Checkout bool   `json:"checkout"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			dto, err := app.CreateBranch(h.cosmosPath, repoID, body.Name, body.Checkout)
			h.writeOrErr(w, dto, err)
		case "git/checkout":
			var body struct {
				Branch string `json:"branch"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			dto, err := app.CheckoutBranch(h.cosmosPath, repoID, body.Branch)
			h.writeOrErr(w, dto, err)
		case "git/tags":
			var body struct {
				Name    string `json:"name"`
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			dto, err := app.CreateTag(h.cosmosPath, repoID, body.Name, body.Message)
			h.writeOrErr(w, dto, err)
		case "fs/folder":
			var body struct {
				Path         string   `json:"path"`
				Label        string   `json:"label"`
				Description  string   `json:"description"`
				AllowedTypes []string `json:"allowed_types"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			meta := &model.FolderMeta{Label: body.Label, Description: body.Description, AllowedTypes: body.AllowedTypes}
			if err := app.CreateRepoFolder(loc, body.Path, meta); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]string{"path": body.Path})
		case "fs/folder-meta":
			var body struct {
				Path         string   `json:"path"`
				Label        string   `json:"label"`
				Description  string   `json:"description"`
				AllowedTypes []string `json:"allowed_types"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			if err := app.SetFolderMeta(loc, body.Path, model.FolderMeta{Label: body.Label, Description: body.Description, AllowedTypes: body.AllowedTypes}); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"path": body.Path})
		case "fs/type":
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
		case "fs/type-form":
			var body struct {
				TypeID        string             `json:"type_id"`
				Engine        string             `json:"engine"`
				EngineVersion string             `json:"engine_version"`
				Schema        map[string]any     `json:"schema"`
				Binding       *model.ViewBinding `json:"binding"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			saved, err := app.SaveTypeForm(loc, body.TypeID, model.View{Engine: body.Engine, EngineVersion: body.EngineVersion, Schema: body.Schema, Binding: body.Binding})
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, saved)
		case "fs/move":
			var body struct {
				From  string `json:"from"`
				ToDir string `json:"to_dir"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
				return
			}
			if err := app.MoveRepoNode(loc, body.From, body.ToDir); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"from": body.From, "to_dir": body.ToDir})
		default:
			http.NotFound(w, r)
		}
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	switch {
	case resource == "":
		writeJSON(w, 200, repoDTO)
	case resource == "cosmos":
		dto, err := app.GetCosmos(loc)
		h.writeOrErr(w, dto, err)
	case resource == "namespaces":
		dto, err := app.BuildNamespaceTree(loc)
		h.writeOrErr(w, dto, err)
	case resource == "git/status":
		dto, err := app.GitStatus(h.cosmosPath, repoID)
		h.writeOrErr(w, dto, err)
	case resource == "git/branches":
		dto, err := app.GitBranches(h.cosmosPath, repoID)
		h.writeOrErr(w, dto, err)
	case resource == "git/tags":
		dto, err := app.GitTags(h.cosmosPath, repoID)
		h.writeOrErr(w, dto, err)
	case resource == "type-form":
		v, ok, err := app.LoadTypeForm(loc, r.URL.Query().Get("type_id"))
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"exists": ok, "view": v})
	default:
		http.NotFound(w, r)
	}
}

// apiMounts lists or adds server mounts (ADR-0022 §5).
func (h *handler) apiMounts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/mounts" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.ListMounts(h.cosmosPath)
		h.writeOrErr(w, dto, err)
	case http.MethodPost:
		var body struct {
			Endpoint string `json:"endpoint"`
			Label    string `json:"label"`
			Token    string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		dto, err := app.AddMount(h.cosmosPath, body.Endpoint, body.Label, body.Token)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, dto)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiMountRoutes removes a server mount by id (ADR-0022 §5) or proxies a request
// to a remote mounted server (ADR-0023). Proxy paths are
// /api/v1/mounts/{id}/r/<remote-path>; the local server forwards the request to
// http://{endpoint}/<remote-path> with the mount token as X-API-Key.
func (h *handler) apiMountRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/mounts/")
	if i := strings.Index(rest, "/r/"); i >= 0 {
		h.proxyMount(w, r, rest[:i], rest[i+len("/r/"):])
		return
	}
	if id, ok := strings.CutSuffix(rest, "/r"); ok {
		h.proxyMount(w, r, id, "")
		return
	}
	id := rest
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := app.RemoveMount(h.cosmosPath, id); err != nil {
		h.apiErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var mountProxyClient = &http.Client{Timeout: 10 * time.Second}

func isMutating(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

func (h *handler) proxyMount(w http.ResponseWriter, r *http.Request, mountID, remotePath string) {
	endpoint, token, err := app.MountTarget(h.cosmosPath, mountID)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	if isMutating(r.Method) && token == "" {
		h.apiErr(w, app.Error(app.CodeMountNotAuthenticated, "mount has no token; writes to this server are not allowed", http.StatusForbidden, nil))
		return
	}
	target := app.RemoteBaseURL(endpoint) + "/" + remotePath
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		h.apiErr(w, app.Error(app.CodeInternalError, err.Error(), http.StatusBadGateway, err))
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	if token != "" {
		req.Header.Set("X-API-Key", token)
	}
	resp, err := mountProxyClient.Do(req)
	if err != nil {
		h.apiErr(w, app.Error(app.CodeInternalError, "remote server unreachable: "+err.Error(), http.StatusBadGateway, err))
		return
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// apiIndex returns the ID→address index (ADR-0028).
func (h *handler) apiIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/index" {
		http.NotFound(w, r)
		return
	}
	dto, err := app.BuildIDIndex(h.cosmosPath)
	h.writeOrErr(w, dto, err)
}

// apiIndexResolve resolves a single ID to its current address (ADR-0028).
func (h *handler) apiIndexResolve(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/index/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	dto, err := app.ResolveID(h.cosmosPath, id)
	h.writeOrErr(w, dto, err)
}

// apiPing answers the discovery PING with this server's identity and peers (ADR-0025).
func (h *handler) apiPing(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/ping" {
		http.NotFound(w, r)
		return
	}
	dto, err := app.Ping(h.cosmosPath)
	h.writeOrErr(w, dto, err)
}

// apiDiscover aggregates 1-hop peer gossip into mount candidates (ADR-0025).
func (h *handler) apiDiscover(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/discover" {
		http.NotFound(w, r)
		return
	}
	dto, err := app.DiscoverServers(h.cosmosPath)
	h.writeOrErr(w, dto, err)
}
func (h *handler) writeOrErr(w http.ResponseWriter, dto any, err error) {
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}

// apiServices lists or creates services (flat model, no domains).
func (h *handler) apiServices(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/services" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.ListServices(h.cosmosPath)
		h.writeOrErr(w, dto, err)
	case http.MethodPost:
		_ = r.ParseForm()
		dto, err := app.AddServiceIn(h.cosmosPath, r.FormValue("dir"), r.FormValue("name"), r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, dto)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiDecisions lists or creates decisions (flat model, no domains).
func (h *handler) apiDecisions(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/decisions" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.ListDecisions(h.cosmosPath)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case http.MethodPost:
		var req app.CreateDecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		dto, err := app.CreateDecision(h.cosmosPath, req)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, dto)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// apiDecisionRoutes serves /api/v1/decisions/{id} and its sub-resources.
func (h *handler) apiDecisionRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/decisions/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	h.apiDecisionByID(w, r, parts[0], parts[1:])
}

func (h *handler) apiDecisionByID(w http.ResponseWriter, r *http.Request, id string, tail []string) {
	if len(tail) == 1 && tail[0] == "evaluate" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req app.EvaluateDecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		result, trace, err := app.EvaluateDecisionWithTrace(h.cosmosPath, id, req, evaluatorFromRequest(r))
		if err != nil {
			// If evaluation succeeded but persisting the trace failed we still want to
			// surface the result, since callers may treat the trace as best-effort.
			if result != nil && trace == nil {
				writeJSON(w, http.StatusOK, map[string]any{
					"result":      result,
					"trace_error": err.Error(),
				})
				return
			}
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"result": result,
			"trace":  trace,
		})
		return
	}
	if len(tail) >= 1 && tail[0] == "traces" {
		h.apiDecisionTraces(w, r, id, tail[1:])
		return
	}
	if len(tail) >= 1 && tail[0] == "versions" {
		h.apiDecisionVersions(w, r, id, tail[1:])
		return
	}
	if len(tail) >= 1 && tail[0] == "scenarios" {
		h.apiDecisionScenarios(w, r, id, tail[1:])
		return
	}
	if len(tail) == 1 && tail[0] == "definitions" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		defs, err := app.GetDecisionDefinitions(h.cosmosPath, id)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, defs)
		return
	}
	if len(tail) == 1 && tail[0] == "dmn" {
		switch r.Method {
		case http.MethodGet:
			xml, err := app.GetDecisionDMN(h.cosmosPath, id)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			_, _ = w.Write([]byte(xml))
		case http.MethodPut:
			data, err := io.ReadAll(r.Body)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			dto, err := app.UpdateDecisionDMN(h.cosmosPath, id, string(data))
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.GetDecision(h.cosmosPath, id)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case http.MethodPut:
		var req app.UpdateDecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		dto, err := app.UpdateDecision(h.cosmosPath, id, req)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case http.MethodDelete:
		if err := app.DeleteDecision(h.cosmosPath, id); err != nil {
			h.apiErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *handler) apiServiceRefs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/services/refs" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dto, err := app.AllServiceRefs(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}

func (h *handler) apiProductRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	parts := strings.Split(rest, "/")
	if len(parts) == 2 && parts[0] != "" && parts[1] == "collaboration" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		dto, err := app.GetProductCollaboration(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "processes" {
		if r.Method == http.MethodGet {
			dto, err := app.ListProductProcesses(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		if r.Method == http.MethodPost {
			var req app.CreateProcessRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.CreateProductProcess(h.cosmosPath, parts[0], req)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if len(parts) == 1 && parts[0] != "" {
		dto, err := app.GetBlueprint(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "fulfillment-services" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req app.AddFulfillmentServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		dto, err := app.AddProductFulfillmentService(h.cosmosPath, parts[0], req)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] == "fulfillment-services" {
		idx, err := strconv.Atoi(parts[2])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, app.ErrorResponse(app.Error(app.CodeInvalidInput, "fulfillment index is invalid", http.StatusBadRequest, err)))
			return
		}
		if r.Method == http.MethodDelete {
			dto, err := app.RemoveProductFulfillmentService(h.cosmosPath, parts[0], idx)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		if r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req app.UpdateFulfillmentServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		dto, err := app.UpdateProductFulfillmentService(h.cosmosPath, parts[0], idx, req)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	http.NotFound(w, r)
}

func (h *handler) apiProcessRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/processes/")
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodGet:
			dto, err := app.GetProcess(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		case http.MethodDelete:
			if err := app.DeleteProcess(h.cosmosPath, parts[0]); err != nil {
				h.apiErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "bpmn" {
		if r.Method == http.MethodGet {
			xmlText, err := app.GetProcessBPMN(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			_, _ = w.Write([]byte(xmlText))
			return
		}
		if r.Method == http.MethodPut {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
				return
			}
			dto, err := app.UpdateProcessBPMN(h.cosmosPath, parts[0], string(data))
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "tasks" {
		tasks, err := app.ProcessTasks(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": tasks, "count": len(tasks)})
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "participant" {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req app.UpdateParticipantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		dto, err := app.UpdateProcessParticipant(h.cosmosPath, parts[0], req)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "triggers" {
		switch r.Method {
		case http.MethodGet:
			triggers, err := app.GetProcessTriggers(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": triggers, "count": len(triggers)})
		case http.MethodPut:
			var req app.UpdateProcessTriggersRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.UpdateProcessTriggers(h.cosmosPath, parts[0], req)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "task-mappings" {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req app.UpdateTaskMappingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		dto, err := app.UpdateProcessTaskMappings(h.cosmosPath, parts[0], req)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "steps" {
		switch r.Method {
		case http.MethodGet:
			dto, err := app.GetProcess(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": dto.Steps, "count": len(dto.Steps)})
		case http.MethodPost:
			var req app.UpsertProcessStepRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.AddProcessStep(h.cosmosPath, parts[0], req)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] == "steps" && parts[2] != "" {
		switch r.Method {
		case http.MethodPut:
			var req app.UpsertProcessStepRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.UpdateProcessStep(h.cosmosPath, parts[0], parts[2], req)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		case http.MethodDelete:
			dto, err := app.RemoveProcessStep(h.cosmosPath, parts[0], parts[2])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	htmlNotFound(w, r)
}

// apiServiceRoutes serves /api/v1/services/{service} and its sub-resources in
// the flat (no-domain) model.
func (h *handler) apiServiceRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/services/")
	if rest == "refs" {
		h.apiServiceRefs(w, r)
		return
	}
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	service := parts[0]

	// POST /api/v1/services/{service}/{kind}/{id}/move
	if len(parts) == 4 && parts[3] == "move" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		kind := map[string]string{"capabilities": "capability", "data-objects": "data-object", "user-interfaces": "user-interface", "methods": "method"}[parts[1]]
		if kind == "" {
			htmlNotFound(w, r)
			return
		}
		var req struct {
			ToService string `json:"to_service"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.apiErr(w, app.Error(app.CodeInvalidInput, "invalid JSON body", http.StatusBadRequest, err))
			return
		}
		dto, err := app.MoveServiceElement(h.cosmosPath, service, kind, parts[2], req.ToService)
		h.writeOrErr(w, dto, err)
		return
	}

	// POST /api/v1/services/{service}/capabilities
	if len(parts) == 2 && parts[1] == "capabilities" {
		if r.Method == http.MethodPost {
			var cap model.ServiceCapability
			if err := json.NewDecoder(r.Body).Decode(&cap); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.AddServiceCapability(h.cosmosPath, service, cap)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// PUT/DELETE /api/v1/services/{service}/capabilities/{id}
	if len(parts) == 3 && parts[1] == "capabilities" {
		capID := parts[2]
		switch r.Method {
		case http.MethodPut:
			var patch model.ServiceCapability
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.UpdateServiceCapability(h.cosmosPath, service, capID, patch)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case http.MethodDelete:
			dto, err := app.RemoveServiceCapability(h.cosmosPath, service, capID)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// POST /api/v1/services/{service}/data-objects
	if len(parts) == 2 && parts[1] == "data-objects" {
		if r.Method == http.MethodPost {
			var obj model.ServiceDataObject
			if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.AddServiceDataObject(h.cosmosPath, service, obj)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// PUT/DELETE /api/v1/services/{service}/data-objects/{id}
	if len(parts) == 3 && parts[1] == "data-objects" {
		doID := parts[2]
		switch r.Method {
		case http.MethodPut:
			var patch model.ServiceDataObject
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.UpdateServiceDataObject(h.cosmosPath, service, doID, patch)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case http.MethodDelete:
			dto, err := app.RemoveServiceDataObject(h.cosmosPath, service, doID)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// POST /api/v1/services/{service}/user-interfaces
	if len(parts) == 2 && parts[1] == "user-interfaces" {
		if r.Method == http.MethodPost {
			var ui model.ServiceUserInterface
			if err := json.NewDecoder(r.Body).Decode(&ui); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.AddServiceUserInterface(h.cosmosPath, service, ui)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// PUT/DELETE /api/v1/services/{service}/user-interfaces/{id}
	if len(parts) == 3 && parts[1] == "user-interfaces" {
		uiID := parts[2]
		switch r.Method {
		case http.MethodPut:
			var patch model.ServiceUserInterface
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.UpdateServiceUserInterface(h.cosmosPath, service, uiID, patch)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case http.MethodDelete:
			dto, err := app.RemoveServiceUserInterface(h.cosmosPath, service, uiID)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// GET/POST /api/v1/services/{service}/methods
	if len(parts) == 2 && parts[1] == "methods" {
		if r.Method == http.MethodPost {
			var req struct {
				Method string `json:"method"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.AddServiceMethod(h.cosmosPath, service, req.Method)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		if r.Method == http.MethodGet {
			dto, err := app.GetService(h.cosmosPath, service)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, map[string]any{"methods": dto.Methods})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// GET/PUT/DELETE /api/v1/services/{service}/methods/{method}
	if len(parts) == 3 && parts[1] == "methods" {
		method := parts[2]
		switch r.Method {
		case http.MethodGet:
			dto, err := app.GetServiceMethod(h.cosmosPath, service, method)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case http.MethodPut:
			var req struct {
				Summary    string                  `json:"summary"`
				HTTPMethod string                  `json:"http_method"`
				Path       string                  `json:"path"`
				Parameters []model.MethodParameter `json:"parameters"`
				Headers    []model.MethodHeader    `json:"headers"`
				Security   *model.MethodSecurity   `json:"security"`
				Payload    *model.MethodPayload    `json:"payload"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			patch := model.MethodDefinition{
				Summary:    req.Summary,
				HTTPMethod: req.HTTPMethod,
				Path:       req.Path,
				Parameters: req.Parameters,
				Headers:    req.Headers,
				Security:   req.Security,
				Payload:    req.Payload,
			}
			dto, err := app.UpdateMethod(h.cosmosPath, service, method, patch)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		case http.MethodDelete:
			dto, err := app.RemoveServiceMethod(h.cosmosPath, service, method)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	// GET/PUT/DELETE /api/v1/services/{service}
	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		dto, err := app.GetService(h.cosmosPath, service)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
	case http.MethodPut:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if err := app.RenameService(h.cosmosPath, service, req.Name); err != nil {
			h.apiErr(w, err)
			return
		}
		dto, err := app.GetService(h.cosmosPath, req.Name)
		h.writeOrErr(w, dto, err)
	case http.MethodDelete:
		if err := app.DeleteService(h.cosmosPath, service); err != nil {
			h.apiErr(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
func (h *handler) apiNamespaces(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/namespaces" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dto, err := app.BuildNamespaceTree(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}

func (h *handler) apiGraph(w http.ResponseWriter, r *http.Request) {
	dto, err := app.BuildGraph(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	if r.URL.Query().Get("format") == "json" {
		writeJSON(w, 200, dto)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(dto.Content))
}
func (h *handler) apiValidate(w http.ResponseWriter, r *http.Request) {
	dto, err := app.ValidateCosmos(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}
func (h *handler) apiBlueprints(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/blueprints" && r.URL.Path != "/api/blueprints" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		var bp model.Blueprint
		if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if err := app.CreateBlueprintIn(h.cosmosPath, r.URL.Query().Get("dir"), bp); err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, bp)
		return
	}
	dto, err := app.ListBlueprints(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}
func (h *handler) apiBlueprintRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/api/v1/blueprints/"), "/api/blueprints/")
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodDelete:
			if err := app.DeleteBlueprint(h.cosmosPath, parts[0]); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"deleted": parts[0]})
		case http.MethodPatch:
			var fields map[string]string
			if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.PatchBlueprint(h.cosmosPath, parts[0], fields)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		default:
			dto, err := app.GetBlueprint(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		}
		return
	}
	// /api/v1/blueprints/{id}/publish
	if len(parts) == 2 && parts[0] != "" && parts[1] == "publish" && r.Method == http.MethodPost {
		dto, err := app.PublishBlueprint(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	// /api/v1/blueprints/{id}/validate
	if len(parts) == 2 && parts[0] != "" && parts[1] == "validate" && r.Method == http.MethodGet {
		dto, err := app.ValidateBlueprint(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
		return
	}
	// /api/v1/blueprints/{id}/requirements              POST   → add requirement
	// /api/v1/blueprints/{id}/requirements/{reqID}      PATCH  → set status (open/fulfilled)
	// /api/v1/blueprints/{id}/requirements/{reqID}      DELETE → remove requirement
	if len(parts) >= 2 && parts[0] != "" && parts[1] == "requirements" {
		if len(parts) == 2 && r.Method == http.MethodPost {
			var body struct {
				Label         string   `json:"label"`
				AttributeRefs []string `json:"attribute_refs"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label required"})
				return
			}
			dto, err := app.AddBlueprintRequirementWithAttributeRefs(h.cosmosPath, parts[0], body.Label, body.AttributeRefs)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		if len(parts) == 3 && parts[2] != "" && r.Method == http.MethodPatch {
			var body struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status required (open or fulfilled)"})
				return
			}
			dto, err := app.SetBlueprintRequirementStatus(h.cosmosPath, parts[0], parts[2], body.Status)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		if len(parts) == 3 && parts[2] != "" && r.Method == http.MethodDelete {
			dto, err := app.DeleteBlueprintRequirement(h.cosmosPath, parts[0], parts[2])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
	}
	// /api/v1/blueprints/{id}/service-blueprints        POST  → add service blueprint
	// /api/v1/blueprints/{id}/service-blueprints/{svcID} DELETE → remove service blueprint
	if len(parts) >= 2 && parts[0] != "" && parts[1] == "service-blueprints" {
		if len(parts) == 2 && r.Method == http.MethodPost {
			var body struct {
				ServiceID string `json:"service_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ServiceID == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "service_id required"})
				return
			}
			dto, err := app.AddServiceBlueprintToProduct(h.cosmosPath, parts[0], body.ServiceID)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		if len(parts) == 3 && parts[2] != "" && r.Method == http.MethodDelete {
			dto, err := app.RemoveServiceBlueprintFromProduct(h.cosmosPath, parts[0], parts[2])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
	}
	// /api/v1/blueprints/{id}/attributes                        POST   → add attribute
	// /api/v1/blueprints/{id}/attributes/{attrID}               DELETE → remove attribute
	// /api/v1/blueprints/{id}/attributes/{attrID}/rules         POST   → add rule
	// /api/v1/blueprints/{id}/attributes/{attrID}/rules/{ruleID} DELETE → remove rule
	if len(parts) >= 2 && parts[0] != "" && parts[1] == "attributes" {
		if len(parts) == 2 && r.Method == http.MethodPost {
			var body struct {
				Label      string `json:"label"`
				Type       string `json:"type"`
				Required   bool   `json:"required"`
				ServiceRef string `json:"service_ref"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label required"})
				return
			}
			dto, err := app.AddBlueprintAttributeWithServiceRef(h.cosmosPath, parts[0], body.Label, body.Type, body.Required, body.ServiceRef)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		if len(parts) == 3 && parts[2] != "" && r.Method == http.MethodDelete {
			dto, err := app.DeleteBlueprintAttribute(h.cosmosPath, parts[0], parts[2])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
		if len(parts) == 4 && parts[2] != "" && parts[3] == "rules" && r.Method == http.MethodPost {
			var body struct {
				Label string `json:"label"`
				Type  string `json:"type"`
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" || body.Type == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "label and type required"})
				return
			}
			dto, err := app.AddAttributeRule(h.cosmosPath, parts[0], parts[2], body.Label, body.Type, body.Value)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, dto)
			return
		}
		if len(parts) == 5 && parts[3] == "rules" && parts[4] != "" && r.Method == http.MethodDelete {
			dto, err := app.DeleteAttributeRule(h.cosmosPath, parts[0], parts[2], parts[4])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
	}
	http.NotFound(w, r)
}
func (h *handler) apiInstances(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/instances" && r.URL.Path != "/api/instances" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		var inst model.Instance
		if err := json.NewDecoder(r.Body).Decode(&inst); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if err := app.CreateInstance(h.cosmosPath, inst); err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, inst)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dto, err := app.ListInstances(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}
func (h *handler) apiInstanceRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/api/v1/instances/"), "/api/instances/")
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		switch r.Method {
		case http.MethodDelete:
			if err := app.DeleteInstance(h.cosmosPath, parts[0]); err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"deleted": parts[0]})
		case http.MethodPatch:
			var fields map[string]string
			if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.PatchInstance(h.cosmosPath, parts[0], fields)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		default:
			dto, err := app.GetInstance(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
		}
		return
	}
	if len(parts) == 2 && parts[0] != "" {
		switch parts[1] {
		case "compliance":
			dto, err := app.GetInstanceCompliance(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, 200, dto)
			return
		case "verify":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			dto, err := app.VerifyInstance(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		case "attribute-validation":
			dto, err := app.ValidateInstanceAttributes(h.cosmosPath, parts[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		case "attribute-values":
			if r.Method != http.MethodPatch {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var values map[string]string
			if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			dto, err := app.SetInstanceAttributeValues(h.cosmosPath, parts[0], values)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
			return
		}
	}
	http.NotFound(w, r)
}
func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func (h *handler) routes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.formPost(w, r)
		return
	}
	switch {
	case r.URL.Path == "/":
		h.dashboard(w, r)
	case r.URL.Path == "/cosmos":
		h.cosmosPage(w, r)
	case r.URL.Path == "/services":
		h.servicesPage(w, r)
	case r.URL.Path == "/namespaces":
		h.namespacesPage(w, r)
	case r.URL.Path == "/graph":
		h.graphPage(w, r)
	case r.URL.Path == "/validate":
		h.validatePage(w, r)
	case r.URL.Path == "/requirements":
		h.requirementsPage(w, r)
	case r.URL.Path == "/rules":
		h.rulesPage(w, r)
	case r.URL.Path == "/verify":
		h.verifyPage(w, r)
	case r.URL.Path == "/blueprints":
		h.blueprintsPage(w, r)
	case strings.HasPrefix(r.URL.Path, "/blueprints/"):
		h.blueprintPage(w, r, strings.TrimPrefix(r.URL.Path, "/blueprints/"))
	case r.URL.Path == "/instances":
		h.instancesPage(w, r)
	case strings.HasPrefix(r.URL.Path, "/instances/"):
		h.instancePage(w, r, strings.TrimPrefix(r.URL.Path, "/instances/"))
	case r.URL.Path == "/api":
		h.apiPage(w, r)
	case strings.HasPrefix(r.URL.Path, "/services/"):
		name := strings.TrimPrefix(r.URL.Path, "/services/")
		if name != "" && !strings.Contains(name, "/") {
			h.serviceDetailPage(w, r, name)
			return
		}
		h.errorPage(w, r, 404, "Not found", "Route not found")
	default:
		h.errorPage(w, r, 404, "Not found", "Route not found")
	}
}
func (h *handler) formPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	switch r.URL.Path {
	case "/services":
		_, err := app.AddService(h.cosmosPath, r.FormValue("name"), r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Create service failed", err.Error())
			return
		}
		http.Redirect(w, r, "/services?service="+r.FormValue("name"), 303)
	case "/products/create":
		req := app.CreateProductOfferingRequest{ID: r.FormValue("id"), Name: r.FormValue("name"), Version: r.FormValue("version"), Status: r.FormValue("status"), Summary: r.FormValue("summary"), Owner: r.FormValue("owner"), OwningDomain: r.FormValue("owning_domain")}
		dto, err := app.CreateProductOffering(h.cosmosPath, req)
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Create product failed", err.Error())
			return
		}
		http.Redirect(w, r, "/cosmos?selected=product:"+dto.ID, 303)
	case "/products/fulfillment":
		req := app.AddFulfillmentServiceRequest{ServiceRef: first(r.FormValue("service_ref"), r.FormValue("service_ref_manual")), Role: r.FormValue("role"), Required: r.FormValue("required") != "", Description: r.FormValue("description")}
		dto, err := app.AddProductFulfillmentService(h.cosmosPath, r.FormValue("product_id"), req)
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Add fulfillment service failed", err.Error())
			return
		}
		if r.FormValue("return_to") == "cosmos" {
			http.Redirect(w, r, "/cosmos?selected=product:"+dto.ID+"&expand=fulfillment#fulfillment", 303)
			return
		}
		http.Redirect(w, r, "/domains?selected=product:"+dto.ID+"#fulfillment", 303)
	case "/products/fulfillment/update":
		idx, err := strconv.Atoi(r.FormValue("index"))
		if err != nil {
			h.errorPage(w, r, http.StatusBadRequest, "Update fulfillment service failed", "fulfillment index is invalid")
			return
		}
		req := app.UpdateFulfillmentServiceRequest{ServiceRef: first(r.FormValue("service_ref"), r.FormValue("service_ref_manual")), Role: r.FormValue("role"), Required: r.FormValue("required") != "", Description: r.FormValue("description")}
		dto, err := app.UpdateProductFulfillmentService(h.cosmosPath, r.FormValue("product_id"), idx, req)
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Update fulfillment service failed", err.Error())
			return
		}
		http.Redirect(w, r, "/cosmos?selected=product:"+dto.ID+"&expand=fulfillment#fulfillment", 303)
	case "/products/fulfillment/delete":
		idx, err := strconv.Atoi(r.FormValue("index"))
		if err != nil {
			h.errorPage(w, r, http.StatusBadRequest, "Remove fulfillment service failed", "fulfillment index is invalid")
			return
		}
		dto, err := app.RemoveProductFulfillmentService(h.cosmosPath, r.FormValue("product_id"), idx)
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Remove fulfillment service failed", err.Error())
			return
		}
		http.Redirect(w, r, "/cosmos?selected=product:"+dto.ID+"&expand=fulfillment#fulfillment", 303)
	default:
		h.errorPage(w, r, 404, "Not found", "Route not found")
	}
}

func (h *handler) dashboard(w http.ResponseWriter, r *http.Request) {
	co, err := app.GetCosmos(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Cosmos missing", err.Error())
		return
	}
	h.page(w, "index", map[string]any{"ActiveNav": "dashboard", "PageTitle": "Dashboard", "Cosmos": co})
}
func (h *handler) cosmosPage(w http.ResponseWriter, r *http.Request) {
	co, err := app.GetCosmos(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Cosmos missing", err.Error())
		return
	}
	doc, _ := app.DoctorCosmos(h.cosmosPath)
	ns, _ := app.BuildExplorerTreeForHost(h.cosmosPath, r.Host)
	h.page(w, "cosmos", map[string]any{
		"ActiveNav": "cosmos", "PageTitle": "Cosmos",
		"Cosmos": co, "Doctor": doc,
		"NamespaceTree": ns,
	})
}
func (h *handler) servicesPage(w http.ResponseWriter, r *http.Request) {
	services, err := app.ListServices(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Services unavailable", err.Error())
		return
	}
	svc := r.URL.Query().Get("service")
	h.page(w, "services", map[string]any{"ActiveNav": "services", "PageTitle": "Services", "Services": services.Services, "SelectedServiceName": svc})
}
func (h *handler) serviceDetailPage(w http.ResponseWriter, r *http.Request, service string) {
	s, err := app.GetService(h.cosmosPath, service)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Service not found", err.Error())
		return
	}
	h.page(w, "service_detail", map[string]any{"ActiveNav": "services", "PageTitle": s.Name, "ServiceDTO": s, "SourcePath": filepath.Join(s.Path, "service.yaml"), "SourceLang": "yaml"})
}
func (h *handler) namespacesPage(w http.ResponseWriter, r *http.Request) {
	ns, err := app.BuildNamespaceTree(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Namespace tree unavailable", err.Error())
		return
	}
	h.page(w, "namespaces", map[string]any{"ActiveNav": "namespaces", "PageTitle": "Namespace Tree", "NamespaceTree": ns})
}
func (h *handler) graphPage(w http.ResponseWriter, r *http.Request) {
	g, err := app.BuildGraph(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Graph unavailable", err.Error())
		return
	}
	h.page(w, "graph", map[string]any{"ActiveNav": "graph", "PageTitle": "Graph", "Mermaid": g.Content})
}
func (h *handler) validatePage(w http.ResponseWriter, r *http.Request) {
	val, err := app.ValidateCosmos(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Validation failed", err.Error())
		return
	}
	var errFindings, warnFindings, infoFindings []app.FindingDTO
	for _, f := range val.Findings {
		switch f.Severity {
		case "error":
			errFindings = append(errFindings, f)
		case "warning":
			warnFindings = append(warnFindings, f)
		default:
			infoFindings = append(infoFindings, f)
		}
	}
	h.page(w, "validate", map[string]any{
		"ActiveNav": "validate", "PageTitle": "Validation",
		"Validation":    val,
		"ErrorFindings": errFindings,
		"WarnFindings":  warnFindings,
		"InfoFindings":  infoFindings,
	})
}

type requirementRow struct {
	RequirementID   string
	BlueprintID     string
	BlueprintType   string
	BlueprintStatus string
}

func (h *handler) requirementsPage(w http.ResponseWriter, r *http.Request) {
	bp, err := app.ListBlueprints(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Requirements unavailable", err.Error())
		return
	}
	var rows []requirementRow
	for _, b := range bp.Blueprints {
		for _, req := range b.EvidenceRequirements {
			rows = append(rows, requirementRow{RequirementID: req, BlueprintID: b.ID, BlueprintType: b.Type, BlueprintStatus: b.Status})
		}
	}
	h.page(w, "requirements", map[string]any{"ActiveNav": "requirements", "PageTitle": "Requirements", "RequirementRows": rows})
}

type ruleRow struct {
	RuleID          string
	BlueprintID     string
	BlueprintType   string
	BlueprintStatus string
}

func (h *handler) rulesPage(w http.ResponseWriter, r *http.Request) {
	bp, err := app.ListBlueprints(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Rules unavailable", err.Error())
		return
	}
	var ruleRows, qualityRows []ruleRow
	for _, b := range bp.Blueprints {
		for _, rule := range b.Rules {
			ruleRows = append(ruleRows, ruleRow{RuleID: rule, BlueprintID: b.ID, BlueprintType: b.Type, BlueprintStatus: b.Status})
		}
		for _, qc := range b.QualityCriteria {
			qualityRows = append(qualityRows, ruleRow{RuleID: qc, BlueprintID: b.ID, BlueprintType: b.Type, BlueprintStatus: b.Status})
		}
	}
	h.page(w, "rules", map[string]any{"ActiveNav": "rules", "PageTitle": "Rules", "RuleRows": ruleRows, "QualityCriteriaRows": qualityRows})
}
func (h *handler) verifyPage(w http.ResponseWriter, r *http.Request) {
	h.page(w, "verify", map[string]any{"ActiveNav": "verify", "PageTitle": "Verification", "Verification": nil})
}
func (h *handler) blueprintsPage(w http.ResponseWriter, r *http.Request) {
	bp, err := app.ListBlueprints(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Blueprints unavailable", err.Error())
		return
	}
	h.page(w, "blueprints", map[string]any{"ActiveNav": "blueprints", "PageTitle": "Blueprints", "Blueprints": bp.Blueprints})
}
func (h *handler) blueprintPage(w http.ResponseWriter, r *http.Request, id string) {
	bp, err := app.GetBlueprint(h.cosmosPath, id)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Blueprint not found", err.Error())
		return
	}
	allBps, _ := app.ListBlueprints(h.cosmosPath)
	var serviceBps []app.BlueprintDTO
	for _, ref := range bp.RequiredServiceBlueprints {
		for _, b := range allBps.Blueprints {
			if b.ID == ref {
				serviceBps = append(serviceBps, b)
				break
			}
		}
	}
	h.page(w, "blueprint_detail", map[string]any{
		"ActiveNav":         "blueprints",
		"PageTitle":         bp.ID,
		"Blueprint":         bp,
		"AllBlueprints":     allBps.Blueprints,
		"ServiceBlueprints": serviceBps,
		"SourcePath":        bp.Path,
		"SourceLang":        "yaml",
	})
}
func (h *handler) instancesPage(w http.ResponseWriter, r *http.Request) {
	inst, err := app.ListInstances(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Instances unavailable", err.Error())
		return
	}
	bps, _ := app.ListBlueprints(h.cosmosPath)
	h.page(w, "instances", map[string]any{"ActiveNav": "instances", "PageTitle": "Instances", "Instances": inst.Instances, "Blueprints": bps.Blueprints})
}
func (h *handler) instancePage(w http.ResponseWriter, r *http.Request, id string) {
	inst, err := app.GetInstance(h.cosmosPath, id)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Instance not found", err.Error())
		return
	}
	comp, _ := app.GetInstanceCompliance(h.cosmosPath, id)
	allInst, _ := app.ListInstances(h.cosmosPath)
	var subInstances []app.InstanceDTO
	for _, si := range allInst.Instances {
		if si.OwningProductInstance == id {
			subInstances = append(subInstances, si)
		}
	}
	bps, _ := app.ListBlueprints(h.cosmosPath)
	h.page(w, "instance_detail", map[string]any{
		"ActiveNav":    "instances",
		"PageTitle":    inst.ID,
		"Instance":     inst,
		"Compliance":   comp,
		"SubInstances": subInstances,
		"Blueprints":   bps.Blueprints,
		"SourcePath":   inst.Path,
		"SourceLang":   "yaml",
	})
}
func (h *handler) apiProvisionServiceInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/product-instances/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[1] != "service-instances" || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	productID := parts[0]
	var inst model.Instance
	if err := json.NewDecoder(r.Body).Decode(&inst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	dto, err := app.ProvisionServiceInstance(h.cosmosPath, productID, inst)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}

func (h *handler) apiPage(w http.ResponseWriter, r *http.Request) {
	h.page(w, "api", map[string]any{"ActiveNav": "api", "PageTitle": "API", "Endpoints": endpointSummaries()})
}

func (h *handler) page(w http.ResponseWriter, name string, extra map[string]any) {
	co, _ := app.GetCosmos(h.cosmosPath)
	data := map[string]any{"CosmosPath": h.cosmosPath, "ShellCosmos": co, "ContentTemplate": "content_" + name, "Version": versionpkg.Get(), "Update": h.updater.Status()}
	for k, v := range extra {
		data[k] = v
	}
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
func (h *handler) errorPage(w http.ResponseWriter, r *http.Request, code int, title, msg string) {
	co, _ := app.GetCosmos(h.cosmosPath)
	w.WriteHeader(code)
	data := map[string]any{"CosmosPath": h.cosmosPath, "ShellCosmos": co, "PageTitle": title, "ActiveNav": "", "ContentTemplate": "content_error", "ErrorTitle": title, "Error": msg, "StatusCode": code, "Version": versionpkg.Get(), "Update": h.updater.Status()}
	if err := h.tmpl.ExecuteTemplate(w, "error", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
func statusOf(err error) int {
	if ae, ok := app.AsAppError(err); ok && ae.StatusCode != 0 {
		return ae.StatusCode
	}
	return 500
}
func (h *handler) apiErr(w http.ResponseWriter, err error) {
	writeJSON(w, statusOf(err), app.ErrorResponse(err))
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func htmlNotFound(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) }

// evaluatorFromRequest builds a model.Evaluator from an HTTP request.
// IP is taken from X-Forwarded-For if present, otherwise from RemoteAddr.
// User identity is taken from the X-Nomos-User header (optional).
func evaluatorFromRequest(r *http.Request) model.Evaluator {
	ip := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			ip = strings.TrimSpace(xff[:i])
		} else {
			ip = strings.TrimSpace(xff)
		}
	} else if i := strings.LastIndex(ip, ":"); i >= 0 {
		ip = ip[:i]
	}
	return model.Evaluator{
		ID:        r.Header.Get("X-Nomos-User"),
		IP:        ip,
		UserAgent: r.UserAgent(),
	}
}

func (h *handler) apiDecisionTraces(w http.ResponseWriter, r *http.Request, id string, tail []string) {
	switch {
	case len(tail) == 0:
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		dto, err := app.ListDecisionTraces(h.cosmosPath, id)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case len(tail) == 1 && tail[0] == "verify":
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		dto, err := app.VerifyDecisionTraces(h.cosmosPath, id)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case len(tail) == 1:
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tr, err := app.GetDecisionTrace(h.cosmosPath, id, tail[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tr)
	default:
		htmlNotFound(w, r)
	}
}

// apiDomainDecisionVersions serves the immutable version snapshots a decision
// accumulates as it evolves:
//
//	GET .../versions                         → list all snapshots
//	GET .../versions/{version}               → metadata at that version
//	GET .../versions/{version}/dmn           → raw DMN XML at that version
//	GET .../versions/{version}/definitions   → parsed DRG at that version
//
// The Cosmos Explorer uses the per-version DMN/definitions to render a trace's
// decision table against the rules that produced its outputs — not the current
// HEAD, which may have drifted.
func (h *handler) apiDecisionVersions(w http.ResponseWriter, r *http.Request, id string, tail []string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	switch {
	case len(tail) == 0:
		dto, err := app.ListDecisionVersions(h.cosmosPath, id)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case len(tail) == 1:
		dto, err := app.GetDecisionVersion(h.cosmosPath, id, tail[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto)
	case len(tail) == 2 && tail[1] == "dmn":
		xml, err := app.GetDecisionVersionDMN(h.cosmosPath, id, tail[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = w.Write([]byte(xml))
	case len(tail) == 2 && tail[1] == "definitions":
		defs, err := app.GetDecisionVersionDefinitions(h.cosmosPath, id, tail[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, defs)
	default:
		htmlNotFound(w, r)
	}
}

// apiDomainDecisionScenarios handles all routes under
// /api/v1/domains/{domain}/decisions/{id}/scenarios:
//
//	GET    .../scenarios                       → list
//	POST   .../scenarios                       → create (free-form or from_trace_id)
//	POST   .../scenarios/from-trace/{traceID}  → shortcut: promote a trace
//	GET    .../scenarios/{scenarioID}          → fetch one
//	PUT    .../scenarios/{scenarioID}          → partial update
//	DELETE .../scenarios/{scenarioID}          → remove
func (h *handler) apiDecisionScenarios(w http.ResponseWriter, r *http.Request, id string, tail []string) {
	switch {
	case len(tail) == 0:
		switch r.Method {
		case http.MethodGet:
			dto, err := app.ListDecisionScenarios(h.cosmosPath, id)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, dto)
		case http.MethodPost:
			var req app.CreateDecisionScenarioRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			s, err := app.CreateDecisionScenario(h.cosmosPath, id, req)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, s)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case len(tail) == 2 && tail[0] == "from-trace":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Accept name (and optional description) either in the JSON body or
		// as query params so the shortcut works for both UI fetches and curl.
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description,omitempty"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Name == "" {
			body.Name = r.URL.Query().Get("name")
		}
		if body.Description == "" {
			body.Description = r.URL.Query().Get("description")
		}
		s, err := app.CreateDecisionScenarioFromTrace(h.cosmosPath, id, tail[1], body.Name, body.Description)
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, s)
	case len(tail) == 1:
		switch r.Method {
		case http.MethodGet:
			s, err := app.GetDecisionScenario(h.cosmosPath, id, tail[0])
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, s)
		case http.MethodPut:
			var req app.UpdateDecisionScenarioRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
				return
			}
			s, err := app.UpdateDecisionScenario(h.cosmosPath, id, tail[0], req)
			if err != nil {
				h.apiErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, s)
		case http.MethodDelete:
			if err := app.DeleteDecisionScenario(h.cosmosPath, id, tail[0]); err != nil {
				h.apiErr(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	default:
		htmlNotFound(w, r)
	}
}
