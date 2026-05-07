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
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud/services/rule-validation-api"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "blumer.com/services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "identity.blumer.com/services/user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "governance.blumer.com/services/provisioning-rules"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "blumer.cloud/services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "home.blumer.cloud/services/home-dashboard"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "zytlog.blumer.cloud/services/zytlog-api"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "beispiel.ch/services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints/products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints/services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "instances/products"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: unknown\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "domain.yaml"), []byte("name: identity.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud/domain.yaml"), []byte("name: platform.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "blumer.com/domain.yaml"), []byte("name: blumer.com\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.com/domain.yaml"), []byte("name: identity.blumer.com\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "governance.blumer.com/domain.yaml"), []byte("name: governance.blumer.com\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "blumer.cloud/domain.yaml"), []byte("name: blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "home.blumer.cloud/domain.yaml"), []byte("name: home.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "zytlog.blumer.cloud/domain.yaml"), []byte("name: zytlog.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "beispiel.ch/domain.yaml"), []byte("name: beispiel.ch\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud/services/user-account/service.yaml"), []byte("name: user-account\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud/services/rule-validation-api/service.yaml"), []byte("name: rule-validation-api\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.com/services/user-account/service.yaml"), []byte("name: user-account\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "governance.blumer.com/services/provisioning-rules/service.yaml"), []byte("name: provisioning-rules\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "home.blumer.cloud/services/home-dashboard/service.yaml"), []byte("name: home-dashboard\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "zytlog.blumer.cloud/services/zytlog-api/service.yaml"), []byte("name: zytlog-api\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints/products/account.yaml"), []byte("id: PB-ACC-MBX-001\ntype: product_blueprint\nname: Benutzerkonto mit Mailbox\nversion: 0.1.0\nstatus: draft\nowner: Team\nrequired_inputs:\n  - person_reference\nrequired_service_blueprints:\n  - SB-1\nrequired_services:\n  - service_ref: identity.blumer.cloud/user-account\n    service_blueprint_ref: SB-1\n    required: true\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints/services/account-service.yaml"), []byte("id: SB-1\ntype: service_blueprint\nname: Account Service\nversion: 0.1.0\nstatus: draft\nowner: Team\nnamespace_service_ref: identity.blumer.cloud/user-account\ncapabilities:\n  - create_account\n"), 0o644))
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
	if rr := get(h, "/domains"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Domains Explorer", "Namespaces are top-level zones", "identity.blumer.cloud", "user-account", "explorer-layout", "explorer-tree", "details-panel", "Domain count", "Service count")
	}
	if rr := get(h, "/domains?selected=domain:identity.blumer.cloud"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Domain Details", "identity.blumer.cloud", "Service count", "user-account")
	}
	if rr := get(h, "/domains?selected=service:identity.blumer.cloud/user-account"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Service Details", "user-account", "identity.blumer.cloud")
	}
	if rr := get(h, "/domains?selected=domain:does-not-exist.example"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "Cosmos Summary", "Local Cosmos")
	}
	if rr := get(h, "/cosmos"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		body := rr.Body.String()
		hasAll(t, body, "Namespaces", "com", "blumer")
		if strings.Contains(body, ">Domains</span>") {
			t.Fatalf("cosmos hierarchy should not render Domains as the main DNS label")
		}
	}
	if rr := get(h, "/domains/identity.blumer.cloud"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "identity.blumer.cloud")
	}
	if rr := get(h, "/services/identity.blumer.cloud/user-account"); rr.Code != 200 {
		t.Fatal(rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "user-account", "identity.blumer.cloud", "Service Metadata")
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
	if get(h, "/api/v1/domains").Code != 200 || get(h, "/api/v1/blueprints").Code != 200 || get(h, "/api/v1/instances").Code != 200 || get(h, "/api/v1/graph").Code != 200 || get(h, "/api/v1/validate").Code != 200 {
		t.Fatal("api failed")
	}
	if get(h, "/does-not-exist").Code != 404 {
		t.Fatal()
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
	for _, path := range []string{"/health", "/api/v1/cosmos", "/api/v1/domains", "/api/v1/blueprints", "/api/v1/instances/{instance}/compliance"} {
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
	hasAll(t, apiPage.Body.String(), "Open Swagger UI", "GET /openapi.json", "GET /swagger")
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
	jsonEndpoints := []string{"/api/v1/cosmos", "/api/v1/domains", "/api/v1/namespaces", "/api/v1/validate"}
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
	hasAll(t, body, "Namespaces", "com", "cloud", "blumer", "identity", "home", "identity.blumer.com", "identity.blumer.cloud", "treePath", "displayPath")
}

func TestRESTDetailRoutesAndErrors(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/api/v1/domains/identity.blumer.cloud"); rr.Code != 200 {
		t.Fatalf("domain status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "cloud / blumer / identity", "user-account")
	}
	if rr := get(h, "/api/v1/domains/identity.blumer.cloud/services"); rr.Code != 200 {
		t.Fatalf("services status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "user-account")
	}
	if rr := get(h, "/api/v1/domains/identity.blumer.cloud/services/user-account"); rr.Code != 200 {
		t.Fatalf("service status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "identity.blumer.cloud", "user-account")
	}
	if rr := get(h, "/api/v1/domains/does-not-exist.example"); rr.Code != 404 || !strings.Contains(rr.Body.String(), "DOMAIN_NOT_FOUND") {
		t.Fatalf("expected domain 404, got %d %s", rr.Code, rr.Body.String())
	}
	if rr := get(h, "/api/v1/domains/identity.blumer.cloud/services/does-not-exist"); rr.Code != 404 || !strings.Contains(rr.Body.String(), "SERVICE_NOT_FOUND") {
		t.Fatalf("expected service 404, got %d %s", rr.Code, rr.Body.String())
	}
}

func TestBlueprintAndInstanceAPIRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	if rr := get(h, "/api/v1/blueprints"); rr.Code != 200 {
		t.Fatalf("blueprints status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "PB-ACC-MBX-001", "Benutzerkonto mit Mailbox", "required_services", "identity.blumer.cloud/user-account")
	}
	if rr := get(h, "/api/blueprints/PB-ACC-MBX-001"); rr.Code != 200 {
		t.Fatalf("blueprint detail status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "product_blueprint", "required_services", "identity.blumer.cloud/user-account")
	}
	if rr := get(h, "/api/v1/blueprints/SB-1"); rr.Code != 200 {
		t.Fatalf("service blueprint detail status=%d", rr.Code)
	} else {
		hasAll(t, rr.Body.String(), "service_blueprint", "namespace_service_ref", "identity.blumer.cloud/user-account")
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
		"/services":                         {"Services", "Create service", "provisioning-rules"},
		"/namespaces":                       {"Namespace Tree", "identity.blumer.cloud", "user-account"},
		"/blueprints":                       {"Blueprints", "PB-ACC-MBX-001", "SB-1"},
		"/blueprints/PB-ACC-MBX-001":        {"PB-ACC-MBX-001", "Required inputs", "identity.blumer.cloud/user-account"},
		"/instances":                        {"Instances", "PI-ACC-MBX-EXAMPLE-001"},
		"/instances/PI-ACC-MBX-EXAMPLE-001": {"Compliance", "compliant"},
		"/verify":                           {"Verification", "Verify domain", "Evidence files"},
		"/api":                              {"GET /health", "GET /api/v1/cosmos", "GET /api/v1/instances"},
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
	paths := []string{"/domains/missing.example", "/services/identity.blumer.cloud/missing", "/blueprints/missing", "/instances/missing"}
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
func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestCreateDomainAndService(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)
	if rr := postForm(h, "/api/v1/domains", "dns=example.com&owner=Web"); rr.Code != http.StatusCreated {
		t.Fatalf("domain create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/api/v1/domains", "dns=bad"); rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid domain status=%d", rr.Code)
	}
	if rr := postForm(h, "/api/v1/domains", "dns=example.com"); rr.Code != http.StatusConflict {
		t.Fatalf("duplicate domain status=%d", rr.Code)
	}
	if rr := postForm(h, "/api/v1/domains", "dns=example.com&force=on&owner=Web"); rr.Code != http.StatusCreated {
		t.Fatalf("force domain status=%d", rr.Code)
	}
	if rr := postForm(h, "/api/v1/domains", "namespace=net&label=example&owner=Web"); rr.Code != http.StatusCreated || !strings.Contains(rr.Body.String(), "example.net") {
		t.Fatalf("namespace domain create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/api/v1/domains/blumer.com/children", "label=identity-api&owner=Web"); rr.Code != http.StatusCreated || !strings.Contains(rr.Body.String(), "identity-api.blumer.com") {
		t.Fatalf("child domain create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/api/v1/namespaces/cloud/domains", "label=api-blumer&owner=Web"); rr.Code != http.StatusCreated || !strings.Contains(rr.Body.String(), "api-blumer.cloud") {
		t.Fatalf("namespace route create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/api/v1/domains/example.com/services", "name=web-ui&owner=Web"); rr.Code != http.StatusCreated {
		t.Fatalf("service create status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := postForm(h, "/api/v1/domains/example.com/services", "name=web-ui"); rr.Code != http.StatusConflict {
		t.Fatalf("duplicate service status=%d", rr.Code)
	}
	if rr := get(h, "/services?domain=example.com"); rr.Code != 200 || !strings.Contains(rr.Body.String(), "web-ui") {
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

func TestAPIDomainAndServiceDELETE(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)

	// Delete service
	del := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/identity.blumer.cloud/services/user-account", nil)
	h.ServeHTTP(del, req)
	if del.Code != http.StatusOK {
		t.Fatalf("expected 200 on service delete, got %d: %s", del.Code, del.Body.String())
	}
	if !strings.Contains(del.Body.String(), `"deleted":"identity.blumer.cloud/user-account"`) {
		t.Fatalf("unexpected delete response: %s", del.Body.String())
	}
	if get(h, "/api/v1/domains/identity.blumer.cloud/services/user-account").Code != 404 {
		t.Fatal("expected 404 after service delete")
	}

	// Delete domain
	del2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/platform.blumer.cloud", nil)
	h.ServeHTTP(del2, req2)
	if del2.Code != http.StatusOK {
		t.Fatalf("expected 200 on domain delete, got %d: %s", del2.Code, del.Body.String())
	}
	if !strings.Contains(del2.Body.String(), `"deleted":"platform.blumer.cloud"`) {
		t.Fatalf("unexpected delete response: %s", del2.Body.String())
	}
	if get(h, "/api/v1/domains/platform.blumer.cloud").Code != 404 {
		t.Fatal("expected 404 after domain delete")
	}

	// Delete non-existent domain
	del3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/does-not-exist.example", nil)
	h.ServeHTTP(del3, req3)
	if del3.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent domain, got %d", del3.Code)
	}
}
