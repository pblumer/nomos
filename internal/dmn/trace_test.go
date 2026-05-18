package dmn

import (
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
)

func sampleTraceInput() TraceBuildInput {
	return TraceBuildInput{
		Domain:          "ch.blumer.governance",
		DecisionID:      "DEC-001",
		DecisionName:    "Provisioning Eligibility",
		DecisionVersion: "0.1.0",
		RuleBytes:       []byte("<dmn>fixture</dmn>"),
		Inputs:          map[string]any{"age": 18.0, "category": "premium"},
		Result: &Result{
			Outputs:      map[string]any{"eligible": true, "level": "gold"},
			MatchedRules: []string{"R1"},
			HitPolicy:    "UNIQUE",
		},
		Evaluator: model.Evaluator{ID: "user@example.com", IP: "127.0.0.1"},
		Engine:    model.EngineInfo{Name: "nomos", Version: "dev"},
		Timestamp: "2026-05-18T14:30:00.000000000Z",
	}
}

func TestBuildTrace_HashDeterministic(t *testing.T) {
	in := sampleTraceInput()
	tr1, err := BuildTrace(in)
	if err != nil {
		t.Fatalf("BuildTrace: %v", err)
	}
	tr2, err := BuildTrace(in)
	if err != nil {
		t.Fatalf("BuildTrace: %v", err)
	}
	if tr1.TraceID == "" {
		t.Fatal("TraceID is empty")
	}
	if !strings.HasPrefix(tr1.TraceID, "sha256:") {
		t.Fatalf("TraceID missing prefix: %s", tr1.TraceID)
	}
	if tr1.TraceID != tr2.TraceID {
		t.Fatalf("trace hash not deterministic: %s vs %s", tr1.TraceID, tr2.TraceID)
	}
}

func TestBuildTrace_InputOrderIndependent(t *testing.T) {
	a := sampleTraceInput()
	a.Inputs = map[string]any{"age": 18.0, "category": "premium"}
	b := sampleTraceInput()
	b.Inputs = map[string]any{"category": "premium", "age": 18.0}

	ta, _ := BuildTrace(a)
	tb, _ := BuildTrace(b)
	if ta.TraceID != tb.TraceID {
		t.Fatalf("map key order changed hash: %s vs %s", ta.TraceID, tb.TraceID)
	}
}

func TestBuildTrace_DifferentInputsDifferentHash(t *testing.T) {
	a := sampleTraceInput()
	b := sampleTraceInput()
	b.Inputs = map[string]any{"age": 19.0, "category": "premium"}

	ta, _ := BuildTrace(a)
	tb, _ := BuildTrace(b)
	if ta.TraceID == tb.TraceID {
		t.Fatal("expected different hashes for different inputs")
	}
}

func TestBuildTrace_DifferentRulesDifferentHash(t *testing.T) {
	a := sampleTraceInput()
	b := sampleTraceInput()
	b.RuleBytes = []byte("<dmn>different</dmn>")

	ta, _ := BuildTrace(a)
	tb, _ := BuildTrace(b)
	if ta.TraceID == tb.TraceID {
		t.Fatal("expected different hashes for different rules")
	}
	if ta.RuleHash == tb.RuleHash {
		t.Fatal("expected different rule_hash for different rules")
	}
}

func TestBuildTrace_ParentChain(t *testing.T) {
	a := sampleTraceInput()
	first, _ := BuildTrace(a)

	b := sampleTraceInput()
	b.Timestamp = "2026-05-18T14:30:01.000000000Z"
	b.ParentTraceID = first.TraceID
	second, _ := BuildTrace(b)

	if second.ParentTraceID != first.TraceID {
		t.Fatalf("parent chain broken")
	}
	if first.TraceID == second.TraceID {
		t.Fatal("parent ref should affect hash")
	}
}

func TestVerifyTrace_OK(t *testing.T) {
	tr, _ := BuildTrace(sampleTraceInput())
	if err := VerifyTrace(tr); err != nil {
		t.Fatalf("VerifyTrace failed on valid trace: %v", err)
	}
}

func TestVerifyTrace_TamperDetected(t *testing.T) {
	tr, _ := BuildTrace(sampleTraceInput())
	// Tamper with the outputs after the fact.
	tr.Outputs = map[string]any{"eligible": false}
	if err := VerifyTrace(tr); err == nil {
		t.Fatal("expected tamper to be detected, got nil error")
	}
}
