package graph

import (
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
)

func simpleGraph() model.Servicegraph {
	return model.Servicegraph{
		ID:   "SG-001",
		Type: "servicegraph",
		Name: "Simple",
		Nodes: []model.GraphNode{
			{ID: "A", Type: "precondition", Name: "Precond A"},
			{ID: "B", Type: "activity", Name: "Activity B"},
			{ID: "C", Type: "activity", Name: "Activity C"},
			{ID: "D", Type: "validation", Name: "Validate D"},
		},
		Edges: []model.GraphEdge{
			{ID: "E1", Source: "A", Target: "B", Type: "depends_on", Binding: "hard"},
			{ID: "E2", Source: "A", Target: "C", Type: "depends_on", Binding: "hard"},
			{ID: "E3", Source: "B", Target: "D", Type: "depends_on", Binding: "hard"},
			{ID: "E4", Source: "C", Target: "D", Type: "depends_on", Binding: "hard"},
		},
	}
}

func cyclicGraph() model.Servicegraph {
	return model.Servicegraph{
		ID:   "SG-CYC",
		Type: "servicegraph",
		Nodes: []model.GraphNode{
			{ID: "X", Type: "activity", Name: "X"},
			{ID: "Y", Type: "activity", Name: "Y"},
			{ID: "Z", Type: "activity", Name: "Z"},
		},
		Edges: []model.GraphEdge{
			{ID: "E1", Source: "X", Target: "Y", Type: "depends_on", Binding: "hard"},
			{ID: "E2", Source: "Y", Target: "Z", Type: "depends_on", Binding: "hard"},
			{ID: "E3", Source: "Z", Target: "X", Type: "depends_on", Binding: "hard"},
		},
	}
}

func TestTopologicalOrder(t *testing.T) {
	order, err := TopologicalOrder(simpleGraph())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A must come before B and C; B and C must come before D
	idxA, idxB, idxC, idxD := indexOf(order, "A"), indexOf(order, "B"), indexOf(order, "C"), indexOf(order, "D")
	if idxA >= idxB || idxA >= idxC {
		t.Errorf("A should come before B and C: %v", order)
	}
	if idxB >= idxD || idxC >= idxD {
		t.Errorf("B and C should come before D: %v", order)
	}
}

func TestTopologicalOrderCycle(t *testing.T) {
	_, err := TopologicalOrder(cyclicGraph())
	if err != ErrCycleDetected {
		t.Errorf("expected ErrCycleDetected, got %v", err)
	}
}

func TestParallelGroups(t *testing.T) {
	groups := ParallelGroups(simpleGraph())
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(groups), groups)
	}
	// Group 0: A (no deps)
	if !contains(groups[0], "A") {
		t.Errorf("group 0 should contain A: %v", groups[0])
	}
	// Group 1: B, C (both depend only on A)
	if !contains(groups[1], "B") || !contains(groups[1], "C") {
		t.Errorf("group 1 should contain B and C: %v", groups[1])
	}
	// Group 2: D
	if !contains(groups[2], "D") {
		t.Errorf("group 2 should contain D: %v", groups[2])
	}
}

func TestExecutableNodes(t *testing.T) {
	sg := simpleGraph()
	// Nothing completed: only A is executable (no deps)
	exec := ExecutableNodes(sg, map[string]bool{})
	if len(exec) != 1 || exec[0] != "A" {
		t.Errorf("expected [A], got %v", exec)
	}
	// A completed: B and C become executable
	exec = ExecutableNodes(sg, map[string]bool{"A": true})
	if len(exec) != 2 || exec[0] != "B" || exec[1] != "C" {
		t.Errorf("expected [B C], got %v", exec)
	}
	// A, B, C completed: D executable
	exec = ExecutableNodes(sg, map[string]bool{"A": true, "B": true, "C": true})
	if len(exec) != 1 || exec[0] != "D" {
		t.Errorf("expected [D], got %v", exec)
	}
}

func TestServicegraphMermaid(t *testing.T) {
	out := ServicegraphMermaid(simpleGraph())
	if !strings.HasPrefix(out, "graph TD") {
		t.Error("should start with graph TD")
	}
	if !strings.Contains(out, "-->|depends_on|") {
		t.Error("should contain hard edge notation")
	}
	// Check node shapes
	if !strings.Contains(out, "{{") {
		t.Error("precondition should use hexagon shape")
	}
	if !strings.Contains(out, "((") {
		t.Error("validation should use circle shape")
	}
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

func contains(slice []string, item string) bool {
	return indexOf(slice, item) >= 0
}
