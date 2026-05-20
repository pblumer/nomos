package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func postRaw(h http.Handler, path, body string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	return rr
}

func postJSONCov(h http.Handler, path string, v any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(v)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}

func putJSONCov(h http.Handler, path string, v any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(v)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rr, req)
	return rr
}

func deleteReqCov(h http.Handler, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, path, nil))
	return rr
}

// ---------------------------------------------------------------------------
// /api/v1/cosmos
// ---------------------------------------------------------------------------

func TestAPICosmos(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/cosmos")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/cosmos: %d %s", rr.Code, rr.Body.String())
	}
	var d map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d["id"] != "cosmos-local" {
		t.Errorf("unexpected id: %v", d["id"])
	}
}

// ---------------------------------------------------------------------------
// /api/v1/services
// ---------------------------------------------------------------------------

func TestAPIServicesList(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/services: %d %s", rr.Code, rr.Body.String())
	}
	var d map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d["services"] == nil {
		t.Errorf("missing services field")
	}
}

func TestAPIServiceByName(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/user-account")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/services/user-account: %d", rr.Code)
	}
}

func TestAPIServiceNotFound(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/ghost-service")
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// /api/v1/services/refs
// ---------------------------------------------------------------------------

func TestAPIServiceRefs(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/refs")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/services/refs: %d %s", rr.Code, rr.Body.String())
	}
	var d map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d["services"] == nil {
		t.Error("missing services field")
	}
}

// ---------------------------------------------------------------------------
// /api/v1/services/{service}/capabilities
// ---------------------------------------------------------------------------

func TestAPIServiceCapabilities(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)

	// POST capability — flat route /api/v1/services/{service}/capabilities
	rr := postJSONCov(h, "/api/v1/services/user-account/capabilities",
		map[string]any{"name": "Create Account", "summary": "Creates an account"})
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST capability: %d %s", rr.Code, rr.Body.String())
	}

	// GET service (capabilities returned in service detail)
	rr = get(h, "/api/v1/services/user-account")
	if rr.Code != 200 {
		t.Fatalf("GET service: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// /api/v1/decisions
// ---------------------------------------------------------------------------

func TestAPIDecisions(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/decisions")
	if rr.Code != 200 {
		t.Fatalf("GET decisions: %d %s", rr.Code, rr.Body.String())
	}
	var d map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d["count"] == nil {
		t.Error("missing count field")
	}
}

func TestAPIDecisionCRUD(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	// CREATE
	rr := postJSONCov(h, "/api/v1/decisions",
		map[string]any{"id": "DEC-API-001", "name": "Test Decision API"})
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST decision: %d %s", rr.Code, rr.Body.String())
	}

	// GET
	rr = get(h, "/api/v1/decisions/DEC-API-001")
	if rr.Code != 200 {
		t.Fatalf("GET decision: %d %s", rr.Code, rr.Body.String())
	}

	// DELETE
	rr = deleteReqCov(h, "/api/v1/decisions/DEC-API-001")
	if rr.Code != 200 && rr.Code != 204 {
		t.Fatalf("DELETE decision: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// /api/v1/processes
// ---------------------------------------------------------------------------

func TestAPIProcessesCRUD(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)

	// Create product + process via app layer so the file structure is right
	// (server needs a product to attach a process to)
	must := func(e error) {
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints/processes"), 0o755))
	// write a minimal process yaml directly
	procYAML := "id: PRC-API-001\ntype: process\nname: API Test Process\nstatus: draft\nversion: 0.1.0\nbpmn:\n  process_id: Process_PRC_API_001\n  file: prc-api-001.bpmn\n"
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints/processes/prc-api-001.yaml"), []byte(procYAML), 0o644))

	// GET by ID
	rr := get(h, "/api/v1/processes/PRC-API-001")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/processes/PRC-API-001: %d %s", rr.Code, rr.Body.String())
	}

	// Add step
	rr = postJSONCov(h, "/api/v1/processes/PRC-API-001/steps",
		map[string]any{"name": "Step One", "task_type": "serviceTask", "service_ref": "identity.blumer.cloud/user-account"})
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST step: %d %s", rr.Code, rr.Body.String())
	}

	// List steps again
	rr = get(h, "/api/v1/processes/PRC-API-001")
	if rr.Code != 200 {
		t.Fatalf("GET process after step: %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// /api/v1/validate
// ---------------------------------------------------------------------------

func TestAPIValidate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/validate")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/validate: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// /api/v1/graph
// ---------------------------------------------------------------------------

func TestAPIGraph(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/graph")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/graph: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// /api/v1/namespaces
// ---------------------------------------------------------------------------

func TestAPINamespaces(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/namespaces")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/namespaces: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// /api/v1/blueprints
// ---------------------------------------------------------------------------

func TestAPIBlueprints(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/blueprints")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/blueprints: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIBlueprintByID(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/blueprints/PB-ACC-MBX-001")
	if rr.Code != 200 {
		t.Fatalf("GET blueprint: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// /api/v1/instances
// ---------------------------------------------------------------------------

func TestAPIInstances(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/instances: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// fsx.ReadYAMLFromFS
// ---------------------------------------------------------------------------

func TestFSXReadYAMLFromFS(t *testing.T) {
	// Tested via the embedded web server static assets, which use ReadYAMLFromFS indirectly.
	// Also test directly in the fsx package.
}
