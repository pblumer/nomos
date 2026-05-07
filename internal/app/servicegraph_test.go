package app

import (
	"os"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

func setupSGCosmos(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(storage.NomosDir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(storage.CosmosFile(dir), []byte("id: test\ntype: cosmos\nname: Test\nversion: 0.1.0\nstatus: active\nowner: test\nsummary: test\ndomains: []\n"), 0644)
	os.MkdirAll(storage.DomainsDir(dir), 0755)
	return dir
}

func testServicegraph() model.Servicegraph {
	return model.Servicegraph{
		ID: "SG-TEST-001", Type: "servicegraph", Name: "Test SG",
		Version: "0.1.0", Status: "draft", Owner: "test",
		Summary: "test", RelatedProduct: "PROD-001",
		Nodes: []model.GraphNode{
			{ID: "N-001", Type: "service", Name: "Svc", Mandatory: true},
			{ID: "N-002", Type: "activity", Name: "Act", Mandatory: true, Reusable: true},
		},
		Edges: []model.GraphEdge{
			{ID: "E-001", Source: "N-001", Target: "N-002", Type: "composed_of", Binding: "hard"},
		},
	}
}

func TestCreateAndListServicegraph(t *testing.T) {
	dir := setupSGCosmos(t)
	sg := testServicegraph()

	if err := CreateServicegraph(dir, sg); err != nil {
		t.Fatalf("create: %v", err)
	}

	list, err := ListServicegraphs(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list.Count != 1 {
		t.Fatalf("expected 1, got %d", list.Count)
	}
	if list.Servicegraphs[0].ID != "SG-TEST-001" {
		t.Errorf("unexpected ID: %s", list.Servicegraphs[0].ID)
	}
}

func TestGetServicegraph(t *testing.T) {
	dir := setupSGCosmos(t)
	CreateServicegraph(dir, testServicegraph())

	sg, err := GetServicegraph(dir, "SG-TEST-001")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if sg.NodeCount != 2 {
		t.Errorf("node count: got %d, want 2", sg.NodeCount)
	}
	if sg.EdgeCount != 1 {
		t.Errorf("edge count: got %d, want 1", sg.EdgeCount)
	}
}

func TestGetServicegraphNotFound(t *testing.T) {
	dir := setupSGCosmos(t)
	_, err := GetServicegraph(dir, "NONEXISTENT")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDeleteServicegraph(t *testing.T) {
	dir := setupSGCosmos(t)
	CreateServicegraph(dir, testServicegraph())

	if err := DeleteServicegraph(dir, "SG-TEST-001"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	list, _ := ListServicegraphs(dir)
	if list.Count != 0 {
		t.Errorf("expected 0 after delete, got %d", list.Count)
	}
}

func TestCreateServicegraphDuplicate(t *testing.T) {
	dir := setupSGCosmos(t)
	sg := testServicegraph()
	CreateServicegraph(dir, sg)
	err := CreateServicegraph(dir, sg)
	if err == nil {
		t.Error("expected conflict error")
	}
}

func TestGetServicegraphMermaid(t *testing.T) {
	dir := setupSGCosmos(t)
	CreateServicegraph(dir, testServicegraph())

	g, err := GetServicegraphMermaid(dir, "SG-TEST-001")
	if err != nil {
		t.Fatalf("mermaid: %v", err)
	}
	if g.Format != "mermaid" {
		t.Errorf("format: %s", g.Format)
	}
	if g.Content == "" {
		t.Error("empty mermaid content")
	}
}

func TestGetServicegraphExecutionOrder(t *testing.T) {
	dir := setupSGCosmos(t)
	sg := testServicegraph()
	// Add a depends_on edge
	sg.Edges = append(sg.Edges, model.GraphEdge{ID: "E-002", Source: "N-001", Target: "N-002", Type: "depends_on", Binding: "hard"})
	CreateServicegraph(dir, sg)

	exOrder, err := GetServicegraphExecutionOrder(dir, "SG-TEST-001")
	if err != nil {
		t.Fatalf("execution: %v", err)
	}
	if exOrder.ServicegraphID != "SG-TEST-001" {
		t.Errorf("unexpected ID: %s", exOrder.ServicegraphID)
	}
	if len(exOrder.Steps) == 0 {
		t.Error("expected at least one step group")
	}
}
