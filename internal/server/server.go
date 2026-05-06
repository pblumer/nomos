package server

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
	versionpkg "github.com/nomos/nomos/internal/version"
)

//go:embed web/templates/* web/static/*
var webFS embed.FS

type handler struct {
	cosmosPath string
	tmpl       *template.Template
}

func NewHandler(cosmosPath string) http.Handler {
	t := template.Must(template.New("web").ParseFS(webFS, "web/templates/*.html"))
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
	mux.HandleFunc("/api/v1/blueprints", h.apiBlueprints)
	mux.HandleFunc("/api/v1/blueprints/", h.apiBlueprintRoutes)
	mux.HandleFunc("/api/v1/instances", h.apiInstances)
	mux.HandleFunc("/api/v1/instances/", h.apiInstanceRoutes)
	mux.HandleFunc("/api/v1/verify/domain/", h.apiVerifyDomain)
	mux.HandleFunc("/api/blueprints", h.apiBlueprints)
	mux.HandleFunc("/api/blueprints/", h.apiBlueprintRoutes)
	mux.HandleFunc("/api/instances", h.apiInstances)
	mux.HandleFunc("/api/instances/", h.apiInstanceRoutes)
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
	if r.Method == http.MethodPost {
		h.createDomain(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	if len(parts) == 1 && parts[0] != "" && r.Method == http.MethodGet {
		dto, err := app.GetDomain(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	if len(parts) == 2 && parts[1] == "services" {
		if r.Method == http.MethodPost {
			h.createService(w, r, parts[0])
			return
		}
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
	htmlNotFound(w, r)
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
		if err := app.CreateBlueprint(h.cosmosPath, bp); err != nil {
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
		dto, err := app.GetBlueprint(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	http.NotFound(w, r)
}
func (h *handler) apiInstances(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/instances" && r.URL.Path != "/api/instances" {
		http.NotFound(w, r)
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
		dto, err := app.GetInstance(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] == "compliance" {
		dto, err := app.GetInstanceCompliance(h.cosmosPath, parts[0])
		if err != nil {
			h.apiErr(w, err)
			return
		}
		writeJSON(w, 200, dto)
		return
	}
	http.NotFound(w, r)
}
func (h *handler) apiVerifyDomain(w http.ResponseWriter, r *http.Request) {
	domain := strings.TrimPrefix(r.URL.Path, "/api/v1/verify/domain/")
	if domain == "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	dto, err := app.VerifyDomain(r.Context(), h.cosmosPath, domain)
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, 200, dto)
}

func (h *handler) createDomain(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	dto, err := app.AddDomain(h.cosmosPath, first(r.FormValue("dns"), r.FormValue("domain")), r.FormValue("owner"), r.FormValue("force") != "")
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto)
}
func (h *handler) createService(w http.ResponseWriter, r *http.Request, domain string) {
	_ = r.ParseForm()
	dto, err := app.AddService(h.cosmosPath, domain, r.FormValue("name"), r.FormValue("owner"), r.FormValue("force") != "")
	if err != nil {
		h.apiErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto)
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
	case r.URL.Path == "/domains":
		h.domainsPage(w, r)
	case strings.HasPrefix(r.URL.Path, "/domains/"):
		h.domainPage(w, r, strings.TrimPrefix(r.URL.Path, "/domains/"))
	case r.URL.Path == "/services":
		h.servicesPage(w, r)
	case r.URL.Path == "/namespaces":
		h.namespacesPage(w, r)
	case r.URL.Path == "/graph":
		h.graphPage(w, r)
	case r.URL.Path == "/validate":
		h.validatePage(w, r)
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
		p := strings.Split(strings.TrimPrefix(r.URL.Path, "/services/"), "/")
		if len(p) >= 2 {
			h.serviceDetailPage(w, r, p[0], p[1])
			return
		}
		h.errorPage(w, r, 404, "Not found", "Route not found")
	default:
		h.errorPage(w, r, 404, "Not found", "Route not found")
	}
}
func (h *handler) formPost(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if h.contextualFormPost(w, r) {
		return
	}
	switch r.URL.Path {
	case "/domains":
		_, err := app.AddDomain(h.cosmosPath, r.FormValue("dns"), r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Create domain failed", err.Error())
			return
		}
		http.Redirect(w, r, "/domains?selected=domain:"+r.FormValue("dns"), 303)
	case "/services":
		_, err := app.AddService(h.cosmosPath, r.FormValue("domain"), r.FormValue("name"), r.FormValue("owner"), r.FormValue("force") != "")
		if err != nil {
			h.errorPage(w, r, statusOf(err), "Create service failed", err.Error())
			return
		}
		http.Redirect(w, r, "/services?domain="+r.FormValue("domain")+"&service="+r.FormValue("name"), 303)
	case "/verify":
		_, err := app.VerifyDomain(r.Context(), h.cosmosPath, r.FormValue("domain"))
		if err != nil {
			http.Redirect(w, r, "/verify", 303)
			return
		}
		http.Redirect(w, r, "/verify", 303)
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
	val, _ := app.ValidateCosmos(h.cosmosPath)
	bp, _ := app.ListBlueprints(h.cosmosPath)
	inst, _ := app.ListInstances(h.cosmosPath)
	h.page(w, "index", map[string]any{"ActiveNav": "dashboard", "PageTitle": "Dashboard", "Cosmos": co, "Validation": val, "BlueprintCount": bp.Count, "InstanceCount": inst.Count})
}
func (h *handler) cosmosPage(w http.ResponseWriter, r *http.Request) {
	co, err := app.GetCosmos(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Cosmos missing", err.Error())
		return
	}
	doc, _ := app.DoctorCosmos(h.cosmosPath)
	h.page(w, "cosmos", map[string]any{"ActiveNav": "cosmos", "PageTitle": "Cosmos", "Cosmos": co, "Doctor": doc})
}
func (h *handler) domainsPage(w http.ResponseWriter, r *http.Request) {
	h.renderDomainsPage(w, r, http.StatusOK, domainFormState{})
}

func (h *handler) domainPage(w http.ResponseWriter, r *http.Request, domain string) {
	d, err := app.GetDomain(h.cosmosPath, domain)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Domain not found", err.Error())
		return
	}
	h.page(w, "domain_detail", map[string]any{"ActiveNav": "domains", "PageTitle": d.Canonical, "DomainDTO": d})
}
func (h *handler) servicesPage(w http.ResponseWriter, r *http.Request) {
	domains, err := app.ListDomains(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Services unavailable", err.Error())
		return
	}
	selected := first(r.URL.Query().Get("domain"))
	if selected == "" && len(domains.Domains) > 0 {
		selected = domains.Domains[0].Canonical
	}
	var services app.ServicesDTO
	if selected != "" {
		services, _ = app.ListServices(h.cosmosPath, selected)
	}
	svc := r.URL.Query().Get("service")
	h.page(w, "services", map[string]any{"ActiveNav": "services", "PageTitle": "Services", "Domains": domains.Domains, "SelectedDomain": selected, "Services": services.Services, "SelectedServiceName": svc})
}
func (h *handler) serviceDetailPage(w http.ResponseWriter, r *http.Request, domain, service string) {
	s, err := app.GetService(h.cosmosPath, domain, service)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Service not found", err.Error())
		return
	}
	h.page(w, "service_detail", map[string]any{"ActiveNav": "services", "PageTitle": s.Name, "ServiceDTO": s})
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
	h.page(w, "validate", map[string]any{"ActiveNav": "validate", "PageTitle": "Validation", "Validation": val})
}
func (h *handler) verifyPage(w http.ResponseWriter, r *http.Request) {
	ev, err := app.ListVerificationEvidence(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Verification unavailable", err.Error())
		return
	}
	h.page(w, "verify", map[string]any{"ActiveNav": "verify", "PageTitle": "Verification", "Verification": ev})
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
	h.page(w, "blueprint_detail", map[string]any{"ActiveNav": "blueprints", "PageTitle": bp.ID, "Blueprint": bp})
}
func (h *handler) instancesPage(w http.ResponseWriter, r *http.Request) {
	inst, err := app.ListInstances(h.cosmosPath)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Instances unavailable", err.Error())
		return
	}
	h.page(w, "instances", map[string]any{"ActiveNav": "instances", "PageTitle": "Instances", "Instances": inst.Instances})
}
func (h *handler) instancePage(w http.ResponseWriter, r *http.Request, id string) {
	inst, err := app.GetInstance(h.cosmosPath, id)
	if err != nil {
		h.errorPage(w, r, statusOf(err), "Instance not found", err.Error())
		return
	}
	comp, _ := app.GetInstanceCompliance(h.cosmosPath, id)
	h.page(w, "instance_detail", map[string]any{"ActiveNav": "instances", "PageTitle": inst.ID, "Instance": inst, "Compliance": comp})
}
func (h *handler) apiPage(w http.ResponseWriter, r *http.Request) {
	h.page(w, "api", map[string]any{"ActiveNav": "api", "PageTitle": "API"})
}

func (h *handler) page(w http.ResponseWriter, name string, extra map[string]any) {
	co, _ := app.GetCosmos(h.cosmosPath)
	data := map[string]any{"CosmosPath": h.cosmosPath, "ShellCosmos": co, "ContentTemplate": "content_" + name}
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
	data := map[string]any{"CosmosPath": h.cosmosPath, "ShellCosmos": co, "PageTitle": title, "ActiveNav": "", "ContentTemplate": "content_error", "ErrorTitle": title, "Error": msg, "StatusCode": code}
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
