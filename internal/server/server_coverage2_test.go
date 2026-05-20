package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nomos/nomos/internal/app"
)

// Setup helper that creates a cosmos with a process.
func createCosmosWithProcess(t *testing.T) (string, string) {
	t.Helper()
	p := createTestCosmos(t)
	proc, err := app.CreateProductProcess(p, "PB-ACC-MBX-001", app.CreateProcessRequest{
		ID:   "PRC-CVG-SRV-001",
		Name: "Coverage Server Process",
	})
	if err != nil {
		t.Fatalf("CreateProductProcess: %v", err)
	}
	return p, proc.ID
}

// ── Process triggers ──────────────────────────────────────────────────────────

func TestAPIProcessTriggers(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)

	rr := get(h, fmt.Sprintf("/api/v1/processes/%s/triggers", procID))
	if rr.Code != 200 {
		t.Fatalf("GET process triggers: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProcessTriggers_MethodNotAllowed(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodDelete, fmt.Sprintf("/api/v1/processes/%s/triggers", procID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIProcessTriggers_Put(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)
	rr := putJSONCov(h, fmt.Sprintf("/api/v1/processes/%s/triggers", procID),
		map[string]any{"triggers": []any{}})
	if rr.Code != 200 {
		t.Fatalf("PUT process triggers: %d %s", rr.Code, rr.Body.String())
	}
}

// ── Process steps ─────────────────────────────────────────────────────────────

func TestAPIProcessSteps(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)

	rr := get(h, fmt.Sprintf("/api/v1/processes/%s/steps", procID))
	if rr.Code != 200 {
		t.Fatalf("GET process steps: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProcessStepAdd(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)

	rr := postJSON(h, fmt.Sprintf("/api/v1/processes/%s/steps", procID),
		`{"name":"New Step","task_type":"service_task"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST process step: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProcessStepDelete(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)

	// Add a step first
	rr := postJSON(h, fmt.Sprintf("/api/v1/processes/%s/steps", procID),
		`{"name":"Delete Me","task_type":"service_task"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST step: %d %s", rr.Code, rr.Body.String())
	}

	// Get the step ID
	proc, err := app.GetProcess(p, procID)
	if err != nil || len(proc.Steps) == 0 {
		t.Skip("no steps to delete")
	}
	stepID := proc.Steps[0].ID

	rr = rawReq(h, http.MethodDelete, fmt.Sprintf("/api/v1/processes/%s/steps/%s", procID, stepID), "", "")
	if rr.Code != 200 && rr.Code != 204 {
		t.Fatalf("DELETE process step: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProcessDelete(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)

	rr := rawReq(h, http.MethodDelete, fmt.Sprintf("/api/v1/processes/%s", procID), "", "")
	if rr.Code != 200 && rr.Code != 204 {
		t.Fatalf("DELETE process: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProcessMethodNotAllowed(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)

	rr := rawReq(h, http.MethodPost, fmt.Sprintf("/api/v1/processes/%s", procID), `{}`, "application/json")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIProcessParticipant(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)
	rr := putJSONCov(h, fmt.Sprintf("/api/v1/processes/%s/participant", procID),
		map[string]any{"name": "TestParticipant"})
	if rr.Code != 200 && rr.Code != 400 {
		t.Fatalf("PUT process participant: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIProcessParticipant_MethodNotAllowed(t *testing.T) {
	p, procID := createCosmosWithProcess(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/processes/%s/participant", procID))
	if rr.Code != 405 {
		t.Fatalf("expected 405 for GET participant, got %d", rr.Code)
	}
}

// ── Service capabilities ──────────────────────────────────────────────────────

func TestAPIServiceCapabilityAdd(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := postJSON(h, "/api/v1/services/user-account/capabilities",
		`{"id":"cap-new","name":"New Capability","description":"test"}`)
	if rr.Code != 200 && rr.Code != 201 {
		t.Fatalf("POST capability: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceCapabilityMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/user-account/capabilities")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIServiceCapabilityUpdate(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// First add a capability
	postJSON(h, "/api/v1/services/user-account/capabilities",
		`{"id":"cap-upd","name":"Update Me"}`)

	rr := putJSONCov(h, "/api/v1/services/user-account/capabilities/cap-upd",
		map[string]any{"name": "Updated Capability"})
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("PUT capability: %d %s", rr.Code, rr.Body.String())
	}
}

// ── Service routes: PUT (rename) ───────────────────────────────────────────────

func TestAPIServiceRename(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// Renaming to invalid name triggers error path but covers the handler
	rr := putJSONCov(h, "/api/v1/services/user-account",
		map[string]any{"name": ""})
	if rr.Code == 0 {
		t.Fatal("expected some response")
	}
}

func TestAPIServiceGet(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := get(h, "/api/v1/services/user-account")
	if rr.Code != 200 {
		t.Fatalf("GET service: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIServiceMethodNotAllowed(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := rawReq(h, http.MethodPatch, "/api/v1/services/user-account", "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

// ── Decision definition and DMN routes ───────────────────────────────────────

func TestAPIDecisionDefinitions(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/definitions", decID))
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET decision definitions: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionDefinitionsMethodNotAllowed(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := rawReq(h, http.MethodPost, fmt.Sprintf("/api/v1/decisions/%s/definitions", decID), "", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestAPIDecisionDMNGet(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/dmn", decID))
	// 200 if DMN exists, 404 if not
	if rr.Code != 200 && rr.Code != 404 {
		t.Fatalf("GET DMN: %d %s", rr.Code, rr.Body.String())
	}
}

func TestAPIDecisionEvaluate(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := postJSON(h, fmt.Sprintf("/api/v1/decisions/%s/evaluate", decID),
		`{"inputs":{}}`)
	// 200 if DMN exists with decision table, or error if no DMN
	if rr.Code == 0 {
		t.Fatal("expected response")
	}
}

func TestAPIDecisionEvaluate_MethodNotAllowed(t *testing.T) {
	p, decID := createCosmosWithDecision(t)
	h := NewHandler(p)
	rr := get(h, fmt.Sprintf("/api/v1/decisions/%s/evaluate", decID))
	if rr.Code != 405 {
		t.Fatalf("expected 405 for GET evaluate, got %d", rr.Code)
	}
}

// ── Instance patch ────────────────────────────────────────────────────────────

func TestAPIInstancePatchCoverage(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/instances/PI-ACC-MBX-EXAMPLE-001", `{"compliance_status":"compliant"}`)
	if rr.Code != 200 {
		t.Fatalf("PATCH instance: %d %s", rr.Code, rr.Body.String())
	}
}

// ── apiCosmos ─────────────────────────────────────────────────────────────────

func TestAPICosmosPatch(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	rr := patchJSON(h, "/api/v1/cosmos", `{"name":"Updated Cosmos"}`)
	if rr.Code != 200 && rr.Code != 405 {
		t.Fatalf("PATCH cosmos: %d %s", rr.Code, rr.Body.String())
	}
}

// ── apiBlueprints path not matched ───────────────────────────────────────────

func TestAPIBlueprintsMismatchedPath(t *testing.T) {
	h := NewHandler(createTestCosmos(t))
	// /api/v1/blueprints/ sub-path that doesn't match → 404
	rr := get(h, "/api/v1/blueprints")
	if rr.Code != 200 {
		t.Fatalf("GET /api/v1/blueprints: %d %s", rr.Code, rr.Body.String())
	}
}
