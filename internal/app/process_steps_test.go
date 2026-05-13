package app

import (
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
)

// helpers for process-step tests

func createProcessForStepTests(t *testing.T) (path, productID, processID string) {
	t.Helper()
	p := createAppTestCosmos(t)
	prod, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{
		ID: "PROD-STEP-001", Name: "Step Test Product", Summary: "for step tests",
	})
	if err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, prod.ID, CreateProcessRequest{
		ID: "PRC-STEP-001", Name: "Step Test Process",
	})
	if err != nil {
		t.Fatal(err)
	}
	return p, prod.ID, proc.ID
}

// ---------------------------------------------------------------------------
// AddProcessStep
// ---------------------------------------------------------------------------

func TestAddProcessStep_Success(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	got, err := AddProcessStep(p, procID, UpsertProcessStepRequest{
		Name:       "Provision Account",
		ServiceRef: "identity.blumer.cloud/user-account",
		Method:     "create",
		Role:       "primary",
		Required:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(got.Steps))
	}
	s := got.Steps[0]
	if s.Name != "Provision Account" || s.ServiceRef != "identity.blumer.cloud/user-account" || s.Method != "create" {
		t.Fatalf("unexpected step: %+v", s)
	}
	if s.ID == "" {
		t.Fatal("step ID must be set")
	}
}

func TestAddProcessStep_EmptyName(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	_, err := AddProcessStep(p, procID, UpsertProcessStepRequest{Name: "  "})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddProcessStep_ProcessNotFound(t *testing.T) {
	p := createAppTestCosmos(t)

	_, err := AddProcessStep(p, "PRC-DOES-NOT-EXIST", UpsertProcessStepRequest{Name: "Step"})
	if err == nil {
		t.Fatal("expected error for missing process")
	}
}

func TestAddProcessStep_WithDependsOn(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	first, err := AddProcessStep(p, procID, UpsertProcessStepRequest{
		Name: "First Step", ServiceRef: "identity.blumer.cloud/user-account",
	})
	if err != nil {
		t.Fatal(err)
	}
	firstID := first.Steps[0].ID

	second, err := AddProcessStep(p, procID, UpsertProcessStepRequest{
		Name:       "Second Step",
		ServiceRef: "identity.blumer.cloud/user-account",
		DependsOn:  []string{firstID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(second.Steps))
	}
	dep := second.Steps[1].DependsOn
	if len(dep) != 1 || dep[0] != firstID {
		t.Fatalf("expected depends_on [%s], got %v", firstID, dep)
	}
}

func TestAddProcessStep_WithInputsAndOutputs(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	got, err := AddProcessStep(p, procID, UpsertProcessStepRequest{
		Name:       "Scoped Step",
		ServiceRef: "identity.blumer.cloud/user-account",
		Inputs: []StepInputBindingDTO{
			{Name: "username", Source: "instance.username", Required: true},
			{Name: "domain", Source: "instance.domain"},
		},
		Outputs: []StepOutputSchemaDTO{
			{Name: "accountID", Type: "string", Description: "created account id"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := got.Steps[0]
	if len(s.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(s.Inputs))
	}
	if s.Inputs[0].Name != "username" || s.Inputs[0].Source != "instance.username" || !s.Inputs[0].Required {
		t.Fatalf("unexpected input[0]: %+v", s.Inputs[0])
	}
	if s.Inputs[1].Name != "domain" || s.Inputs[1].Required {
		t.Fatalf("unexpected input[1]: %+v", s.Inputs[1])
	}
	if len(s.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(s.Outputs))
	}
	if s.Outputs[0].Name != "accountID" || s.Outputs[0].Type != "string" {
		t.Fatalf("unexpected output: %+v", s.Outputs[0])
	}
}

// ---------------------------------------------------------------------------
// UpdateProcessStep
// ---------------------------------------------------------------------------

func TestUpdateProcessStep_Success(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	added, err := AddProcessStep(p, procID, UpsertProcessStepRequest{
		Name: "Original Name", ServiceRef: "identity.blumer.cloud/user-account",
	})
	if err != nil {
		t.Fatal(err)
	}
	stepID := added.Steps[0].ID

	updated, err := UpdateProcessStep(p, procID, stepID, UpsertProcessStepRequest{
		Name:       "Updated Name",
		ServiceRef: "identity.blumer.cloud/user-account",
		Method:     "update",
		Required:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	s := updated.Steps[0]
	if s.Name != "Updated Name" || s.Method != "update" {
		t.Fatalf("update not persisted: %+v", s)
	}
}

func TestUpdateProcessStep_UpdatesInputsAndOutputs(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	added, err := AddProcessStep(p, procID, UpsertProcessStepRequest{
		Name: "Step", ServiceRef: "identity.blumer.cloud/user-account",
		Inputs:  []StepInputBindingDTO{{Name: "old", Source: "instance.old"}},
		Outputs: []StepOutputSchemaDTO{{Name: "result", Type: "string"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	stepID := added.Steps[0].ID

	updated, err := UpdateProcessStep(p, procID, stepID, UpsertProcessStepRequest{
		Name:       "Step",
		ServiceRef: "identity.blumer.cloud/user-account",
		Inputs:     []StepInputBindingDTO{{Name: "new", Source: "instance.new", Required: true}},
		Outputs:    []StepOutputSchemaDTO{{Name: "id", Type: "number", Description: "numeric id"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := updated.Steps[0]
	if len(s.Inputs) != 1 || s.Inputs[0].Name != "new" || !s.Inputs[0].Required {
		t.Fatalf("inputs not updated: %+v", s.Inputs)
	}
	if len(s.Outputs) != 1 || s.Outputs[0].Name != "id" || s.Outputs[0].Type != "number" {
		t.Fatalf("outputs not updated: %+v", s.Outputs)
	}
}

func TestUpdateProcessStep_EmptyName(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	added, err := AddProcessStep(p, procID, UpsertProcessStepRequest{Name: "Step"})
	if err != nil {
		t.Fatal(err)
	}
	stepID := added.Steps[0].ID

	_, err = UpdateProcessStep(p, procID, stepID, UpsertProcessStepRequest{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateProcessStep_NotFound(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	_, err := UpdateProcessStep(p, procID, "step-does-not-exist", UpsertProcessStepRequest{Name: "X"})
	if err == nil {
		t.Fatal("expected error for missing step")
	}
}

// ---------------------------------------------------------------------------
// RemoveProcessStep
// ---------------------------------------------------------------------------

func TestRemoveProcessStep_Success(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	added, err := AddProcessStep(p, procID, UpsertProcessStepRequest{Name: "To Remove"})
	if err != nil {
		t.Fatal(err)
	}
	stepID := added.Steps[0].ID

	result, err := RemoveProcessStep(p, procID, stepID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 0 {
		t.Fatalf("expected 0 steps after remove, got %d", len(result.Steps))
	}
}

func TestRemoveProcessStep_OnlyRemovesTarget(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	r1, err := AddProcessStep(p, procID, UpsertProcessStepRequest{Name: "Keep"})
	if err != nil {
		t.Fatal(err)
	}
	keepID := r1.Steps[0].ID

	r2, err := AddProcessStep(p, procID, UpsertProcessStepRequest{Name: "Remove"})
	if err != nil {
		t.Fatal(err)
	}
	removeID := r2.Steps[1].ID

	result, err := RemoveProcessStep(p, procID, removeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Steps) != 1 || result.Steps[0].ID != keepID {
		t.Fatalf("wrong step retained: %+v", result.Steps)
	}
}

func TestRemoveProcessStep_NotFound(t *testing.T) {
	p, _, procID := createProcessForStepTests(t)

	_, err := RemoveProcessStep(p, procID, "step-does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing step")
	}
}

// ---------------------------------------------------------------------------
// IO conversion helpers (unit tests — no filesystem)
// ---------------------------------------------------------------------------

func TestStepsInputsToDTO_RoundTrip(t *testing.T) {
	ins := []model.StepInputBinding{
		{Name: "a", Source: "instance.a", Required: true},
		{Name: "b", Source: "step.s1.out", Required: false},
	}
	dtos := stepsInputsToDTO(ins)
	if len(dtos) != 2 {
		t.Fatalf("expected 2, got %d", len(dtos))
	}
	if dtos[0].Name != "a" || dtos[0].Source != "instance.a" || !dtos[0].Required {
		t.Fatalf("unexpected dto[0]: %+v", dtos[0])
	}
	if dtos[1].Name != "b" || dtos[1].Required {
		t.Fatalf("unexpected dto[1]: %+v", dtos[1])
	}
}

func TestStepsOutputsToDTO_RoundTrip(t *testing.T) {
	outs := []model.StepOutputSchema{
		{Name: "id", Type: "string", Description: "the id"},
		{Name: "count", Type: "number"},
	}
	dtos := stepsOutputsToDTO(outs)
	if len(dtos) != 2 {
		t.Fatalf("expected 2, got %d", len(dtos))
	}
	if dtos[0].Name != "id" || dtos[0].Type != "string" || dtos[0].Description != "the id" {
		t.Fatalf("unexpected dto[0]: %+v", dtos[0])
	}
	if dtos[1].Name != "count" || dtos[1].Type != "number" || dtos[1].Description != "" {
		t.Fatalf("unexpected dto[1]: %+v", dtos[1])
	}
}

func TestDtoInputsToModel_RoundTrip(t *testing.T) {
	dtos := []StepInputBindingDTO{
		{Name: "x", Source: "instance.x", Required: true},
		{Name: "y", Source: "step.s2.out"},
	}
	models := dtoInputsToModel(dtos)
	if len(models) != 2 {
		t.Fatalf("expected 2, got %d", len(models))
	}
	if models[0].Name != "x" || models[0].Source != "instance.x" || !models[0].Required {
		t.Fatalf("unexpected model[0]: %+v", models[0])
	}
	if models[1].Name != "y" || models[1].Required {
		t.Fatalf("unexpected model[1]: %+v", models[1])
	}
}

func TestDtoOutputsToModel_RoundTrip(t *testing.T) {
	dtos := []StepOutputSchemaDTO{
		{Name: "token", Type: "string", Description: "auth token"},
	}
	models := dtoOutputsToModel(dtos)
	if len(models) != 1 {
		t.Fatalf("expected 1, got %d", len(models))
	}
	if models[0].Name != "token" || models[0].Type != "string" || models[0].Description != "auth token" {
		t.Fatalf("unexpected model: %+v", models[0])
	}
}

func TestIOConversionHelpers_EmptySlices(t *testing.T) {
	if dtos := stepsInputsToDTO(nil); len(dtos) != 0 {
		t.Fatalf("expected empty, got %v", dtos)
	}
	if dtos := stepsOutputsToDTO(nil); len(dtos) != 0 {
		t.Fatalf("expected empty, got %v", dtos)
	}
	if ms := dtoInputsToModel(nil); len(ms) != 0 {
		t.Fatalf("expected empty, got %v", ms)
	}
	if ms := dtoOutputsToModel(nil); len(ms) != 0 {
		t.Fatalf("expected empty, got %v", ms)
	}
}
