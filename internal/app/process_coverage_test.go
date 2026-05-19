package app

import (
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
)

func createProcessTestCosmos(t *testing.T) (string, string) {
	t.Helper()
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{
		ID: "PROD-COV-001", Name: "Coverage Product",
	})
	if err != nil {
		t.Fatal(err)
	}
	return p, product.ID
}

func TestListProcesses_Empty(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := ListProcesses(p)
	if err != nil {
		t.Fatalf("ListProcesses: %v", err)
	}
	if dto.Count != 0 {
		t.Errorf("expected 0 processes, got %d", dto.Count)
	}
}

func TestListProcesses_WithProcess(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	if _, err := CreateProductProcess(p, productID, CreateProcessRequest{ID: "PRC-LIST-001", Name: "List Test"}); err != nil {
		t.Fatal(err)
	}
	dto, err := ListProcesses(p)
	if err != nil {
		t.Fatalf("ListProcesses: %v", err)
	}
	if dto.Count == 0 {
		t.Error("expected at least one process")
	}
}

func TestValidateProcess(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	proc, err := CreateProductProcess(p, productID, CreateProcessRequest{ID: "PRC-VAL-001", Name: "Validate Test"})
	if err != nil {
		t.Fatal(err)
	}
	val, err := ValidateProcess(p, proc.ID)
	if err != nil {
		t.Fatalf("ValidateProcess: %v", err)
	}
	// result can be valid or invalid; we just want no error
	_ = val
}

func TestValidateProcess_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := ValidateProcess(p, "ghost-id")
	if err == nil {
		t.Error("expected error for missing process")
	}
}

func TestNormalizeDMNTypeRef(t *testing.T) {
	cases := []struct{ in, want string }{
		{"boolean", "boolean"},
		{"Boolean", "boolean"},
		{"integer", "number"},
		{"long", "number"},
		{"double", "number"},
		{"decimal", "number"},
		{"number", "number"},
		{"string", "string"},
		{"String", "string"},
		{"date", "date"},
		{"date and time", "date"},
		{"datetime", "date"},
		{"  Boolean  ", "boolean"},
		{"custom", "custom"},
	}
	for _, c := range cases {
		got := normalizeDMNTypeRef(c.in)
		if got != c.want {
			t.Errorf("normalizeDMNTypeRef(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsHappyGatewayValue(t *testing.T) {
	trueVals := []string{"true", "True", "TRUE", "yes", "Yes", "YES", "1", "  true  "}
	for _, v := range trueVals {
		if !isHappyGatewayValue(v) {
			t.Errorf("expected true for %q", v)
		}
	}
	falseVals := []string{"false", "no", "0", ""}
	for _, v := range falseVals {
		if isHappyGatewayValue(v) {
			t.Errorf("expected false for %q", v)
		}
	}
}

func TestGatewayBranches_ExplicitConditions(t *testing.T) {
	steps := []model.ProcessStep{
		{ID: "step-rule", TaskType: "businessRuleTask", DecisionRef: ""},
		{ID: "step-gw", TaskType: "exclusiveGateway", Gateway: &model.DecisionGateway{
			Conditions: []model.GatewayCondition{
				{Output: "result", Operator: "==", Value: "true", Label: "Ja"},
				{Output: "result", Operator: "==", Value: "false", Label: "Nein"},
			},
		}},
		{ID: "step-next", TaskType: "serviceTask"},
	}
	branches := gatewayBranches("", steps, 1, "step-gw", "step-next")
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	if branches[0].label != "Ja" {
		t.Errorf("branch 0 label = %q", branches[0].label)
	}
	if !strings.Contains(branches[0].cond, "result") {
		t.Errorf("branch 0 condition missing output var: %q", branches[0].cond)
	}
}

func TestGatewayBranches_AutoResolve(t *testing.T) {
	steps := []model.ProcessStep{
		{ID: "step-rule", TaskType: "businessRuleTask", DecisionRef: ""},
		{ID: "step-gw", TaskType: "exclusiveGateway"},
		{ID: "step-next", TaskType: "serviceTask"},
	}
	branches := gatewayBranches("", steps, 1, "step-gw", "step-next")
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches (auto-resolve), got %d", len(branches))
	}
	if branches[0].label != "Ja" || branches[1].label != "Nein" {
		t.Errorf("unexpected labels: %q %q", branches[0].label, branches[1].label)
	}
}

func TestGatewayBranches_NonHappyPathCreatesErrEnd(t *testing.T) {
	steps := []model.ProcessStep{
		{ID: "step-gw", TaskType: "exclusiveGateway", Gateway: &model.DecisionGateway{
			Conditions: []model.GatewayCondition{
				{Output: "canProceed", Value: "false", Label: "Rejected"},
			},
		}},
	}
	branches := gatewayBranches("", steps, 0, "step-gw", "EndEvent")
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches))
	}
	if branches[0].errEndID == "" {
		t.Error("expected errEndID for non-happy path")
	}
}

func TestUpdateProcessParticipant(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	proc, err := CreateProductProcess(p, productID, CreateProcessRequest{ID: "PRC-PART-001", Name: "Part Test"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateProcessParticipant(p, proc.ID, UpdateParticipantRequest{
		Ref:  "identity.blumer.cloud",
		Name: "Identity Domain",
	})
	if err != nil {
		t.Fatalf("UpdateProcessParticipant: %v", err)
	}
	if updated.Participant.Ref != "identity.blumer.cloud" {
		t.Errorf("unexpected participant ref: %q", updated.Participant.Ref)
	}
}

func TestUpdateProcessParticipant_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := UpdateProcessParticipant(p, "ghost", UpdateParticipantRequest{})
	if err == nil {
		t.Error("expected error for missing process")
	}
}

func TestDeleteProcess(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	proc, err := CreateProductProcess(p, productID, CreateProcessRequest{ID: "PRC-DEL-001", Name: "Delete Test"})
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteProcess(p, proc.ID); err != nil {
		t.Fatalf("DeleteProcess: %v", err)
	}
	_, err = GetProcess(p, proc.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestDeleteProcess_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	if err := DeleteProcess(p, "ghost"); err == nil {
		t.Error("expected error for missing process")
	}
}
