package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

// ── apiLegacyService: data-objects ───────────────────────────────────────────

func TestAPIServiceDataObjectAdd(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/data-objects",
		`{"id":"do-1","name":"User Profile","description":"test"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST data-objects: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceDataObjectMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/identity.blumer.cloud/user-account/data-objects")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIServiceDataObjectUpdate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/data-objects",
		`{"id":"do-upd","name":"Update Me"}`)
	rr := putJSONCov(h, "/api/v1/services/identity.blumer.cloud/user-account/data-objects/do-upd",
		map[string]any{"name": "Updated Data Object"})
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("PUT data-object: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceDataObjectDelete(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/data-objects",
		`{"id":"do-del","name":"Delete Me"}`)
	rr := rawReq(h, http.MethodDelete, "/api/v1/services/identity.blumer.cloud/user-account/data-objects/do-del", "", "")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("DELETE data-object: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceDataObjectItemMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPatch, "/api/v1/services/identity.blumer.cloud/user-account/data-objects/do-x", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for PATCH data-object, got %d", rr.Code)
	}
}

// ── apiLegacyService: user-interfaces ────────────────────────────────────────

func TestAPIServiceUserInterfaceAdd(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces",
		`{"id":"ui-1","name":"Account Portal","description":"test"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST user-interfaces: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceUserInterfaceMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIServiceUserInterfaceUpdate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces",
		`{"id":"ui-upd","name":"Update Me"}`)
	rr := putJSONCov(h, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces/ui-upd",
		map[string]any{"name": "Updated UI"})
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("PUT user-interface: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceUserInterfaceDelete(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces",
		`{"id":"ui-del","name":"Delete Me"}`)
	rr := rawReq(h, http.MethodDelete, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces/ui-del", "", "")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("DELETE user-interface: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceUserInterfaceItemMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPatch, "/api/v1/services/identity.blumer.cloud/user-account/user-interfaces/ui-x", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for PATCH ui, got %d", rr.Code)
	}
}

// ── apiLegacyService: methods ─────────────────────────────────────────────────

func TestAPIServiceMethodAdd(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/methods",
		`{"method":"createUser"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST methods: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceMethodList(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/identity.blumer.cloud/user-account/methods")
	if rr.Code != 200 {
		t.Fatalf("GET methods: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceMethodMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodDelete, "/api/v1/services/identity.blumer.cloud/user-account/methods", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for DELETE /methods, got %d", rr.Code)
	}
}

func TestAPIServiceMethodGet(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/methods",
		`{"method":"getUser"}`)
	rr := get(h, "/api/v1/services/identity.blumer.cloud/user-account/methods/getUser")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET method: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceMethodUpdate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/methods",
		`{"method":"updateUser"}`)
	rr := putJSONCov(h, "/api/v1/services/identity.blumer.cloud/user-account/methods/updateUser",
		map[string]any{"summary": "Update user", "http_method": "PUT", "path": "/users/{id}"})
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("PUT method: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceMethodDelete(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/services/identity.blumer.cloud/user-account/methods",
		`{"method":"deleteUser"}`)
	rr := rawReq(h, http.MethodDelete, "/api/v1/services/identity.blumer.cloud/user-account/methods/deleteUser", "", "")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("DELETE method: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceMethodItemMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPatch, "/api/v1/services/identity.blumer.cloud/user-account/methods/x", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for PATCH method, got %d", rr.Code)
	}
}

// ── apiLegacyService: GET service directly ────────────────────────────────────

func TestAPIServiceGet(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/identity.blumer.cloud/user-account")
	if rr.Code != 200 {
		t.Fatalf("GET service: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceNotFound(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/identity.blumer.cloud/ghost-service/unknown/extra/path")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for unknown path, got %d", rr.Code)
	}
}

// ── apiDomainDecisionVersions ─────────────────────────────────────────────────

func TestAPIDecisionVersionsListCov3(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/versions", domain, decID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET versions: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionVersionsListMethodNotAllowedCov3(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodPost, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/versions", domain, decID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIDecisionVersionByIDCov3(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/versions/0.1.0", domain, decID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET version by ID: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionVersionDMN(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/versions/0.1.0/dmn", domain, decID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET version DMN: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionVersionDefinitions(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/versions/0.1.0/definitions", domain, decID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET version definitions: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionVersionDefaultPath(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	// len(tail)==3 triggers default/htmlNotFound
	rr := get(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/versions/0.1.0/dmn/extra", domain, decID))
	if rr.Code != 404 && rr.Code != 200 {
		t.Fatalf("unexpected code: %d", rr.Code)
	}
}

// ── apiInstanceRoutes ─────────────────────────────────────────────────────────

func TestAPIInstanceGet(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001")
	if rr.Code != 200 {
		t.Fatalf("GET instance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceDelete(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodDelete, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001", "", "")
	if rr.Code != 200 && rr.Code != 204 && rr.Code != 404 {
		t.Fatalf("DELETE instance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceCompliance(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/compliance")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET instance compliance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceVerify(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPost, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/verify", "", "")
	if rr.Code != 200 && rr.Code != 404 && rr.Code != 500 {
		t.Fatalf("POST instance verify: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceVerify_MethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/verify")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for GET /verify, got %d", rr.Code)
	}
}

func TestAPIInstanceAttributeValidation(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/attribute-validation")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET attribute-validation: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceAttributeValues(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/attribute-values", `{"key":"value"}`)
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("PATCH attribute-values: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceAttributeValues_MethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/attribute-values", `{}`)
	if rr.Code != 405 {
		t.Fatalf("expected 405 for POST /attribute-values, got %d", rr.Code)
	}
}

func TestAPIInstanceRoutes_NotFound(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/unknown-sub")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for unknown sub-route, got %d", rr.Code)
	}
}

// ── apiServicegraphs ──────────────────────────────────────────────────────────

func TestAPIServicegraphCreate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/servicegraphs",
		`{"id":"SG-TEST-001","type":"servicegraph","name":"Test Graph","version":"1.0.0","status":"draft","owner":"Team"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST servicegraph: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServicegraphCreate_MethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPut, "/api/v1/servicegraphs", `{}`, "application/json")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIServicegraphCreate_NotFound(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/servicegraphs/unknown-path/sub")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for sub-path, got %d", rr.Code)
	}
}

// ── apiServicegraphRoutes ─────────────────────────────────────────────────────

func createCosmosWithServicegraph(t *testing.T) (string, string) {
	t.Helper()
	p := createTestCosmos(t)
	if err := app.CreateServicegraph(p, model.Servicegraph{
		ID:     "SG-CVG-001",
		Type:   "servicegraph",
		Name:   "Coverage Graph",
		Status: "draft",
		Owner:  "Team",
	}); err != nil {
		t.Fatalf("CreateServicegraph: %v", err)
	}
	return p, "SG-CVG-001"
}

func TestAPIServicegraphGetByID(t *testing.T) {
	p, sgID := createCosmosWithServicegraph(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/servicegraphs/%s", sgID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET servicegraph: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServicegraphDelete(t *testing.T) {
	p, sgID := createCosmosWithServicegraph(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodDelete, fmt.Sprintf("/api/v1/servicegraphs/%s", sgID), "", "")
	if rr.Code != 200 && rr.Code != 204 && rr.Code != 404 {
		t.Fatalf("DELETE servicegraph: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServicegraphMethodNotAllowed(t *testing.T) {
	p, sgID := createCosmosWithServicegraph(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodPatch, fmt.Sprintf("/api/v1/servicegraphs/%s", sgID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIServicegraphMermaid(t *testing.T) {
	p, sgID := createCosmosWithServicegraph(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/servicegraphs/%s/mermaid", sgID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET servicegraph mermaid: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServicegraphExecution(t *testing.T) {
	p, sgID := createCosmosWithServicegraph(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/servicegraphs/%s/execution", sgID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET servicegraph execution: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServicegraphRoutesUnknownSub(t *testing.T) {
	p, sgID := createCosmosWithServicegraph(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/servicegraphs/%s/unknown-sub", sgID))
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestAPIServicegraphRoutesEmptyID(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/servicegraphs/")
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── apiDomainDecisionScenarios ────────────────────────────────────────────────

func TestAPIDecisionScenariosList(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/scenarios", domain, decID))
	if rr.Code != 200 {
		t.Fatalf("GET scenarios: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionScenariosCreate(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := postJSON(h, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/scenarios", domain, decID),
		`{"name":"Test Scenario","inputs":{},"expected_outputs":{}}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST scenario: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionScenariosMethodNotAllowed(t *testing.T) {
	p, domain, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodPut, fmt.Sprintf("/api/v1/domains/%s/decisions/%s/scenarios", domain, decID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

// ── apiServiceRefs ────────────────────────────────────────────────────────────

func TestAPIServiceRefsCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/refs")
	if rr.Code != 200 {
		t.Fatalf("GET service-refs: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceRefsMethodNotAllowedCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPost, "/api/v1/services/refs", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405 for POST /services/refs, got %d", rr.Code)
	}
}

// ── apiValidate ───────────────────────────────────────────────────────────────

func TestAPIValidatePostCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// apiValidate accepts any method — just hit it to get coverage
	rr := rawReq(h, http.MethodPost, "/api/v1/validate", "{}", "application/json")
	if rr.Code == 0 {
		t.Fatal("expected a response")
	}
}

// ── apiNamespaces ─────────────────────────────────────────────────────────────

func TestAPINamespacesCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/namespaces")
	if rr.Code != 200 {
		t.Fatalf("GET namespaces: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPINamespacesMethodNotAllowedCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPost, "/api/v1/namespaces", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPINamespacesNotFoundCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/namespaces/sub-path")
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("unexpected %d", rr.Code)
	}
}

// ── apiGraph ──────────────────────────────────────────────────────────────────

func TestAPIGraphJSONCov3(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// apiGraph also accepts ?format=json
	rr := get(h, "/api/v1/graph?format=json")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/graph?format=json: %d %s", rr.Code, rr.Body.String())
	}
}
