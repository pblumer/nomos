package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

const traceTestDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="DEC-001" name="Provisioning Eligibility">
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1" label="Category">
        <inputExpression typeRef="string"><text>category</text></inputExpression>
      </input>
      <output id="o1" name="eligible" typeRef="boolean"/>
      <rule id="r1">
        <inputEntry><text>"premium"</text></inputEntry>
        <outputEntry><text>true</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>false</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

func createTraceTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	decDir := filepath.Join(storage.DecisionsDir(p), "DEC-001")
	must(os.MkdirAll(decDir, 0o755))
	must(os.WriteFile(storage.CosmosFile(p),
		[]byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(decDir, "decision.yaml"),
		[]byte("id: DEC-001\ntype: decision\nname: Provisioning Eligibility\nversion: 0.1.0\nstatus: draft\nowner: Team\ndmn_file: decision.dmn\n"), 0o644))
	must(os.WriteFile(filepath.Join(decDir, "decision.dmn"),
		[]byte(traceTestDMN), 0o644))
	return p
}

func TestEvaluateDecisionWithTrace_WritesArtifact(t *testing.T) {
	p := createTraceTestCosmos(t)
	req := EvaluateDecisionRequest{Inputs: map[string]any{"category": "premium"}}
	res, trace, err := EvaluateDecisionWithTrace(p, "DEC-001", req, model.Evaluator{ID: "tester", IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("EvaluateDecisionWithTrace: %v", err)
	}
	if res.Outputs["eligible"] != true {
		t.Fatalf("eligible = %v, want true", res.Outputs["eligible"])
	}
	if trace == nil || !strings.HasPrefix(trace.TraceID, "sha256:") {
		t.Fatalf("trace not built correctly: %+v", trace)
	}
	if trace.ParentTraceID != "" {
		t.Fatalf("first trace should have empty parent, got %q", trace.ParentTraceID)
	}
	if trace.Evaluator.ID != "tester" || trace.Evaluator.IP != "127.0.0.1" {
		t.Fatalf("evaluator not captured: %+v", trace.Evaluator)
	}

	// Verify the trace file exists on disk.
	tracesDir := filepath.Join(storage.DecisionsDir(p), "DEC-001", "traces")
	ents, err := os.ReadDir(tracesDir)
	if err != nil {
		t.Fatalf("read traces dir: %v", err)
	}
	if len(ents) != 1 {
		t.Fatalf("expected 1 trace file, got %d", len(ents))
	}
}

func TestEvaluateDecisionWithTrace_ChainsByParent(t *testing.T) {
	p := createTraceTestCosmos(t)
	req1 := EvaluateDecisionRequest{Inputs: map[string]any{"category": "premium"}}
	_, t1, err := EvaluateDecisionWithTrace(p, "DEC-001", req1, model.Evaluator{ID: "a"})
	if err != nil {
		t.Fatal(err)
	}
	req2 := EvaluateDecisionRequest{Inputs: map[string]any{"category": "basic"}}
	_, t2, err := EvaluateDecisionWithTrace(p, "DEC-001", req2, model.Evaluator{ID: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if t2.ParentTraceID != t1.TraceID {
		t.Fatalf("chain broken: t2.parent=%s want t1.TraceID=%s", t2.ParentTraceID, t1.TraceID)
	}
	if t1.TraceID == t2.TraceID {
		t.Fatal("two different evaluations must produce different trace IDs")
	}
}

func TestListAndGetDecisionTrace(t *testing.T) {
	p := createTraceTestCosmos(t)
	_, _, _ = EvaluateDecisionWithTrace(p, "DEC-001",
		EvaluateDecisionRequest{Inputs: map[string]any{"category": "premium"}}, model.Evaluator{})
	_, second, _ := EvaluateDecisionWithTrace(p, "DEC-001",
		EvaluateDecisionRequest{Inputs: map[string]any{"category": "basic"}}, model.Evaluator{})

	list, err := ListDecisionTraces(p, "DEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if list.Count != 2 || len(list.Items) != 2 {
		t.Fatalf("want 2 traces, got %d", list.Count)
	}
	got, err := GetDecisionTrace(p, "DEC-001", second.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	if got.TraceID != second.TraceID {
		t.Fatalf("GetDecisionTrace returned wrong trace")
	}
}

func TestVerifyDecisionTraces_DetectsTamper(t *testing.T) {
	p := createTraceTestCosmos(t)
	_, _, _ = EvaluateDecisionWithTrace(p, "DEC-001",
		EvaluateDecisionRequest{Inputs: map[string]any{"category": "premium"}}, model.Evaluator{})
	_, _, _ = EvaluateDecisionWithTrace(p, "DEC-001",
		EvaluateDecisionRequest{Inputs: map[string]any{"category": "basic"}}, model.Evaluator{})

	clean, err := VerifyDecisionTraces(p, "DEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if !clean.OK {
		t.Fatalf("clean traces should verify ok, got: %+v", clean)
	}

	// Tamper with one file on disk.
	tracesDir := filepath.Join(storage.DecisionsDir(p), "DEC-001", "traces")
	ents, _ := os.ReadDir(tracesDir)
	if len(ents) == 0 {
		t.Fatal("no traces written")
	}
	first := filepath.Join(tracesDir, ents[0].Name())
	data, _ := os.ReadFile(first)
	tampered := strings.Replace(string(data), "category: premium", "category: basic", 1)
	if tampered == string(data) {
		t.Skip("tamper substitution did not change file contents")
	}
	if err := os.WriteFile(first, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	bad, err := VerifyDecisionTraces(p, "DEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if bad.OK {
		t.Fatalf("tampered traces must NOT verify ok, got: %+v", bad)
	}
}
