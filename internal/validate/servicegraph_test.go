package validate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

func setupSGTestCosmos(t *testing.T, sgFile string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(storage.NomosDir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(storage.CosmosFile(dir), []byte("id: test\ntype: cosmos\nname: Test\nversion: 0.1.0\nstatus: active\nowner: test\nsummary: test\n"), 0644)
	sgDir := filepath.Join(storage.CatalogDir(dir), "servicegraphs")
	os.MkdirAll(sgDir, 0755)
	data, err := os.ReadFile(sgFile)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(sgDir, "test.yaml"), data, 0644)
	return dir
}

func TestValidateServicegraphValid(t *testing.T) {
	dir := setupSGTestCosmos(t, "testdata/valid-servicegraph.yaml")
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range res.Findings {
		if f.Severity == "error" {
			t.Errorf("unexpected error finding: %s - %s", f.Code, f.Message)
		}
	}
}

func TestValidateServicegraphCycle(t *testing.T) {
	dir := setupSGTestCosmos(t, "testdata/invalid-cycle.yaml")
	res, _ := Validate(dir)
	found := false
	for _, f := range res.Findings {
		if f.Code == "SG_CYCLE_DETECTED" {
			found = true
		}
	}
	if !found {
		t.Error("expected SG_CYCLE_DETECTED finding")
	}
}

func TestValidateServicegraphMissingFields(t *testing.T) {
	dir := setupSGTestCosmos(t, "testdata/invalid-missing-fields.yaml")
	res, _ := Validate(dir)
	codes := make(map[string]bool)
	for _, f := range res.Findings {
		codes[f.Code] = true
	}
	expected := []string{"SG_REQUIRED_FIELD", "SG_NO_NODES", "SG_NO_EDGES"}
	for _, code := range expected {
		if !codes[code] {
			t.Errorf("expected finding %s not found", code)
		}
	}
}

func TestValidateServicegraphNodeEdgeErrors(t *testing.T) {
	sg := model.Servicegraph{
		ID: "SG-001", Type: "servicegraph", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "x",
		Nodes: []model.GraphNode{
			{ID: "N-001", Type: "invalid_type", Name: "Bad"},
			{ID: "N-001", Type: "activity", Name: "Dup"}, // duplicate
		},
		Edges: []model.GraphEdge{
			{ID: "E-001", Source: "N-001", Target: "MISSING", Type: "depends_on", Binding: "hard"},
			{ID: "E-002", Source: "N-001", Target: "N-001", Type: "invalid_edge", Binding: "invalid"},
			{ID: "E-001", Source: "N-001", Target: "N-001", Type: "depends_on", Binding: "hard"}, // dup edge
		},
	}
	res := Result{Status: "ok", Findings: []Finding{}}
	validateServicegraph(sg, "test.yaml", &res)
	codes := make(map[string]int)
	for _, f := range res.Findings {
		codes[f.Code]++
	}
	assertions := map[string]bool{
		"SG_NODE_TYPE_INVALID":    codes["SG_NODE_TYPE_INVALID"] > 0,
		"SG_NODE_ID_DUPLICATE":    codes["SG_NODE_ID_DUPLICATE"] > 0,
		"SG_EDGE_TARGET_MISSING":  codes["SG_EDGE_TARGET_MISSING"] > 0,
		"SG_EDGE_TYPE_INVALID":    codes["SG_EDGE_TYPE_INVALID"] > 0,
		"SG_EDGE_BINDING_INVALID": codes["SG_EDGE_BINDING_INVALID"] > 0,
		"SG_EDGE_ID_DUPLICATE":    codes["SG_EDGE_ID_DUPLICATE"] > 0,
		"SG_NO_SERVICE_NODE":      codes["SG_NO_SERVICE_NODE"] > 0,
		"SG_NO_COMPOSED_OF":       codes["SG_NO_COMPOSED_OF"] > 0,
	}
	for code, ok := range assertions {
		if !ok {
			t.Errorf("expected %s finding", code)
		}
	}
}

func TestValidateServicegraphRuleScopeMissing(t *testing.T) {
	sg := model.Servicegraph{
		ID: "SG-001", Type: "servicegraph", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "x",
		Nodes: []model.GraphNode{
			{ID: "N-001", Type: "service", Name: "Svc"},
		},
		Edges: []model.GraphEdge{
			{ID: "E-001", Source: "N-001", Target: "N-001", Type: "composed_of", Binding: "hard"},
		},
		Rules: []model.GraphRule{
			{ID: "G-001", Name: "Test", Type: "activation", Scope: "NONEXISTENT", Binding: "must"},
		},
	}
	res := Result{Status: "ok", Findings: []Finding{}}
	validateServicegraph(sg, "test.yaml", &res)
	found := false
	for _, f := range res.Findings {
		if f.Code == "SG_RULE_SCOPE_MISSING" {
			found = true
		}
	}
	if !found {
		t.Error("expected SG_RULE_SCOPE_MISSING")
	}
}
