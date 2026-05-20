package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nomos/nomos/internal/storage"
)

func makeTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "user-account"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p),
		[]byte("id: cosmos-test\nname: Test Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Test Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "user-account", "service.yaml"),
		[]byte("name: user-account\nowner: Identity Team\nstatus: draft\n"), 0o644))
	return p
}

func callTool[I any](fn func(context.Context, *mcp.CallToolRequest, I) (*mcp.CallToolResult, any, error), in I) (*mcp.CallToolResult, error) {
	res, _, err := fn(context.Background(), &mcp.CallToolRequest{}, in)
	return res, err
}

// ── read tools ────────────────────────────────────────────────────────────────

func TestToolCosmosInfo(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolCosmosInfo(p), emptyIn{})
	if err != nil {
		t.Fatalf("toolCosmosInfo: %v", err)
	}
	if res == nil || len(res.Content) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestToolListServices(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolListServices(p), emptyIn{})
	if err != nil {
		t.Fatalf("toolListServices: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolGetService(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolGetService(p), getServiceIn{Service: "user-account"})
	if err != nil {
		t.Fatalf("toolGetService: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolGetService_NotFound(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolGetService(p), getServiceIn{Service: "ghost-svc"})
	if err == nil {
		t.Error("expected error for unknown service")
	}
}

func TestToolValidate(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolValidate(p), emptyIn{})
	if err != nil {
		t.Fatalf("toolValidate: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolListDecisions(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolListDecisions(p), emptyIn{})
	if err != nil {
		t.Fatalf("toolListDecisions: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolListBlueprints(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolListBlueprints(p), emptyIn{})
	if err != nil {
		t.Fatalf("toolListBlueprints: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

// ── write tools ───────────────────────────────────────────────────────────────

func TestToolServiceAdd(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolServiceAdd(p), serviceAddIn{
		Name:  "new-service",
		Owner: "Team",
	})
	if err != nil {
		t.Fatalf("toolServiceAdd: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolServiceAdd_MissingFields(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolServiceAdd(p), serviceAddIn{Name: ""})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestToolProductCreate(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolProductCreate(p), productCreateIn{
		Name:  "My Product",
		Owner: "Team",
	})
	if err != nil {
		t.Fatalf("toolProductCreate: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolProductCreate_MissingFields(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolProductCreate(p), productCreateIn{Name: ""})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestToolDecisionCreate(t *testing.T) {
	p := makeTestCosmos(t)
	res, err := callTool(toolDecisionCreate(p), decisionCreateIn{
		Name:  "DNS Check",
		Owner: "Team",
	})
	if err != nil {
		t.Fatalf("toolDecisionCreate: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolDecisionCreate_MissingFields(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolDecisionCreate(p), decisionCreateIn{Name: ""})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestToolInstanceCreate(t *testing.T) {
	p := makeTestCosmos(t)
	// Create a blueprint first.
	bpDir := filepath.Join(storage.CatalogDir(p), "blueprints", "products")
	if err := os.MkdirAll(bpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bpDir, "PB-TEST-001.yaml"),
		[]byte("id: PB-TEST-001\ntype: product_blueprint\nname: Test BP\nstatus: draft\nowner: Team\nversion: 0.1.0\nrequired_service_blueprints: []\n"),
		0o644); err != nil {
		t.Fatal(err)
	}

	res, err := callTool(toolInstanceCreate(p), instanceCreateIn{
		BlueprintRef: "PB-TEST-001",
		Type:         "product_instance",
		Name:         "Test Instance",
		Owner:        "Team",
	})
	if err != nil {
		t.Fatalf("toolInstanceCreate: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolInstanceCreate_MissingBlueprintRef(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolInstanceCreate(p), instanceCreateIn{BlueprintRef: ""})
	if err == nil {
		t.Error("expected error for missing blueprint_ref")
	}
}

func TestToolInstanceCreate_InvalidType(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolInstanceCreate(p), instanceCreateIn{
		BlueprintRef: "PB-1",
		Type:         "invalid_type",
	})
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestIsBusinessRuleTaskType(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"business_rule_task", true},
		{"businessruletask", true},
		{"bpmn:BusinessRuleTask", true},
		{"service_task", false},
		{"user_task", false},
		{"", false},
	}
	for _, c := range cases {
		got := isBusinessRuleTaskType(c.in)
		if got != c.want {
			t.Errorf("isBusinessRuleTaskType(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestTextResult(t *testing.T) {
	res, _, err := textResult(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("textResult: %v", err)
	}
	if res == nil || len(res.Content) == 0 {
		t.Error("expected content")
	}
}

func TestNew(t *testing.T) {
	p := makeTestCosmos(t)
	srv := New(p)
	if srv == nil {
		t.Error("expected non-nil server")
	}
}

func TestToolProcessCreate(t *testing.T) {
	p := makeTestCosmos(t)
	// Create product first so process can reference it.
	prodRes, err := callTool(toolProductCreate(p), productCreateIn{
		ID:    "PB-PROC-TEST-001",
		Name:  "Process Product",
		Owner: "Team",
	})
	if err != nil {
		t.Fatalf("toolProductCreate: %v", err)
	}
	if prodRes == nil {
		t.Fatal("expected product result")
	}

	res, err := callTool(toolProcessCreate(p), processCreateIn{
		ProductID: "PB-PROC-TEST-001",
		Name:      "Provision Flow",
		Summary:   "Main provisioning process",
	})
	if err != nil {
		t.Fatalf("toolProcessCreate: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolProcessCreate_MissingFields(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolProcessCreate(p), processCreateIn{ProductID: "", Name: ""})
	if err == nil {
		t.Error("expected error for missing product_id/name")
	}
}

func TestToolProcessStepAdd(t *testing.T) {
	p := makeTestCosmos(t)
	if _, err := callTool(toolProductCreate(p), productCreateIn{
		ID:    "PB-STEP-TEST-001",
		Name:  "Step Product",
		Owner: "Team",
	}); err != nil {
		t.Fatalf("toolProductCreate: %v", err)
	}
	if _, err := callTool(toolProcessCreate(p), processCreateIn{
		ProductID: "PB-STEP-TEST-001",
		Name:      "Step Flow",
	}); err != nil {
		t.Fatalf("toolProcessCreate: %v", err)
	}

	// List processes to find the ID.
	type procDTO struct{ ID string }
	processes, err := listProcesses(p, "PB-STEP-TEST-001")
	if err != nil || len(processes) == 0 {
		t.Skip("cannot list processes to get ID")
	}

	res, err := callTool(toolProcessStepAdd(p), processStepAddIn{
		ProcessID: processes[0],
		Name:      "Validate DNS",
		TaskType:  "service_task",
		Required:  true,
	})
	if err != nil {
		t.Fatalf("toolProcessStepAdd: %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

func TestToolProcessStepAdd_MissingFields(t *testing.T) {
	p := makeTestCosmos(t)
	_, err := callTool(toolProcessStepAdd(p), processStepAddIn{ProcessID: "", Name: ""})
	if err == nil {
		t.Error("expected error for missing process_id/name")
	}
}

func TestToolProcessStepAdd_BusinessRuleWithDecision(t *testing.T) {
	p := makeTestCosmos(t)
	if _, err := callTool(toolProductCreate(p), productCreateIn{
		ID:    "PB-BRT-TEST-001",
		Name:  "BRT Product",
		Owner: "Team",
	}); err != nil {
		t.Fatalf("toolProductCreate: %v", err)
	}
	if _, err := callTool(toolProcessCreate(p), processCreateIn{
		ProductID: "PB-BRT-TEST-001",
		Name:      "BRT Flow",
	}); err != nil {
		t.Fatalf("toolProcessCreate: %v", err)
	}
	processes, err := listProcesses(p, "PB-BRT-TEST-001")
	if err != nil || len(processes) == 0 {
		t.Skip("cannot list processes")
	}
	// business_rule_task + decision_ref triggers auto-gateway
	res, err := callTool(toolProcessStepAdd(p), processStepAddIn{
		ProcessID:   processes[0],
		Name:        "Check DNS",
		TaskType:    "business_rule_task",
		DecisionRef: "dec-dns-format-check",
		GatewayName: "DNS valid?",
	})
	if err != nil {
		t.Fatalf("toolProcessStepAdd (brt): %v", err)
	}
	if res == nil {
		t.Error("expected result")
	}
}

// listProcesses scans YAML files in the processes dir and returns process IDs.
func listProcesses(cosmosPath, _ string) ([]string, error) {
	dir := filepath.Join(storage.CatalogDir(cosmosPath), "blueprints", "processes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		// Extract id: field from YAML manually.
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "id:") {
				id := strings.TrimSpace(strings.TrimPrefix(line, "id:"))
				ids = append(ids, id)
				break
			}
		}
	}
	return ids, nil
}
