package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
)

// Scenarios share the traceTestDMN cosmos fixture defined in
// decision_trace_test.go so we can promote real traces into scenarios.

func TestCreateDecisionScenario_FreeForm(t *testing.T) {
	p := createTraceTestCosmos(t)
	req := CreateDecisionScenarioRequest{
		Name:            "Premium happy path",
		Description:     "expects eligible=true",
		Inputs:          map[string]any{"category": "premium"},
		ExpectedOutputs: map[string]any{"eligible": true},
	}
	s, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", req)
	if err != nil {
		t.Fatalf("CreateDecisionScenario: %v", err)
	}
	if !strings.HasPrefix(s.ID, "VSC_") {
		t.Errorf("scenario ID prefix: got %q, want VSC_…", s.ID)
	}
	if s.Name != "Premium happy path" {
		t.Errorf("name: got %q", s.Name)
	}
	if s.Inputs["category"] != "premium" {
		t.Errorf("inputs not persisted: %v", s.Inputs)
	}
	if s.ExpectedOutputs["eligible"] != true {
		t.Errorf("expected_outputs not persisted: %v", s.ExpectedOutputs)
	}
	if s.SourceTraceID != "" {
		t.Errorf("free-form scenario should not have a source_trace_id, got %q", s.SourceTraceID)
	}
	if s.CreatedAt == "" {
		t.Error("created_at not set")
	}
}

func TestCreateDecisionScenario_RequiresNameAndInputs(t *testing.T) {
	p := createTraceTestCosmos(t)
	if _, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", CreateDecisionScenarioRequest{
		Inputs: map[string]any{"category": "premium"},
	}); err == nil {
		t.Error("missing name should error")
	}
	if _, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", CreateDecisionScenarioRequest{
		Name: "x",
	}); err == nil {
		t.Error("missing inputs should error")
	}
}

func TestCreateDecisionScenarioFromTrace(t *testing.T) {
	p := createTraceTestCosmos(t)
	req := EvaluateDecisionRequest{Inputs: map[string]any{"category": "premium"}}
	_, trace, err := EvaluateDecisionWithTrace(p, "governance.blumer.com", "DEC-001", req, model.Evaluator{ID: "tester"})
	if err != nil {
		t.Fatalf("EvaluateDecisionWithTrace: %v", err)
	}
	s, err := CreateDecisionScenarioFromTrace(p, "governance.blumer.com", "DEC-001", trace.TraceID, "Premium baseline", "")
	if err != nil {
		t.Fatalf("CreateDecisionScenarioFromTrace: %v", err)
	}
	if s.SourceTraceID != trace.TraceID {
		t.Errorf("source_trace_id: got %q want %q", s.SourceTraceID, trace.TraceID)
	}
	if s.Inputs["category"] != "premium" {
		t.Errorf("inputs not seeded from trace: %v", s.Inputs)
	}
	if s.ExpectedOutputs["eligible"] != true {
		t.Errorf("expected_outputs not seeded from trace outputs: %v", s.ExpectedOutputs)
	}
	// Mutating the scenario's maps must not bleed back into the trace.
	s.Inputs["category"] = "mutated"
	got, _ := GetDecisionTrace(p, "governance.blumer.com", "DEC-001", trace.TraceID)
	if got.Inputs["category"] != "premium" {
		t.Errorf("scenario mutation leaked into trace: %v", got.Inputs)
	}
}

func TestCreateDecisionScenario_DispatchesToFromTrace(t *testing.T) {
	p := createTraceTestCosmos(t)
	_, trace, err := EvaluateDecisionWithTrace(p, "governance.blumer.com", "DEC-001",
		EvaluateDecisionRequest{Inputs: map[string]any{"category": "premium"}}, model.Evaluator{})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	s, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", CreateDecisionScenarioRequest{
		Name:        "via-dispatch",
		FromTraceID: trace.TraceID,
	})
	if err != nil {
		t.Fatalf("CreateDecisionScenario w/ from_trace_id: %v", err)
	}
	if s.SourceTraceID != trace.TraceID {
		t.Errorf("expected dispatch to from-trace path, source_trace_id=%q", s.SourceTraceID)
	}
}

func TestListDecisionScenarios_SortedAndPersisted(t *testing.T) {
	p := createTraceTestCosmos(t)
	mk := func(name string) {
		if _, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", CreateDecisionScenarioRequest{
			Name:   name,
			Inputs: map[string]any{"category": "premium"},
		}); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	mk("first")
	mk("second")
	mk("third")
	dto, err := ListDecisionScenarios(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("ListDecisionScenarios: %v", err)
	}
	if dto.Count != 3 || len(dto.Items) != 3 {
		t.Fatalf("count mismatch: %+v", dto)
	}
	// Oldest first (creation order).
	if dto.Items[0].Name != "first" || dto.Items[2].Name != "third" {
		t.Errorf("unexpected order: %s, %s, %s", dto.Items[0].Name, dto.Items[1].Name, dto.Items[2].Name)
	}
	// Confirm YAML files were actually written.
	ents, _ := os.ReadDir(filepath.Join(p, ".nomos", "domains", "governance.blumer.com", "decisions", "DEC-001", "scenarios"))
	if len(ents) != 3 {
		t.Errorf("expected 3 scenario files on disk, got %d", len(ents))
	}
}

func TestUpdateDecisionScenario_PartialAndClear(t *testing.T) {
	p := createTraceTestCosmos(t)
	s, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", CreateDecisionScenarioRequest{
		Name:            "orig",
		Inputs:          map[string]any{"category": "premium"},
		ExpectedOutputs: map[string]any{"eligible": true},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	newName := "renamed"
	upd, err := UpdateDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID, UpdateDecisionScenarioRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("update name: %v", err)
	}
	if upd.Name != "renamed" {
		t.Errorf("name not updated: %q", upd.Name)
	}
	if upd.Inputs["category"] != "premium" {
		t.Errorf("inputs should be preserved: %v", upd.Inputs)
	}
	if upd.ExpectedOutputs["eligible"] != true {
		t.Errorf("expected_outputs should be preserved: %v", upd.ExpectedOutputs)
	}
	if upd.UpdatedAt == "" {
		t.Error("updated_at should be set after update")
	}
	// Clear expected outputs.
	upd2, err := UpdateDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID, UpdateDecisionScenarioRequest{
		ClearExpected: true,
	})
	if err != nil {
		t.Fatalf("clear expected: %v", err)
	}
	if upd2.ExpectedOutputs != nil {
		t.Errorf("expected_outputs should be cleared, got %v", upd2.ExpectedOutputs)
	}
}

func TestDeleteDecisionScenario(t *testing.T) {
	p := createTraceTestCosmos(t)
	s, _ := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", CreateDecisionScenarioRequest{
		Name:   "doomed",
		Inputs: map[string]any{"category": "premium"},
	})
	if err := DeleteDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := GetDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID); err == nil {
		t.Error("scenario should be gone after delete")
	}
	if err := DeleteDecisionScenario(p, "governance.blumer.com", "DEC-001", "VSC_XXXXXX"); err == nil {
		t.Error("delete on unknown id should 404")
	}
}
