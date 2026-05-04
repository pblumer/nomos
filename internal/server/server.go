package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

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
	tree, err := cosmosfs.LoadTree(h.cosmosPath)
	if err != nil {
		h.apiErr(w, 500, err)
		return
	}
	sc := 0
	for _, d := range tree.Domains {
		sc += len(d.Services)
	}
	writeJSON(w, 200, map[string]any{"path": h.cosmosPath, "id": tree.Cosmos.ID, "name": tree.Cosmos.Name, "version": tree.Cosmos.Version, "status": tree.Cosmos.Status, "owner": tree.Cosmos.Owner, "domainCount": len(tree.Domains), "serviceCount": sc})
}
func (h *handler) apiDomains(w http.ResponseWriter, r *http.Request) {
	tree, err := cosmosfs.LoadTree(h.cosmosPath)
	if err != nil {
		h.apiErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"domains": tree.Domains})
}
func (h *handler) apiGraph(w http.ResponseWriter, r *http.Request) {
	tree, err := cosmosfs.LoadTree(h.cosmosPath)
	if err != nil {
		h.apiErr(w, 500, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(graph.Mermaid(tree)))
}
func (h *handler) apiValidate(w http.ResponseWriter, r *http.Request) {
	res, _ := validate.Validate(h.cosmosPath)
	writeJSON(w, 200, res)
}
func (h *handler) routes(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		h.page(w, "index", map[string]any{"ActiveNav": "dashboard", "PageTitle": "Dashboard", "ContentTemplate": "content_index"})
		return
	}
	if r.URL.Path == "/domains" {
		h.page(w, "domains", map[string]any{"ActiveNav": "domains", "PageTitle": "Domains", "ContentTemplate": "content_domains"})
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
	if v, ok := data["ActiveNav"]; ok && v == "" {
		data["ActiveNav"] = activeNavFor("/")
	}
	for k, v := range extra {
		data[k] = v
	}
	res, _ := validate.Validate(h.cosmosPath)
	data["Validation"] = res
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
func (h *handler) apiErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"error": map[string]string{"code": "COSMOS_LOAD_FAILED", "message": err.Error()}})
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
