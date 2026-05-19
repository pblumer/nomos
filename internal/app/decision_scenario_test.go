package app

import (
	"testing"

	"github.com/nomos/nomos/internal/model"
)

func TestCreateAndListDecisionScenarios(t *testing.T) {
	p := createTraceTestCosmos(t)
	body := model.DecisionScenario{
		Name:                 "Premium eligible",
		Description:          "Premium customers are eligible",
		Tags:                 []string{"happy-path"},
		Enabled:              true,
		Inputs:               map[string]any{"category": "premium"},
		ExpectedOutputs:      map[string]any{"eligible": true},
		ExpectedMatchedRules: []string{"r1"},
	}
	created, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", body)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != "premium-eligible" {
		t.Fatalf("slug ID expected premium-eligible, got %q", created.ID)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("timestamps not set")
	}

	list, err := ListDecisionScenarios(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list.Count != 1 || len(list.Items) != 1 {
		t.Fatalf("list count: got %d", list.Count)
	}

	// Duplicate ID is rejected.
	if _, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", body); err == nil {
		t.Fatalf("expected duplicate-create error")
	}
}

func TestRunDecisionScenario_PassFail(t *testing.T) {
	p := createTraceTestCosmos(t)
	ok, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", model.DecisionScenario{
		Name:                 "premium happy",
		Enabled:              true,
		Inputs:               map[string]any{"category": "premium"},
		ExpectedOutputs:      map[string]any{"eligible": true},
		ExpectedMatchedRules: []string{"r1"},
	})
	if err != nil {
		t.Fatalf("create ok: %v", err)
	}
	bad, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", model.DecisionScenario{
		Name:            "wrong assertion",
		Enabled:         true,
		Inputs:          map[string]any{"category": "premium"},
		ExpectedOutputs: map[string]any{"eligible": false}, // intentionally wrong
	})
	if err != nil {
		t.Fatalf("create bad: %v", err)
	}

	resOK, err := RunDecisionScenario(p, "governance.blumer.com", "DEC-001", ok.ID, model.Evaluator{ID: "tester"})
	if err != nil {
		t.Fatalf("run ok: %v", err)
	}
	if !resOK.OK {
		t.Fatalf("expected pass, got %+v", resOK)
	}
	if resOK.TraceID == "" {
		t.Fatalf("expected trace_id on passing run")
	}

	resBad, err := RunDecisionScenario(p, "governance.blumer.com", "DEC-001", bad.ID, model.Evaluator{ID: "tester"})
	if err != nil {
		t.Fatalf("run bad: %v", err)
	}
	if resBad.OK {
		t.Fatalf("expected fail, got %+v", resBad)
	}
	if len(resBad.OutputDiff) == 0 {
		t.Fatalf("expected output_diff entries on failure")
	}
}

func TestRunAllDecisionScenarios_SkipsDisabled(t *testing.T) {
	p := createTraceTestCosmos(t)
	if _, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", model.DecisionScenario{
		Name: "enabled", Enabled: true, Inputs: map[string]any{"category": "premium"}, ExpectedOutputs: map[string]any{"eligible": true},
	}); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if _, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", model.DecisionScenario{
		Name: "disabled", Enabled: false, Inputs: map[string]any{"category": "premium"}, ExpectedOutputs: map[string]any{"eligible": true},
	}); err != nil {
		t.Fatalf("create b: %v", err)
	}
	batch, err := RunAllDecisionScenarios(p, "governance.blumer.com", "DEC-001", model.Evaluator{})
	if err != nil {
		t.Fatalf("run-all: %v", err)
	}
	if batch.Total != 2 || batch.Passed != 1 || batch.Skipped != 1 {
		t.Fatalf("counts: total=%d passed=%d skipped=%d failed=%d", batch.Total, batch.Passed, batch.Skipped, batch.Failed)
	}
}

func TestScenarioTrace_CarriesScenarioID(t *testing.T) {
	p := createTraceTestCosmos(t)
	s, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", model.DecisionScenario{
		Name: "tagged", Enabled: true, Inputs: map[string]any{"category": "premium"}, ExpectedOutputs: map[string]any{"eligible": true},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	res, err := RunDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID, model.Evaluator{ID: "tester"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	traces, err := ListDecisionTraces(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if traces.Count != 1 {
		t.Fatalf("expected 1 trace, got %d", traces.Count)
	}
	if traces.Items[0].ScenarioID != s.ID {
		t.Fatalf("trace.scenario_id=%q want %q", traces.Items[0].ScenarioID, s.ID)
	}
	if traces.Items[0].TraceID != res.TraceID {
		t.Fatalf("trace_id mismatch between scenario run (%s) and listed trace (%s)", res.TraceID, traces.Items[0].TraceID)
	}
}

func TestUpdateAndDeleteScenario(t *testing.T) {
	p := createTraceTestCosmos(t)
	s, err := CreateDecisionScenario(p, "governance.blumer.com", "DEC-001", model.DecisionScenario{
		Name: "case", Enabled: true, Inputs: map[string]any{"category": "premium"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	updated, err := UpdateDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID, model.DecisionScenario{
		Name: "case renamed", Description: "now with desc", Enabled: true,
		Inputs: map[string]any{"category": "basic"},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.ID != s.ID {
		t.Fatalf("update changed id: %q -> %q", s.ID, updated.ID)
	}
	if updated.Name != "case renamed" || updated.Inputs["category"] != "basic" {
		t.Fatalf("update did not persist fields: %+v", updated)
	}
	if updated.CreatedAt != s.CreatedAt {
		t.Fatalf("created_at must not change on update")
	}

	if err := DeleteDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := DeleteDecisionScenario(p, "governance.blumer.com", "DEC-001", s.ID); err == nil {
		t.Fatalf("expected 404 on second delete")
	}
}
