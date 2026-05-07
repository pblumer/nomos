package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/model"
)

func setupSGServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cosmos.yaml"), []byte("id: test\ntype: cosmos\nname: Test\nversion: 0.1.0\nstatus: active\nowner: test\nsummary: test\ndomains: []\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "domains"), 0755)
	return NewHandler(dir), dir
}

func TestAPIServicegraphsCRUD(t *testing.T) {
	h, dir := setupSGServer(t)

	// Create
	sg := model.Servicegraph{
		ID: "SG-API-001", Type: "servicegraph", Name: "API Test",
		Version: "0.1.0", Status: "draft", Owner: "test",
		Summary: "test", RelatedProduct: "PROD-001",
		Nodes: []model.GraphNode{
			{ID: "N-001", Type: "service", Name: "Svc", Mandatory: true},
			{ID: "N-002", Type: "activity", Name: "Act", Mandatory: true},
		},
		Edges: []model.GraphEdge{
			{ID: "E-001", Source: "N-001", Target: "N-002", Type: "composed_of", Binding: "hard"},
			{ID: "E-002", Source: "N-001", Target: "N-002", Type: "depends_on", Binding: "hard"},
		},
	}
	body, _ := json.Marshal(sg)
	req := httptest.NewRequest("POST", "/api/v1/servicegraphs", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d, body: %s", w.Code, w.Body.String())
	}

	// List
	req = httptest.NewRequest("GET", "/api/v1/servicegraphs", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("list: status %d", w.Code)
	}
	var listResp app.ServicegraphsDTO
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if listResp.Count != 1 {
		t.Errorf("list count: %d", listResp.Count)
	}

	// Get
	req = httptest.NewRequest("GET", "/api/v1/servicegraphs/SG-API-001", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("get: status %d", w.Code)
	}

	// Mermaid
	req = httptest.NewRequest("GET", "/api/v1/servicegraphs/SG-API-001/mermaid", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("mermaid: status %d, body: %s", w.Code, w.Body.String())
	}

	// Execution
	req = httptest.NewRequest("GET", "/api/v1/servicegraphs/SG-API-001/execution", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("execution: status %d", w.Code)
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/api/v1/servicegraphs/SG-API-001", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("delete: status %d", w.Code)
	}

	// Verify deleted
	req = httptest.NewRequest("GET", "/api/v1/servicegraphs/SG-API-001", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}

	_ = dir
}
