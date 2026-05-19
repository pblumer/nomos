package app

import (
	"testing"

	"github.com/nomos/nomos/internal/model"
)

func createDecisionCosmos(t *testing.T) string {
	t.Helper()
	return createAppTestCosmos(t)
}

func TestListDecisions_Empty(t *testing.T) {
	p := createDecisionCosmos(t)
	dto, err := ListDecisions(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatalf("ListDecisions: %v", err)
	}
	if dto.Count != 0 {
		t.Errorf("expected 0 decisions, got %d", dto.Count)
	}
}

func TestListDecisions_WithDecision(t *testing.T) {
	p := createDecisionCosmos(t)
	if _, err := CreateDecision(p, "identity.blumer.cloud", CreateDecisionRequest{
		ID: "DEC-LIST-001", Name: "Test Decision",
	}); err != nil {
		t.Fatal(err)
	}
	dto, err := ListDecisions(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatalf("ListDecisions: %v", err)
	}
	if dto.Count != 1 {
		t.Errorf("expected 1 decision, got %d", dto.Count)
	}
}

func TestListDecisions_DomainNotFound(t *testing.T) {
	p := createDecisionCosmos(t)
	_, err := ListDecisions(p, "ghost.domain")
	if err == nil {
		t.Error("expected error for unknown domain")
	}
}

func TestGetDecision_Success(t *testing.T) {
	p := createDecisionCosmos(t)
	if _, err := CreateDecision(p, "identity.blumer.cloud", CreateDecisionRequest{
		ID: "DEC-GET-001", Name: "Get Me",
	}); err != nil {
		t.Fatal(err)
	}
	dto, err := GetDecision(p, "identity.blumer.cloud", "DEC-GET-001")
	if err != nil {
		t.Fatalf("GetDecision: %v", err)
	}
	if dto.ID != "DEC-GET-001" || dto.Name != "Get Me" {
		t.Errorf("unexpected decision: %+v", dto)
	}
}

func TestGetDecision_NotFound(t *testing.T) {
	p := createDecisionCosmos(t)
	_, err := GetDecision(p, "identity.blumer.cloud", "ghost")
	if err == nil {
		t.Error("expected error for missing decision")
	}
}

func TestDeleteDecision(t *testing.T) {
	p := createDecisionCosmos(t)
	if _, err := CreateDecision(p, "identity.blumer.cloud", CreateDecisionRequest{
		ID: "DEC-DEL-001", Name: "Delete Me",
	}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteDecision(p, "identity.blumer.cloud", "DEC-DEL-001"); err != nil {
		t.Fatalf("DeleteDecision: %v", err)
	}
	_, err := GetDecision(p, "identity.blumer.cloud", "DEC-DEL-001")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestDeleteDecision_NotFound(t *testing.T) {
	p := createDecisionCosmos(t)
	err := DeleteDecision(p, "identity.blumer.cloud", "ghost")
	if err == nil {
		t.Error("expected error for missing decision")
	}
}

func TestDecisionIOsEqual(t *testing.T) {
	a := []model.DecisionIO{{Name: "result", Type: "boolean"}}
	b := []model.DecisionIO{{Name: "result", Type: "boolean"}}
	if !decisionIOsEqual(a, b) {
		t.Error("expected equal")
	}

	c := []model.DecisionIO{{Name: "result", Type: "string"}}
	if decisionIOsEqual(a, c) {
		t.Error("expected not equal (different type)")
	}

	d := []model.DecisionIO{{Name: "other", Type: "boolean"}}
	if decisionIOsEqual(a, d) {
		t.Error("expected not equal (different name)")
	}

	if decisionIOsEqual(a, nil) {
		t.Error("expected not equal (different length)")
	}

	if !decisionIOsEqual(nil, nil) {
		t.Error("expected equal for two nils")
	}
}

func TestNextDecisionID(t *testing.T) {
	p := createDecisionCosmos(t)
	// get domain path
	domainNode, err := findDomainNode(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatal(err)
	}
	id1 := nextDecisionID(domainNode.Path)
	if id1 == "" {
		t.Error("expected non-empty decision ID")
	}
	// create one decision so count increases
	if _, err := CreateDecision(p, "identity.blumer.cloud", CreateDecisionRequest{Name: "Auto ID Test"}); err != nil {
		t.Fatal(err)
	}
	id2 := nextDecisionID(domainNode.Path)
	if id2 == id1 {
		t.Errorf("expected different IDs, got same: %q", id1)
	}
}

func TestGetProductCollaboration(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{
		ID: "PROD-COLLAB-001", Name: "Collab Product",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-COLLAB-001", Name: "Main Process"}); err != nil {
		t.Fatal(err)
	}
	collab, err := GetProductCollaboration(p, product.ID)
	if err != nil {
		t.Fatalf("GetProductCollaboration: %v", err)
	}
	if collab.ProductID != product.ID {
		t.Errorf("unexpected product ID: %q", collab.ProductID)
	}
}

func TestGetProductCollaboration_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := GetProductCollaboration(p, "ghost-product")
	if err == nil {
		t.Error("expected error for missing product")
	}
}
