package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/graph"
	"github.com/nomos/nomos/internal/validate"
	versionpkg "github.com/nomos/nomos/internal/version"
)

//go:embed web/templates/* web/static/*
var webFS embed.FS

type handler struct {
	cosmosPath string
	tmpl       *template.Template
}

func NewHandler(cosmosPath string) http.Handler {
	t := template.Must(template.ParseFS(webFS, "web/templates/*.html"))
	staticFS := must(fs.Sub(webFS, "web/static"))
	h := &handler{cosmosPath: cosmosPath, tmpl: t}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/v1/cosmos", h.apiCosmos)
	mux.HandleFunc("/api/v1/domains", h.apiDomains)
	mux.HandleFunc("/api/v1/domains/", h.apiDomainRoutes)
	mux.HandleFunc("/api/v1/services/", h.apiLegacyService)
	mux.HandleFunc("/api/v1/namespaces", h.apiNamespaces)
	mux.HandleFunc("/api/v1/graph", h.apiGraph)
	mux.HandleFunc("/api/v1/validate", h.apiValidate)
	mux.HandleFunc("/", h.routes)
	return mux
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
func (h *handler) apiDomains(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/domains" {
		http.NotFound(w, r)
		return
	}
	dto, err := app.ListDomains(h.cosmosPath)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}
func (h *handler) apiDomainRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/domains/")
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && parts[0] != "" {
		dto, err := app.GetDomain(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	if len(parts) == 2 && parts[1] == "services" {
		dto, err := app.ListServices(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	if len(parts) == 3 && parts[1] == "services" && parts[2] != "" {
		dto, err := app.GetService(h.cosmosPath, parts[0], parts[2])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	http.NotFound(w, r)
}
func (h *handler) apiLegacyService(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/v1/services/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	dto, err := app.GetService(h.cosmosPath, parts[0], parts[1])
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}
func (h *handler) apiNamespaces(w http.ResponseWriter, r *http.Request) {
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
func (h *handler) routes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h.page(w, "index", map[string]any{"ActiveNav": "dashboard", "PageTitle": "Dashboard", "ContentTemplate": "content_index"})
		return
	}
	if r.URL.Path == "/domains" {
		explorer, err := buildDomainsExplorer(h.cosmosPath, r.URL.Query().Get("selected"))
		if err != nil {
			h.page(w, "domains", map[string]any{"ActiveNav": "domains", "PageTitle": "Domains Explorer", "ContentTemplate": "content_domains"})
			return
		}
		h.page(w, "domains", map[string]any{"ActiveNav": "domains", "PageTitle": "Domains Explorer", "ContentTemplate": "content_domains", "Explorer": explorer})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/domains/") {
		h.page(w, "domain_detail", map[string]any{"Domain": strings.TrimPrefix(r.URL.Path, "/domains/"), "ActiveNav": "domains", "PageTitle": "Domain", "ContentTemplate": "content_domain_detail"})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/services/") {
		p := strings.Split(strings.TrimPrefix(r.URL.Path, "/services/"), "/")
		if len(p) >= 2 {
			h.page(w, "service_detail", map[string]any{"Domain": p[0], "Service": p[1], "ActiveNav": "domains", "PageTitle": "Service", "ContentTemplate": "content_service_detail"})
			return
		}
	}
	if r.URL.Path == "/graph" {
		h.page(w, "graph", map[string]any{"ActiveNav": "graph", "PageTitle": "Graph", "ContentTemplate": "content_graph"})
		return
	}
	if r.URL.Path == "/validate" {
		h.page(w, "validate", map[string]any{"ActiveNav": "validate", "PageTitle": "Validation", "ContentTemplate": "content_validate"})
		return
	}
	http.NotFound(w, r)
}
func activeNavFor(path string) string {
	switch {
	case path == "/":
		return "dashboard"
	case path == "/domains" || strings.HasPrefix(path, "/domains/") || strings.HasPrefix(path, "/services/"):
		return "domains"
	case path == "/graph":
		return "graph"
	case path == "/validate":
		return "validate"
	default:
		return ""
	}
}
func (h *handler) page(w http.ResponseWriter, name string, extra map[string]any) {
	tree, err := cosmosfs.LoadTree(h.cosmosPath)
	if err != nil {
		w.WriteHeader(500)
		if err := h.tmpl.ExecuteTemplate(w, "error", map[string]any{"Error": err.Error(), "PageTitle": "Error", "ActiveNav": activeNavFor("/"), "ContentTemplate": "content_error", "Tree": map[string]any{"Cosmos": map[string]any{"ID": "n/a"}}}); err != nil {
			http.Error(w, err.Error(), 500)
		}
		return
	}
	sv := 0
	for _, d := range tree.Domains {
		sv += len(d.Services)
	}
	data := map[string]any{"Tree": tree, "DomainCount": len(tree.Domains), "ServiceCount": sv, "Mermaid": graph.Mermaid(tree), "CosmosPath": h.cosmosPath, "ActiveNav": activeNavFor("/"), "PageTitle": "Dashboard", "ContentTemplate": "content_index"}
	for k, v := range extra {
		data[k] = v
	}
	res, _ := validate.Validate(h.cosmosPath)
	data["Validation"] = res
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
func (h *handler) apiErr(w http.ResponseWriter, err error) {
	status := 500
	if ae, ok := app.AsAppError(err); ok && ae.StatusCode != 0 {
		status = ae.StatusCode
	}
	writeJSON(w, status, app.ErrorResponse(err))
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
