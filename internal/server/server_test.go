package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func createTestCosmos(t *testing.T) string {
	p := t.TempDir()
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "mailbox"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "rule-validation-api"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "provisioning-rules"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "home-dashboard"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "zytlog-api"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints/products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints/services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "instances/products"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: unknown\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "user-account/service.yaml"), []byte("name: user-account\ncapabilities:\n  - user-account-management\nsupported_products:\n  - PROD-ACC-MBX-001\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "mailbox/service.yaml"), []byte("name: mailbox\nsla:\n  name: Mailbox SLA\n  target: 8h\n  availability: business-hours\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "rule-validation-api/service.yaml"), []byte("name: rule-validation-api\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "provisioning-rules/service.yaml"), []byte("name: provisioning-rules\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "home-dashboard/service.yaml"), []byte("name: home-dashboard\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "zytlog-api/service.yaml"), []byte("name: zytlog-api\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints/products/account.yaml"), []byte("id: PB-ACC-MBX-001\ntype: product_blueprint\nname: Benutzerkonto mit Mailbox\nversion: 0.1.0\nstatus: draft\nowner: Team\nrequired_inputs:\n  - person_reference\nrequired_service_blueprints:\n  - SB-1\nrequired_services:\n  - service_ref: user-account\n    service_blueprint_ref: SB-1\n    required: true\nfulfillment:\n  required_services:\n    - service_ref: user-account\n      role: primary\n      required: true\n      description: Creates the account.\n      ola:\n        name: Identity Account OLA\n        target: 4h\n        availability: business-hours\n    - service_ref: mailbox\n      role: supporting\n      required: true\n      description: Creates the mailbox.\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints/services/account-service.yaml"), []byte("id: SB-1\ntype: service_blueprint\nname: Account Service\nversion: 0.1.0\nstatus: draft\nowner: Team\nnamespace_service_ref: user-account\ncapabilities:\n  - create_account\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "instances/products/account-instance.yaml"), []byte("id: PI-ACC-MBX-EXAMPLE-001\ntype: product_instance\nname: Beispielinstanz Benutzerkonto mit Mailbox\nblueprint_ref: PB-ACC-MBX-001\nblueprint_version: 0.1.0\ncompliance_status: compliant\nfindings: []\n"), 0o644))
	return p
}
func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}
func hasAll(t *testing.T, b string, m ...string) {
	t.Helper()
	for _, s := range m {
		if !strings.Contains(b, s) {
			t.Fatalf("missing %q", s)
		}
	}
}

func TestWebShellPagesAndAPI(t *testing.T) { /* same as before */
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Nomos", "Governance Platform", "Governance Dashboard", "app-sidebar", "app-topbar", "metric-card", "Cosmos Overview")
	}
	if rr := get(h, "/cosmos"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		body := rr.Body.String()
		hasAll(t, body, "Namespaces")
	}
	if rr := get(h, "/services"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Services", "user-account")
	}
	if rr := get(h, "/services/user-account"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "user-account", "Metadata")
	}
	if rr := get(h, "/graph"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Cosmos Graph", "graph TD", "code-block")
	}
	if rr := get(h, "/validate"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else if !(strings.Contains(rr.Body.String(), "No validation findings") || strings.Contains(rr.Body.String(), "Findings")) {
		t.Fatal("missing validation state")
	}
	if rr := get(h, "/validate"); !strings.Contains(rr.Body.String(), "badge") && !strings.Contains(rr.Body.String(), "alert") {
		t.Fatal("missing badge/alert")
	}
	if rr := get(h, "/static/app.css"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "--color-primary", ".app-sidebar", ".metric-card", ".explorer-layout", ".explorer-tree", ".details-panel")
		if !strings.Contains(rr.Header().Get("Content-Type"), "text/css") {
			t.Fatal(rr.Header().Get("Content-Type"))
		}
	}
	if rr := get(h, "/health"); rr.Code != 200 {
		t.Fatal(rr.Code)
	}
	var m map[string]any
	_ = json.Unmarshal(get(h, "/api/v1/cosmos").Body.Bytes(), &m)
	if m["name"] != "Local Cosmos" {
		t.Fatal(m)
	}
	if get(h, "/api/v1/services").Code != 200 || get(h, "/api/v1/blueprints").Code != 200 || get(h, "/api/v1/instances").Code != 200 || get(h, "/api/v1/graph").Code != 200 || get(h, "/api/v1/validate").Code != 200 {
		t.Fatal("api failed")
	}
	if get(h, "/does-not-exist").Code != 404 {
		t.Fatal()
	}
}

func TestVersionEndpointAndTopbar(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/api/v1/version")
	if rr.Code != 200 {
		t.Fatalf("version status=%d", rr.Code)
	}
	var payload struct {
		Version map[string]any `json:"version"`
		Update  map[string]any `json:"update"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("version json invalid: %v", err)
	}
	if payload.Version["name"] != "nomos" {
		t.Fatalf("unexpected version payload: %#v", payload.Version)
	}
	if _, ok := payload.Version["version"]; !ok {
		t.Fatal("version field missing")
	}
	// Update check is opt-in: disabled unless NOMOS_UPDATE_CHECK is set.
	if payload.Update["enabled"] != false {
		t.Fatalf("expected update check disabled by default, got %#v", payload.Update["enabled"])
	}

	// The running version is rendered in the web shell topbar.
	if body := get(h, "/").Body.String(); !strings.Contains(body, "version-badge") {
		t.Fatal("topbar version badge missing from web shell")
	}
}

func TestOpenAPIAndSwaggerRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	openapi := get(h, "/openapi.json")
	if openapi.Code != http.StatusOK {
		t.Fatalf("openapi status=%d body=%s", openapi.Code, openapi.Body.String())
	}
	if !strings.Contains(openapi.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("openapi content-type=%s", openapi.Header().Get("Content-Type"))
	}
	var spec map[string]any
	if err := json.Unmarshal(openapi.Body.Bytes(), &spec); err != nil {
		t.Fatalf("openapi json invalid: %v", err)
	}
	if spec["openapi"] != "3.1.0" {
		t.Fatalf("unexpected openapi version: %v", spec["openapi"])
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("missing paths: %#v", spec["paths"])
	}
	for _, path := range []string{"/health", "/api/v1/cosmos", "/api/v1/services", "/api/v1/blueprints", "/api/v1/instances/{instance}/compliance"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing openapi path %s", path)
		}
	}

	swagger := get(h, "/swagger")
	if swagger.Code != http.StatusOK {
		t.Fatalf("swagger status=%d", swagger.Code)
	}
	hasAll(t, swagger.Body.String(), "SwaggerUIBundle", "/openapi.json", "Nomos API")

	apiPage := get(h, "/api")
	if apiPage.Code != http.StatusOK {
		t.Fatalf("api page status=%d", apiPage.Code)
	}
	hasAll(t, apiPage.Body.String(), "Open Swagger UI", "/openapi.json", "/swagger", "endpoint-table")
}

func TestMissingCosmosStyledError(t *testing.T) {
	h := NewHandler(t.TempDir())
	rr := get(h, "/")
	if rr.Code == 200 {
		t.Fatal("expected non-200")
	}
	hasAll(t, rr.Body.String(), "Nomos", "Error", "Back to dashboard")
	if !strings.Contains(get(h, "/api/v1/cosmos").Body.String(), "cosmos.yaml") {
		t.Fatal()
	}
}

func TestNamespaceTreeAPIAndContentTypes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	jsonEndpoints := []string{"/api/v1/cosmos", "/api/v1/services", "/api/v1/namespaces", "/api/v1/validate"}
	for _, endpoint := range jsonEndpoints {
		rr := get(h, endpoint)
		if rr.Code != 200 {
			t.Fatalf("%s status=%d", endpoint, rr.Code)
		}
		if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("%s content-type=%s", endpoint, rr.Header().Get("Content-Type"))
		}
	}
	graph := get(h, "/api/v1/graph")
	if graph.Code != 200 || !strings.Contains(graph.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("graph status=%d content-type=%s", graph.Code, graph.Header().Get("Content-Type"))
	}

	rr := get(h, "/api/v1/namespaces")
	body := rr.Body.String()
	hasAll(t, body, "Services", "Decisions", "Products", "user-account", "mailbox")
}

func TestRepositoriesAPIReturnsLocalDefault(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/repositories")
	if rr.Code != 200 {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("content-type=%s", rr.Header().Get("Content-Type"))
	}
	var out struct {
		Repositories []struct {
			ID       string `json:"id"`
			Kind     string `json:"kind"`
			Name     string `json:"name"`
			Location string `json:"location"`
		} `json:"repositories"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(out.Repositories))
	}
	r := out.Repositories[0]
	if r.ID != "default" || r.Kind != "filesystem" || r.Name != "Local Cosmos" {
		t.Fatalf("unexpected repository %+v", r)
	}
	if get(h, "/api/v1/repositories/").Code == 200 {
		t.Fatal("trailing-slash path should not be served by exact handler in PR 1")
	}
}

func TestMountsAPICRUD(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/api/v1/mounts"); rr.Code != 200 {
		t.Fatalf("list status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), `"local":true`)
	}
	if rr := postJSON(h, "/api/v1/mounts", `{"endpoint":"nomos.blumer.cloud:7373","label":"Prod"}`); rr.Code != 201 {
		t.Fatalf("add status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postJSON(h, "/api/v1/mounts", `{"endpoint":"nomos.blumer.cloud:7373"}`); rr.Code != 409 {
		t.Fatalf("duplicate status=%d, want 409", rr.Code)
	}
	if rr := get(h, "/api/v1/mounts"); !strings.Contains(rr.Body.String(), "nomos.blumer.cloud:7373") {
		t.Fatalf("listed mounts missing remote: %s", rr.Body.String())
	}
	del := func(path string) int {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, path, nil))
		return rr.Code
	}
	if code := del("/api/v1/mounts/nomos-blumer-cloud-7373"); code != 204 {
		t.Fatalf("delete status=%d, want 204", code)
	}
	if code := del("/api/v1/mounts/local"); code != 400 {
		t.Fatalf("delete local status=%d, want 400", code)
	}
	if code := del("/api/v1/mounts/does-not-exist"); code != 404 {
		t.Fatalf("delete missing status=%d, want 404", code)
	}
}

func TestMountProxyForwardsTokenAndBlocksUnauthenticated(t *testing.T) {
	var gotKey, gotMethod string
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer remote.Close()
	endpoint := strings.TrimPrefix(remote.URL, "http://")

	h := NewHandler(createTestCosmos(t))
	if rr := postJSON(h, "/api/v1/mounts", `{"endpoint":"`+endpoint+`","token":"secret"}`); rr.Code != 201 {
		t.Fatalf("add authenticated mount status=%d body=%s", rr.Code, rr.Body.String())
	}
	var list struct {
		Mounts []struct {
			ID            string `json:"id"`
			Local         bool   `json:"local"`
			Authenticated bool   `json:"authenticated"`
		} `json:"mounts"`
	}
	if err := json.Unmarshal(get(h, "/api/v1/mounts").Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	var id string
	for _, m := range list.Mounts {
		if !m.Local {
			id = m.ID
			if !m.Authenticated {
				t.Fatal("expected mount to report authenticated")
			}
		}
	}
	if id == "" {
		t.Fatal("remote mount id not found")
	}
	if rr := postJSON(h, "/api/v1/mounts/"+id+"/r/api/v1/domains", `{"dns":"x.example"}`); rr.Code != 201 {
		t.Fatalf("proxy POST status=%d body=%s", rr.Code, rr.Body.String())
	}
	if gotKey != "secret" || gotMethod != "POST" {
		t.Fatalf("proxy did not forward correctly: key=%q method=%q", gotKey, gotMethod)
	}
	if rr := get(h, "/api/v1/mounts/"+id+"/r/api/v1/cosmos"); rr.Code != 201 {
		t.Fatalf("proxy GET status=%d", rr.Code)
	}
	if gotMethod != "GET" {
		t.Fatalf("read not forwarded as GET: %q", gotMethod)
	}

	if rr := postJSON(h, "/api/v1/mounts", `{"endpoint":"127.0.0.1:9"}`); rr.Code != 201 {
		t.Fatalf("add unauthenticated mount status=%d", rr.Code)
	}
	if rr := postJSON(h, "/api/v1/mounts/127-0-0-1-9/r/api/v1/domains", `{}`); rr.Code != 403 {
		t.Fatalf("write to unauthenticated mount status=%d, want 403", rr.Code)
	}
}

func TestRepositoryScopedReadsMatchAliases(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	for _, p := range [][2]string{
		{"/api/v1/cosmos", "/api/v1/repositories/default/cosmos"},
		{"/api/v1/namespaces", "/api/v1/repositories/default/namespaces"},
	} {
		alias := get(h, p[0])
		scoped := get(h, p[1])
		if alias.Code != 200 || scoped.Code != 200 {
			t.Fatalf("%s=%d %s=%d", p[0], alias.Code, p[1], scoped.Code)
		}
		if alias.Body.String() != scoped.Body.String() {
			t.Fatalf("payload mismatch:\n%s\n%s", p[0], p[1])
		}
	}
	if meta := get(h, "/api/v1/repositories/default"); meta.Code != 200 {
		t.Fatalf("repo metadata status=%d", meta.Code)
	} else {
		hasAll(t, meta.Body.String(), `"id":"default"`, "filesystem")
	}
	if rr := get(h, "/api/v1/repositories/nope/cosmos"); rr.Code != 404 {
		t.Fatalf("unknown repo status=%d, want 404", rr.Code)
	}
}

func TestRESTDetailRoutesAndErrors(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/api/v1/services"); rr.Code != 200 {
		t.Fatalf("services status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "user-account")
	}
	if rr := get(h, "/api/v1/services/user-account"); rr.Code != 200 {
		t.Fatalf("service status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "user-account")
	}
	if rr := get(h, "/api/v1/services/does-not-exist"); rr.Code != 404 || !strings.Contains(rr.Body.String(), "SERVICE_NOT_FOUND") {
		t.Fatalf("expected service 404, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestBlueprintAndInstanceAPIRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/api/v1/blueprints"); rr.Code != 200 {
		t.Fatalf("blueprints status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "PB-ACC-MBX-001", "Benutzerkonto mit Mailbox", "fulfillment", "resolution_status", "user-account")
	}
	if rr := get(h, "/api/blueprints/PB-ACC-MBX-001"); rr.Code != 200 {
		t.Fatalf("blueprint detail status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "product_blueprint", "fulfillment", "user-account")
	}
	if rr := get(h, "/api/v1/blueprints/SB-1"); rr.Code != 200 {
		t.Fatalf("service blueprint detail status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "service_blueprint", "namespace_service_ref", "user-account")
	}
	if rr := get(h, "/api/v1/instances"); rr.Code != 200 {
		t.Fatalf("instances status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "PI-ACC-MBX-EXAMPLE-001")
	}
	if rr := get(h, "/api/instances/PI-ACC-MBX-EXAMPLE-001/compliance"); rr.Code != 200 {
		t.Fatalf("compliance status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "compliant")
	}
}

func TestExtendedWebPages(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	pages := map[string][]string{
		"/cosmos":                           {"Cosmos"},
		"/services":                         {"Services", "Create service"},
		"/namespaces":                       {"Namespace Tree", "user-account"},
		"/blueprints":                       {"Blueprints", "PB-ACC-MBX-001", "SB-1"},
		"/blueprints/PB-ACC-MBX-001":        {"PB-ACC-MBX-001", "Required inputs", "user-account"},
		"/instances":                        {"Instances", "PI-ACC-MBX-EXAMPLE-001"},
		"/instances/PI-ACC-MBX-EXAMPLE-001": {"Compliance", "compliant"},
		"/api":                              {"/health", "/api/v1/cosmos", "/api/v1/instances"},
	}
	for path, want := range pages {
		rr := get(h, path)
		if rr.Code != 200 {
			t.Fatalf("%s status=%d body=%s", path, rr.Code, rr.Body.String())
		}
		hasAll(t, rr.Body.String(), want...)
	}
}

func TestWebErrorPagesForMissingResources(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	paths := []string{"/services/missing-service", "/blueprints/missing", "/instances/missing"}
	for _, path := range paths {
		rr := get(h, path)
		if rr.Code != 404 {
			t.Fatalf("%s status=%d", path, rr.Code)
		}
		hasAll(t, rr.Body.String(), "Nomos", "Status 404", "Back to dashboard")
	}
}

func postForm(h http.Handler, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)
	return rr
}
func postJSON(h http.Handler, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}
func patchJSON(h http.Handler, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}
func deleteReq(h http.Handler, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	h.ServeHTTP(rr, req)
	return rr
}

func TestServiceBlueprintManagementAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// add service blueprint SB-1 to product PB-ACC-MBX-001
	rr := postJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001/service-blueprints", `{"service_id":"SB-1"}`)
	if rr.Code != 200 {
		t.Fatalf("add service blueprint status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "SB-1", "required_service_blueprints")

	// adding again is idempotent
	rr = postJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001/service-blueprints", `{"service_id":"SB-1"}`)
	if rr.Code != 200 {
		t.Fatalf("idempotent add status=%d", rr.Code)
	}

	// missing service_id → 400
	rr = postJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001/service-blueprints", `{}`)
	if rr.Code != 400 {
		t.Fatalf("expected 400 for missing service_id, got %d", rr.Code)
	}

	// non-existent blueprint → 404
	rr = postJSON(h, "/api/v1/blueprints/DOES-NOT-EXIST/service-blueprints", `{"service_id":"SB-1"}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for non-existent blueprint, got %d", rr.Code)
	}

	// remove service blueprint
	rr = deleteReq(h, "/api/v1/blueprints/PB-ACC-MBX-001/service-blueprints/SB-1")
	if rr.Code != 200 {
		t.Fatalf("remove service blueprint status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestBlueprintRequirementsAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	bpPath := "/api/v1/blueprints/PB-ACC-MBX-001/requirements"

	// add requirement
	rr := postJSON(h, bpPath, `{"label":"Datenschutzkonzept vorhanden"}`)
	if rr.Code != 200 {
		t.Fatalf("add requirement status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "Datenschutzkonzept vorhanden", "requirements")

	// parse out the requirement ID — all requirement IDs are prefixed with "req-"
	body := rr.Body.String()
	idMarker := `"id":"req-`
	idStart := strings.Index(body, idMarker)
	if idStart == -1 {
		t.Fatalf("could not find requirement id in response: %s", body)
	}
	idStart += len(`"id":"`)
	idEnd := strings.Index(body[idStart:], `"`) + idStart
	reqID := body[idStart:idEnd]
	if reqID == "" {
		t.Fatal("could not extract requirement ID from response")
	}

	// missing label → 400
	rr = postJSON(h, bpPath, `{"label":""}`)
	if rr.Code != 400 {
		t.Fatalf("expected 400 for empty label, got %d", rr.Code)
	}

	// set status to fulfilled
	rr = patchJSON(h, bpPath+"/"+reqID, `{"status":"fulfilled"}`)
	if rr.Code != 200 {
		t.Fatalf("set fulfilled status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"status":"fulfilled"`) {
		t.Fatalf("expected status:fulfilled in response: %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"requirements_status":"fulfilled"`) {
		t.Fatalf("expected requirements_status:fulfilled in response: %s", rr.Body.String())
	}

	// set status back to open
	rr = patchJSON(h, bpPath+"/"+reqID, `{"status":"open"}`)
	if rr.Code != 200 {
		t.Fatalf("set open status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"status":"open"`) {
		t.Fatalf("expected status:open in response: %s", rr.Body.String())
	}

	// invalid status → 400
	rr = patchJSON(h, bpPath+"/"+reqID, `{"status":"invalid"}`)
	if rr.Code != 400 {
		t.Fatalf("expected 400 for invalid status, got %d", rr.Code)
	}

	// missing status → 400
	rr = patchJSON(h, bpPath+"/"+reqID, `{}`)
	if rr.Code != 400 {
		t.Fatalf("expected 400 for missing status, got %d", rr.Code)
	}

	// patch non-existent requirement → 404
	rr = patchJSON(h, bpPath+"/req-nonexistent", `{"status":"fulfilled"}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for non-existent requirement, got %d", rr.Code)
	}

	// delete requirement
	rr = deleteReq(h, bpPath+"/"+reqID)
	if rr.Code != 200 {
		t.Fatalf("delete requirement status=%d body=%s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "Datenschutzkonzept vorhanden") {
		t.Fatal("deleted requirement should not appear in response")
	}

	// non-existent blueprint → 404
	rr = postJSON(h, "/api/v1/blueprints/DOES-NOT-EXIST/requirements", `{"label":"test"}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for non-existent blueprint, got %d", rr.Code)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestCreateService(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)
	if rr := postForm(h, "/api/v1/services", "name=web-ui&owner=Web"); rr.Code != http.StatusCreated {
		t.Fatalf("service create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/api/v1/services", "name=web-ui"); rr.Code != http.StatusConflict {
		t.Fatalf("duplicate service status=%d", rr.Code)
	}
	if rr := get(h, "/services"); rr.Code != 200 || !strings.Contains(rr.Body.String(), "web-ui") {
		t.Fatalf("service page status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAPIBlueprintsPOST(t *testing.T) {
	p := t.TempDir()
	mustMkdir(t, storage.NomosDir(p))
	if err := os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products"))
	mustMkdir(t, filepath.Join(storage.CatalogDir(p), "blueprints", "services"))
	h := NewHandler(p)

	body := `{"id":"PB-API-001","type":"product_blueprint","name":"API Blueprint","version":"0.1.0","status":"draft","owner":"API Team","summary":"Created via API"}`
	rr := postJSON(h, "/api/v1/blueprints", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify via GET
	rr = get(h, "/api/v1/blueprints/PB-API-001")
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), "API Blueprint") {
		t.Fatalf("blueprint get after post failed: %d %s", rr.Code, rr.Body.String())
	}

	// Duplicate should fail
	rr = postJSON(h, "/api/v1/blueprints", body)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate, got %d", rr.Code)
	}
}

func TestAPIBlueprintsDELETE(t *testing.T) {
	p := t.TempDir()
	mustMkdir(t, storage.NomosDir(p))
	if err := os.WriteFile(storage.CosmosFile(p), []byte("id: c\nname: C\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products"))
	h := NewHandler(p)

	// Create first
	body := `{"id":"PB-DEL-002","type":"product_blueprint","name":"Delete Me","version":"0.1.0","status":"draft","owner":"Team"}`
	rr := postJSON(h, "/api/v1/blueprints", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// Delete
	del := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/blueprints/PB-DEL-002", nil)
	h.ServeHTTP(del, req)
	if del.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d: %s", del.Code, del.Body.String())
	}
	if !strings.Contains(del.Body.String(), `"deleted":"PB-DEL-002"`) {
		t.Fatalf("unexpected delete response: %s", del.Body.String())
	}

	// Verify gone
	if get(h, "/api/v1/blueprints/PB-DEL-002").Code != 404 {
		t.Fatal("expected 404 after delete")
	}

	// Delete non-existent
	del2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/blueprints/PB-NONEXISTENT", nil)
	h.ServeHTTP(del2, req2)
	if del2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent, got %d", del2.Code)
	}
}

func TestAPIServiceDELETE(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)

	// Delete service
	del := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/services/user-account", nil)
	h.ServeHTTP(del, req)
	if del.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on service delete, got %d: %s", del.Code, del.Body.String())
	}
	if get(h, "/api/v1/services/user-account").Code != 404 {
		t.Fatal("expected 404 after service delete")
	}

	// Delete non-existent service
	del2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/services/does-not-exist", nil)
	h.ServeHTTP(del2, req2)
	if del2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent service, got %d", del2.Code)
	}
}

func TestInstanceAPIRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// GET single instance
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001")
	if rr.Code != 200 {
		t.Fatalf("GET instance status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "PI-ACC-MBX-EXAMPLE-001", "product_instance")

	// GET non-existent → 404
	if rr := get(h, "/api/v1/instances/DOES-NOT-EXIST"); rr.Code != 404 {
		t.Fatalf("expected 404 for missing instance, got %d", rr.Code)
	}

	// PATCH instance
	rr = patchJSON(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001", `{"name":"Renamed","status":"active"}`)
	if rr.Code != 200 {
		t.Fatalf("PATCH instance status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "Renamed", "active")

	// PATCH non-existent → 404
	rr = patchJSON(h, "/api/v1/instances/DOES-NOT-EXIST", `{"name":"x"}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for missing instance PATCH, got %d", rr.Code)
	}

	// GET /compliance
	rr = get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/compliance")
	if rr.Code != 200 {
		t.Fatalf("compliance status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "compliant")

	// POST /verify
	rr = postJSON(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/verify", "")
	if rr.Code != 200 {
		t.Fatalf("verify status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "compliant", "evidence-verify-")

	// DELETE instance
	rr = deleteReq(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001")
	if rr.Code != 200 {
		t.Fatalf("DELETE instance status=%d body=%s", rr.Code, rr.Body.String())
	}

	// deleted instance → 404
	if rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001"); rr.Code != 404 {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}
}

func TestBlueprintPatchPublishValidateAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// PATCH blueprint
	rr := patchJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001", `{"name":"Renamed BP","status":"active"}`)
	if rr.Code != 200 {
		t.Fatalf("PATCH blueprint status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "Renamed BP", "active")

	// PATCH non-existent → 404
	rr = patchJSON(h, "/api/v1/blueprints/DOES-NOT-EXIST", `{"name":"x"}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for PATCH non-existent blueprint, got %d", rr.Code)
	}

	// POST /publish
	rr = postJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001/publish", "")
	if rr.Code != 200 {
		t.Fatalf("publish status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "published")

	// GET /validate
	rr = get(h, "/api/v1/blueprints/PB-ACC-MBX-001/validate")
	if rr.Code != 200 {
		t.Fatalf("validate status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "status")
}

func TestProvisionServiceInstanceAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// Provision a service instance under an existing product instance
	rr := postJSON(h, "/api/v1/product-instances/PI-ACC-MBX-EXAMPLE-001/service-instances",
		`{"id":"SI-TEST-001","blueprint_ref":"SB-1","blueprint_version":"0.1.0","status":"draft","owner":"Team"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("provision status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "SI-TEST-001", "service_instance", "PI-ACC-MBX-EXAMPLE-001")

	// Provision to non-existent product → 404
	rr = postJSON(h, "/api/v1/product-instances/DOES-NOT-EXIST/service-instances",
		`{"id":"SI-TEST-002","blueprint_ref":"SB-1","blueprint_version":"0.1.0","status":"draft","owner":"Team"}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for missing product, got %d", rr.Code)
	}

	// Wrong path → 404
	rr = postJSON(h, "/api/v1/product-instances/PI-ACC-MBX-EXAMPLE-001/wrong-path", `{}`)
	if rr.Code != 404 {
		t.Fatalf("expected 404 for wrong sub-path, got %d", rr.Code)
	}
}

func TestBlueprintAndInstanceWebPages(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// blueprint detail page
	rr := get(h, "/blueprints/SB-1")
	if rr.Code != 200 {
		t.Fatalf("/blueprints/SB-1 status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "SB-1", "service_blueprint")

	// instance detail page
	rr = get(h, "/instances/PI-ACC-MBX-EXAMPLE-001")
	if rr.Code != 200 {
		t.Fatalf("/instances/PI-ACC-MBX-EXAMPLE-001 status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "PI-ACC-MBX-EXAMPLE-001")
}

func TestFlatServiceAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// valid flat service lookup
	rr := get(h, "/api/v1/services/user-account")
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "user-account")

	// non-existent service → 404
	rr = get(h, "/api/v1/services/no-such-service")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for missing service, got %d", rr.Code)
	}
}

func TestRequirementsAndRulesPages(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/requirements")
	if rr.Code != 200 {
		t.Fatalf("/requirements status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "Requirements")

	rr = get(h, "/rules")
	if rr.Code != 200 {
		t.Fatalf("/rules status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "Rules")
}

func TestFormPostRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// POST /services → redirects
	rr3 := httptest.NewRecorder()
	body3 := strings.NewReader("name=new-service&owner=Team")
	req3 := httptest.NewRequest(http.MethodPost, "/services", body3)
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr3, req3)
	if rr3.Code != 303 {
		t.Fatalf("POST /services expected 303 redirect, got %d body=%s", rr3.Code, rr3.Body.String())
	}

	// POST unknown route → error page
	rr5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPost, "/unknown-route", nil)
	h.ServeHTTP(rr5, req5)
	if rr5.Code == 303 {
		t.Fatalf("POST /unknown-route should not redirect")
	}
}

func TestHTMLNotFound(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/no-such-page-xyz")
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestNamespacesAndGraphPages(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/namespaces")
	if rr.Code != 200 {
		t.Fatalf("/namespaces status=%d", rr.Code)
	}

	rr = get(h, "/graph")
	if rr.Code != 200 {
		t.Fatalf("/graph status=%d", rr.Code)
	}
	hasAll(t, rr.Body.String(), "graph")
}

func TestAPINamespacesAndValidate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/api/v1/namespaces")
	if rr.Code != 200 {
		t.Fatalf("/api/v1/namespaces status=%d", rr.Code)
	}

	// wrong path → 404
	rr = get(h, "/api/v1/namespaces/extra")
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	// method not allowed
	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces", nil)
	h.ServeHTTP(rr2, req)
	if rr2.Code != 405 {
		t.Fatalf("expected 405, got %d", rr2.Code)
	}

	rr = get(h, "/api/v1/validate")
	if rr.Code != 200 {
		t.Fatalf("/api/v1/validate status=%d", rr.Code)
	}
}

func TestAPIGraphFormats(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// default format (mermaid text)
	rr := get(h, "/api/v1/graph")
	if rr.Code != 200 {
		t.Fatalf("/api/v1/graph status=%d", rr.Code)
	}

	// json format
	rr = get(h, "/api/v1/graph?format=json")
	if rr.Code != 200 {
		t.Fatalf("/api/v1/graph?format=json status=%d", rr.Code)
	}
	hasAll(t, rr.Body.String(), "mermaid")
}

func TestOpenAPIAndSwaggerUI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/openapi.json")
	if rr.Code != 200 {
		t.Fatalf("/openapi.json status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "openapi")

	rr = get(h, "/swagger")
	if rr.Code != 200 {
		t.Fatalf("/swagger status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstancesListAndCreate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/api/v1/instances")
	if rr.Code != 200 {
		t.Fatalf("/api/v1/instances status=%d body=%s", rr.Code, rr.Body.String())
	}

	// create a new product instance
	rr = postJSON(h, "/api/v1/instances",
		`{"id":"PI-NEW-TEST","type":"product_instance","name":"New","blueprint_ref":"PB-ACC-MBX-001","blueprint_version":"0.1.0","status":"draft","owner":"Team"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create instance status=%d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "PI-NEW-TEST")

	// invalid json → 400
	rr = postJSON(h, "/api/v1/instances", `{bad json}`)
	if rr.Code != 400 {
		t.Fatalf("expected 400 for bad JSON, got %d", rr.Code)
	}
}

func TestCosmosInlineScriptKeepsBusinessRuleDecisionTernaryComplete(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/cosmos")
	if rr.Code != http.StatusOK {
		t.Fatalf("cosmos status=%d", rr.Code)
	}
	body := rr.Body.String()
	if strings.Contains(body, "}]},gateway:task_type==='businessRuleTask'?") {
		t.Fatalf("cosmos inline script contains an incomplete decision ternary before the gateway payload")
	}
	for _, want := range []string{
		"function buildBusinessRuleDecision(taskType)",
		"function buildBusinessRuleGateway(taskType)",
		"decision:buildBusinessRuleDecision(task_type)",
		"gateway:buildBusinessRuleGateway(task_type)",
		"body:JSON.stringify(payload)",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("cosmos inline script missing %q", want)
		}
	}
}

func TestCosmosGeneratedBPMNLayoutUsesShapeBoundsForSequenceFlows(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/cosmos")
	if rr.Code != http.StatusOK {
		t.Fatalf("cosmos status=%d", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		"const layoutById={};",
		"const sequenceGap=72;",
		"{w:150,h:88,y:76}",
		"const x1=source.x+source.w;",
		"const x2=target.x;",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("cosmos generated BPMN layout missing %q", want)
		}
	}
}
