package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestCosmos(t *testing.T) string {
	p := t.TempDir()
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Join(p, "domains/identity.blumer.cloud/services/user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "domains/platform.blumer.cloud/services/rule-validation-api"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "catalog/blueprints/products"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "catalog/blueprints/services"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "catalog/instances/products"), 0o755))
	must(os.WriteFile(filepath.Join(p, "cosmos.yaml"), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: unknown\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/domain.yaml"), []byte("name: identity.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/platform.blumer.cloud/domain.yaml"), []byte("name: platform.blumer.cloud\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/services/user-account/service.yaml"), []byte("name: user-account\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/platform.blumer.cloud/services/rule-validation-api/service.yaml"), []byte("name: rule-validation-api\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "catalog/blueprints/products/account.yaml"), []byte("id: PB-ACC-MBX-001\ntype: product_blueprint\nname: Benutzerkonto mit Mailbox\nversion: 0.1.0\nstatus: draft\nowner: Team\nrequired_inputs:\n  - person_reference\nrequired_service_blueprints:\n  - SB-1\nrequired_services:\n  - service_ref: identity.blumer.cloud/user-account\n    service_blueprint_ref: SB-1\n    required: true\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "catalog/blueprints/services/account-service.yaml"), []byte("id: SB-1\ntype: service_blueprint\nname: Account Service\nversion: 0.1.0\nstatus: draft\nowner: Team\nnamespace_service_ref: identity.blumer.cloud/user-account\ncapabilities:\n  - create_account\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "catalog/instances/products/account-instance.yaml"), []byte("id: PI-ACC-MBX-EXAMPLE-001\ntype: product_instance\nname: Beispielinstanz Benutzerkonto mit Mailbox\nblueprint_ref: PB-ACC-MBX-001\nblueprint_version: 0.1.0\ncompliance_status: compliant\nfindings: []\n"), 0o644))
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
		hasAll(t, rr.Body.String(), "Domains Explorer", "Browse domains and services like a repository tree", "identity.blumer.cloud", "user-account", "explorer-layout", "explorer-tree", "details-panel", "Domain count", "Service count")
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
	hasAll(t, body, "cloud", "blumer", "identity", "platform", "user-account", "identity.blumer.cloud", "treePath", "displayPath")
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
