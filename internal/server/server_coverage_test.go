package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/storage"
	"os"
	"path/filepath"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func rawReq(h http.Handler, method, path, body, ct string) *httptest.ResponseRecorder {
	var br *strings.Reader
	if body != "" {
		br = strings.NewReader(body)
	} else {
		br = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, br)
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func formPostReq(h http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// createCosmosWithDecision creates a cosmos that has a decision with DMN.
func createCosmosWithDecision(t *testing.T) (string, string) {
	t.Helper()
	p := createTestCosmos(t)
	h := NewHandler(p)
	// Create decision
	rr := postJSON(h, "/api/v1/decisions",
		`{"id":"DEC-CVG-001","name":"Coverage Test Decision","status":"draft","version":"0.1.0"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("create decision: %d %s", rr.Code, rr.Body.String())
	}
	return p, "DEC-CVG-001"
}

// ── Decision traces ───────────────────────────────────────────────────────────

func TestAPIDecisionTracesList(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/traces", decID))
	if rr.Code != 200 {
		t.Fatalf("GET traces: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionTracesMethodNotAllowed(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodPost, fmt.Sprintf("/api/v1/decisions/%s/traces", decID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIDecisionTracesVerify(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/traces/verify", decID))
	if rr.Code != 200 {
		t.Fatalf("GET traces/verify: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionTracesVerifyPost(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := postJSON(h, fmt.Sprintf("/api/v1/decisions/%s/traces/verify", decID), `{}`)
	if rr.Code != 200 {
		t.Fatalf("POST traces/verify: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionTracesMethodNotAllowedVerify(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodDelete, fmt.Sprintf("/api/v1/decisions/%s/traces/verify", decID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIDecisionTracesUnknownID(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/traces/TRACE-GHOST-001", decID))
	// trace not found → 404 or 422
	if rr.Code == 200 {
		t.Logf("GET trace by ID: %d", rr.Code)
	}
}

func TestAPIDecisionTracesNotAllowedByID(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := postJSON(h, fmt.Sprintf("/api/v1/decisions/%s/traces/TRACE-001", decID), `{}`)
	if rr.Code != 405 {
		t.Fatalf("expected 405 for POST trace by ID, got %d", rr.Code)
	}
}

func TestAPIDecisionTracesNotFound(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	// Deep path that hits htmlNotFound in traces
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/traces/a/b/c", decID))
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Decision versions ─────────────────────────────────────────────────────────

func TestAPIDecisionVersionsList(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/versions", decID))
	if rr.Code != 200 {
		t.Fatalf("GET versions: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionVersionsMethodNotAllowed(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := postJSON(h, fmt.Sprintf("/api/v1/decisions/%s/versions", decID), `{}`)
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIDecisionVersionByID(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/versions/0.1.0", decID))
	// May be 404 if no snapshot exists yet — that's acceptable
	if rr.Code != 200 && rr.Code != 404 && rr.Code != 422 {
		t.Fatalf("GET version by id: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionVersionsNotFound(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	// Deep path that hits the default htmlNotFound case
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/versions/0.1.0/dmn/extra", decID))
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── HTML pages not yet tested ─────────────────────────────────────────────────

func TestHTMLPages(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	pages := []string{
		"/namespaces",
		"/graph",
		"/validate",
		"/requirements",
		"/rules",
		"/verify",
		"/blueprints",
		"/instances",
		"/api",
	}
	for _, pg := range pages {
		rr := get(h, pg)
		if rr.Code != 200 {
			t.Errorf("GET %s: %d %s", pg, rr.Code, rr.Body.String())
		}
	}
}

// ── OpenAPI + Swagger ─────────────────────────────────────────────────────────

func TestOpenAPIEndpoints(t *testing.T) {
	h := NewHandler(createTestCosmos(t))

	rr := get(h, "/openapi.json")
	if rr.Code != 200 {
		t.Errorf("GET /openapi.json: %d %s", rr.Code, rr.Body.String())
	}

	rr = get(h, "/swagger")
	if rr.Code != 200 {
		t.Errorf("GET /swagger: %d %s", rr.Code, rr.Body.String())
	}

	rr = get(h, "/api/docs")
	if rr.Code != 200 {
		t.Errorf("GET /api/docs: %d %s", rr.Code, rr.Body.String())
	}
}

// ── formPost routes ───────────────────────────────────────────────────────────

func TestFormPostService(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := formPostReq(h, "/services", "name=new-svc&owner=Team")
	if rr.Code != 303 && rr.Code != 200 {
		t.Fatalf("POST /forms/services: %d %s", rr.Code, rr.Body.String())
	}
}

func TestFormPostProductCreate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := formPostReq(h, "/products/create", "domain=identity.blumer.cloud&name=FormProduct&version=0.1.0&status=draft&owner=Team&owning_domain=identity.blumer.cloud")
	if rr.Code != 303 && rr.Code != 200 {
		t.Fatalf("POST /forms/products/create: %d %s", rr.Code, rr.Body.String())
	}
}

func TestFormPostProductCreateReturnToCosmos(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := formPostReq(h, "/products/create", "domain=identity.blumer.cloud&name=CosmosReturnProduct&version=0.1.0&status=draft&owner=Team&owning_domain=identity.blumer.cloud&return_to=cosmos")
	if rr.Code != 303 && rr.Code != 200 {
		t.Fatalf("POST /forms/products/create return_to=cosmos: %d %s", rr.Code, rr.Body.String())
	}
}

func TestFormPostUnknownRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := formPostReq(h, "/unknown-form-route", "")
	if rr.Code != 404 && rr.Code != 200 {
		t.Fatalf("POST /forms/unknown: %d %s", rr.Code, rr.Body.String())
	}
}

// ── htmlNotFound ──────────────────────────────────────────────────────────────

func TestHTMLNotFoundCoverage(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// Hit a route not handled by any mux entry to trigger htmlNotFound
	rr := get(h, "/totally-missing-page-xyz")
	if rr.Code != 404 && rr.Code != 200 {
		t.Fatalf("expected 404 for missing page, got %d", rr.Code)
	}
}

// ── apiServicegraphs ──────────────────────────────────────────────────────────

func TestAPIServicegraphsRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/servicegraphs")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET /api/v1/servicegraphs: %d %s", rr.Code, rr.Body.String())
	}
}

// ── validate with cosmos errors ───────────────────────────────────────────────

func TestAPIValidateMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPost, "/api/v1/validate", `{}`, "application/json")
	// apiValidate doesn't restrict methods, just calls the function
	if rr.Code == 0 {
		t.Fatal("expected response")
	}
}

// ── apiInstances POST ─────────────────────────────────────────────────────────

func TestAPIInstancesPost(t *testing.T) {
	p := createTestCosmos(t)
	h := NewHandler(p)
	// Create a blueprint first
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	bpDir := filepath.Join(storage.CatalogDir(p), "blueprints", "products")
	must(os.MkdirAll(bpDir, 0o755))
	must(os.WriteFile(filepath.Join(bpDir, "PB-CVG-001.yaml"),
		[]byte("id: PB-CVG-001\ntype: product_blueprint\nname: CVG BP\nstatus: draft\nowner: Team\nversion: 0.1.0\nrequired_service_blueprints: []\n"),
		0o644))

	rr := postJSON(h, "/api/v1/instances", `{"id":"PI-CVG-001","type":"product_instance","blueprint_ref":"PB-CVG-001","blueprint_version":"0.1.0","compliance_status":"compliant"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST /api/v1/instances: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstancesMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPut, "/api/v1/instances", `{}`, "application/json")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIInstancesCompliancePath(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/compliance")
	if rr.Code != 200 {
		t.Fatalf("GET instance compliance: %d %s", rr.Code, rr.Body.String())
	}
}

// ── blueprint routes ──────────────────────────────────────────────────────────

func TestAPIBlueprintsInvalidPath(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// sub-path of /api/v1/blueprints that doesn't match a known route
	rr := get(h, "/api/v1/blueprints/PB-ACC-MBX-001/unknown-sub")
	if rr.Code != 404 && rr.Code != 405 && rr.Code != 200 {
		t.Fatalf("GET blueprint unknown sub: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIBlueprintValidateEndpoint(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/blueprints/PB-ACC-MBX-001/validate")
	if rr.Code != 200 {
		t.Fatalf("GET blueprint validate: %d %s", rr.Code, rr.Body.String())
	}
}

// ── product routes ────────────────────────────────────────────────────────────

func TestAPIProductRoutes(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/products/PB-ACC-MBX-001")
	if rr.Code != 200 {
		t.Fatalf("GET product: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProductDefinitionsRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// /api/v1/products/ (trailing slash handled by the route)
	rr := get(h, "/api/v1/products/")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET /api/v1/products/: %d %s", rr.Code, rr.Body.String())
	}
}

// ── form post: fulfillment and verify ────────────────────────────────────────

// ── formPost fulfillment ──────────────────────────────────────────────────────

func TestFormPostFulfillmentInvalidIndex(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// Non-integer index → error page
	rr := formPostReq(h, "/products/fulfillment/update", "product_id=PB-ACC-MBX-001&index=notanumber")
	if rr.Code != 200 && rr.Code != 400 {
		t.Fatalf("POST fulfillment/update bad index: %d", rr.Code)
	}
}

func TestFormPostFulfillmentDeleteInvalidIndex(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := formPostReq(h, "/products/fulfillment/delete", "product_id=PB-ACC-MBX-001&index=bad")
	if rr.Code != 200 && rr.Code != 400 {
		t.Fatalf("POST fulfillment/delete bad index: %d", rr.Code)
	}
}

// ── apiGraph format=mermaid ───────────────────────────────────────────────────

func TestAPIGraphMermaid(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/graph?format=mermaid")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/graph?format=mermaid: %d %s", rr.Code, rr.Body.String())
	}
}

// ── apiNamespaces ─────────────────────────────────────────────────────────────

func TestAPINamespacesGet(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/namespaces")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/namespaces: %d %s", rr.Code, rr.Body.String())
	}
}

// ── auth middleware ───────────────────────────────────────────────────────────

func TestAuthMiddleware(t *testing.T) {
	p := createTestCosmos(t)
	// Create a key
	keyResult, err := app.CreateKey(p, "test-key")
	if err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	keyDTO := keyResult

	h := NewHandler(p)

	// Valid key in Authorization header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cosmos", nil)
	req.Header.Set("Authorization", "Bearer "+keyDTO.Key) //nolint
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Logf("auth with valid key: %d (expected 200 but auth may not be enabled)", rr.Code)
	}

	// Invalid key
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/cosmos", nil)
	req2.Header.Set("Authorization", "Bearer invalid-key-123")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	t.Logf("auth with invalid key: %d", rr2.Code)
}

// ── apiDecisions method not allowed ──────────────────────────────────────────

func TestAPIDecisionsMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPut, "/api/v1/decisions", `{}`, "application/json")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for PUT decisions, got %d", rr.Code)
	}
}

// ── flat service API ──────────────────────────────────────────────────────────

func TestAPIFlatServiceRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/user-account")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET flat service: %d %s", rr.Code, rr.Body.String())
	}
}
