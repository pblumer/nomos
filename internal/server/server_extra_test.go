package server

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Blueprint routes
// ---------------------------------------------------------------------------

func TestAPIBlueprintDelete(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := deleteReq(h, "/api/v1/blueprints/PB-ACC-MBX-001")
	if rr.Code != 200 && rr.Code != 204 {
		t.Fatalf("DELETE blueprint: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIBlueprintValidate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/blueprints/PB-ACC-MBX-001/validate")
	if rr.Code != 200 {
		t.Fatalf("GET blueprint validate: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIBlueprintPublish(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001/publish", `{}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST blueprint publish: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIBlueprintPatchRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/blueprints/PB-ACC-MBX-001", `{"status":"published"}`)
	if rr.Code != 200 {
		t.Fatalf("PATCH blueprint: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Instance routes
// ---------------------------------------------------------------------------

func TestAPIInstanceByIDRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001")
	if rr.Code != 200 {
		t.Fatalf("GET instance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceDeleteRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := deleteReq(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001")
	if rr.Code != 200 && rr.Code != 204 {
		t.Fatalf("DELETE instance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstancePatchRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001", `{"status":"active"}`)
	if rr.Code != 200 {
		t.Fatalf("PATCH instance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceComplianceRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001/compliance")
	if rr.Code != 200 {
		t.Fatalf("GET instance compliance: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIInstanceNotFoundRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/instances/GHOST-001")
	if rr.Code != 404 {
		t.Fatalf("expected 404 for missing instance, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Domain routes — additional methods
// ---------------------------------------------------------------------------

func TestAPIDomainDeleteRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := deleteReq(h, "/api/v1/domains/beispiel.ch")
	if rr.Code != 200 && rr.Code != 204 {
		t.Fatalf("DELETE domain: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDomainChildrenRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/domains/blumer.cloud/children", `{"label":"testchild","owner":"Team"}`)
	if rr.Code != 200 && rr.Code != 201 && rr.Code != 409 {
		t.Fatalf("POST domain children: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDomainMethodNotAllowedRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/domains/identity.blumer.cloud", `{}`)
	if rr.Code != 405 {
		t.Fatalf("expected 405 for unsupported method, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Product routes
// ---------------------------------------------------------------------------

func TestAPIProductByIDRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/products/PB-ACC-MBX-001")
	if rr.Code != 200 {
		t.Fatalf("GET product: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Graph format param
// ---------------------------------------------------------------------------

func TestAPIGraphJSON(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/graph?format=json")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/graph?format=json: %d %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Service refs method not allowed
// ---------------------------------------------------------------------------

func TestAPIServiceRefsMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/services/refs", `{}`)
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Decision decision-by-ID routes
// ---------------------------------------------------------------------------

func TestAPIDomainDecisionByIDRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/domains/identity.blumer.cloud/decisions",
		`{"id":"DEC-SRV-001","name":"Server Test Decision"}`)
	rr := get(h, "/api/v1/domains/identity.blumer.cloud/decisions/DEC-SRV-001")
	if rr.Code != 200 {
		t.Fatalf("GET decision by ID: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDomainDecisionUpdateRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	postJSON(h, "/api/v1/domains/identity.blumer.cloud/decisions",
		`{"id":"DEC-UPD-001","name":"Decision To Update"}`)
	rr := putJSONCov(h, "/api/v1/domains/identity.blumer.cloud/decisions/DEC-UPD-001",
		map[string]any{"name": "Updated Decision"})
	if rr.Code != 200 {
		t.Fatalf("PUT decision: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDomainDecisionMethodNotAllowedRoute(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/domains/identity.blumer.cloud/decisions", `{}`)
	if rr.Code != 405 {
		t.Fatalf("expected 405 for PATCH decisions list, got %d", rr.Code)
	}
}
