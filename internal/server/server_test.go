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
	if rr := postJSON(h, "/api/v1/domains/identity-api.blumer.com/children", `{"label":"products","owner":"Web"}`); rr.Code != http.StatusCreated || !strings.Contains(rr.Body.String(), "products.identity-api.blumer.com") || !strings.Contains(rr.Body.String(), `"treePath":"/com/blumer/identity-api/products"`) || !strings.Contains(rr.Body.String(), `.nomos/domains/com/blumer/identity-api/products/domain.yaml`) {
		t.Fatalf("json child domain create status=%d body=%s", rr.Code, rr.Body.String())
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

func TestLegacyServiceAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// valid legacy service lookup
	rr := get(h, "/api/v1/services/identity.blumer.cloud/user-account")
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	hasAll(t, rr.Body.String(), "user-account")

	// not enough path segments → 404
	rr = get(h, "/api/v1/services/identity.blumer.cloud")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for bad path, got %d", rr.Code)
	}

	// non-existent service → 404
	rr = get(h, "/api/v1/services/identity.blumer.cloud/no-such-service")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for missing service, got %d", rr.Code)
	}
}

func TestVerifyDomainAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// empty domain → 404
	rr := postJSON(h, "/api/v1/verify/domain/", "")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for empty domain, got %d", rr.Code)
	}

	// GET method → 405
	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/verify/domain/example.com", nil)
	h.ServeHTTP(rr2, req)
	if rr2.Code != 405 {
		t.Fatalf("expected 405 for GET, got %d", rr2.Code)
	}

	// POST with real domain → triggers VerifyDomain (will likely error due to DNS in test env)
	rr = postJSON(h, "/api/v1/verify/domain/identity.blumer.cloud", "")
	// We expect either 200 (if DNS works) or a 4xx/5xx error — just ensure it was reached
	if rr.Code == 404 {
		t.Fatalf("handler should have been reached (not 404), got %d body=%s", rr.Code, rr.Body.String())
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

	// POST /domains creates a domain and redirects
	rr := httptest.NewRecorder()
	body := strings.NewReader("dns=newtest.example.com&owner=Test+Team")
	req := httptest.NewRequest(http.MethodPost, "/domains", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)
	if rr.Code != 303 {
		t.Fatalf("POST /domains expected 303 redirect, got %d body=%s", rr.Code, rr.Body.String())
	}

	// POST /domains with invalid name → error page (non-303)
	rr2 := httptest.NewRecorder()
	body2 := strings.NewReader("dns=bad/name&owner=Team")
	req2 := httptest.NewRequest(http.MethodPost, "/domains", body2)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr2, req2)
	if rr2.Code == 303 {
		t.Fatalf("POST /domains with bad name should not redirect")
	}

	// POST /services → redirects
	rr3 := httptest.NewRecorder()
	body3 := strings.NewReader("domain=identity.blumer.cloud&name=new-service&owner=Team")
	req3 := httptest.NewRequest(http.MethodPost, "/services", body3)
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr3, req3)
	if rr3.Code != 303 {
		t.Fatalf("POST /services expected 303 redirect, got %d body=%s", rr3.Code, rr3.Body.String())
	}

	// POST /verify → redirects (DNS will fail but handler still redirects)
	rr4 := httptest.NewRecorder()
	body4 := strings.NewReader("domain=identity.blumer.cloud")
	req4 := httptest.NewRequest(http.MethodPost, "/verify", body4)
	req4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr4, req4)
	if rr4.Code != 303 {
		t.Fatalf("POST /verify expected 303 redirect, got %d body=%s", rr4.Code, rr4.Body.String())
	}

	// POST unknown route → error page
	rr5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPost, "/unknown-route", nil)
	h.ServeHTTP(rr5, req5)
	// formPost only handles /domains, /services, /verify; anything else → 404 error page
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

func TestNamespaceRoutesAPI(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// POST /api/v1/namespaces/{ns}/domains
	rr := httptest.NewRecorder()
	body := strings.NewReader("label=newlabel&owner=Team")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/cloud/domains", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST namespace domain status=%d body=%s", rr.Code, rr.Body.String())
	}

	// method not allowed
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/cloud/domains", nil)
	h.ServeHTTP(rr2, req2)
	if rr2.Code != 405 {
		t.Fatalf("expected 405, got %d", rr2.Code)
	}

	// bad path
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/cloud/unknown", nil)
	h.ServeHTTP(rr3, req3)
	if rr3.Code != 404 {
		t.Fatalf("expected 404 for unknown namespace sub-path, got %d", rr3.Code)
	}
}
