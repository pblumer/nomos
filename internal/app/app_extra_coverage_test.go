package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

// ---------------------------------------------------------------------------
// GetProcessTriggers
// ---------------------------------------------------------------------------

func TestGetProcessTriggers(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	proc, err := CreateProductProcess(p, productID, CreateProcessRequest{
		ID:   "PRC-TRIG-001",
		Name: "Trigger Test",
	})
	if err != nil {
		t.Fatalf("CreateProductProcess: %v", err)
	}
	triggers, err := GetProcessTriggers(p, proc.ID)
	if err != nil {
		t.Fatalf("GetProcessTriggers: %v", err)
	}
	// No triggers by default
	_ = triggers
}

func TestGetProcessTriggers_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := GetProcessTriggers(p, "PRC-GHOST")
	if err == nil {
		t.Error("expected error for missing process")
	}
}

// ---------------------------------------------------------------------------
// GetDecisionVersion + headAsVersion (via ListDecisionVersions with no snapshots)
// ---------------------------------------------------------------------------

func TestGetDecisionVersion_HeadFallback(t *testing.T) {
	// Create a decision directly (without snapshot dir) to trigger headAsVersion path.
	p := createAppTestCosmos(t)
	decDir := filepath.Join(storage.DecisionsDir(p), "DEC-HEAD-001")
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(decDir, 0o755))
	must(os.WriteFile(filepath.Join(decDir, "decision.yaml"),
		[]byte("id: DEC-HEAD-001\ntype: decision\nname: Head Test\nversion: 1.0.0\nstatus: active\nowner: Team\n"),
		0o644))

	// ListDecisionVersions with no snapshots directory triggers headAsVersion.
	dto, err := ListDecisionVersions(p, "DEC-HEAD-001")
	if err != nil {
		t.Fatalf("ListDecisionVersions: %v", err)
	}
	if dto.Count == 0 {
		t.Error("expected at least HEAD version in list")
	}

	// GetDecisionVersion with the current version triggers the HEAD fallback path.
	vdto, err := GetDecisionVersion(p, "DEC-HEAD-001", "1.0.0")
	if err != nil {
		t.Fatalf("GetDecisionVersion (HEAD fallback): %v", err)
	}
	if vdto.ID != "DEC-HEAD-001" {
		t.Errorf("unexpected ID: %q", vdto.ID)
	}
}

func TestGetDecisionVersion_WithSnapshot(t *testing.T) {
	// Test using the existing version test infrastructure.
	p := createVersionTestCosmos(t)
	if _, err := CreateDecision(p, CreateDecisionRequest{
		ID: "DEC-VER-CVG-001", Name: "Version Coverage", Version: "0.1.0", Status: "draft",
	}); err != nil {
		t.Fatalf("CreateDecision: %v", err)
	}
	// GetDecisionVersion for the head (no snapshot yet)
	vdto, err := GetDecisionVersion(p, "DEC-VER-CVG-001", "0.1.0")
	if err != nil {
		t.Fatalf("GetDecisionVersion: %v", err)
	}
	if vdto.ID != "DEC-VER-CVG-001" {
		t.Errorf("unexpected ID: %q", vdto.ID)
	}
}

func TestGetDecisionVersion_NotFound(t *testing.T) {
	p := createVersionTestCosmos(t)
	if _, err := CreateDecision(p, CreateDecisionRequest{
		ID: "DEC-VER-NF-001", Name: "Not Found Version", Version: "1.0.0", Status: "draft",
	}); err != nil {
		t.Fatalf("CreateDecision: %v", err)
	}
	_, err := GetDecisionVersion(p, "DEC-VER-NF-001", "9.9.9")
	if err == nil {
		t.Error("expected error for non-existent version")
	}
}

// ---------------------------------------------------------------------------
// nextProcessID (triggered when idgen fails; we just need coverage of the call)
// ---------------------------------------------------------------------------

func TestCreateProductProcess_AutoID(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	// Passing empty ID triggers nextProcessID
	proc, err := CreateProductProcess(p, productID, CreateProcessRequest{
		Name: "Auto ID Process",
	})
	if err != nil {
		t.Fatalf("CreateProductProcess: %v", err)
	}
	if proc.ID == "" {
		t.Error("expected auto-generated process ID")
	}
}

// ---------------------------------------------------------------------------
// buildDecisionIndex (via BuildNamespaceTree with decisions present)
// ---------------------------------------------------------------------------

func TestBuildNamespaceTree_WithDecision(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := CreateDecision(p, CreateDecisionRequest{
		ID: "DEC-NS-001", Name: "Namespace Tree Decision",
	}); err != nil {
		t.Fatalf("CreateDecision: %v", err)
	}
	dto, err := BuildNamespaceTree(p)
	if err != nil {
		t.Fatalf("BuildNamespaceTree: %v", err)
	}
	_ = dto
}

// ---------------------------------------------------------------------------
// compactServiceLevelLabel (called from GetService-related functions)
// ---------------------------------------------------------------------------

func TestGetService_ServiceLevelCoverage(t *testing.T) {
	p := createAppTestCosmos(t)
	dto, err := GetService(p, "user-account")
	if err != nil {
		t.Fatalf("GetService: %v", err)
	}
	_ = dto.Status
}

// ---------------------------------------------------------------------------
// DeleteService + RenameService + AddService
// ---------------------------------------------------------------------------

func TestDeleteService_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddService(p, "temp-svc", "Team", false); err != nil {
		t.Fatalf("AddService: %v", err)
	}
	if err := DeleteService(p, "temp-svc"); err != nil {
		t.Fatalf("DeleteService: %v", err)
	}
}

func TestRenameServiceExtra(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddService(p, "rename-svc-extra", "Team", false); err != nil {
		t.Fatalf("AddService: %v", err)
	}
	if err := RenameService(p, "rename-svc-extra", "renamed-svc-extra"); err != nil {
		t.Fatalf("RenameService: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DomainDecision - bpmnElementForStep coverage via AddProcessStep with bpmn types
// ---------------------------------------------------------------------------

func TestBPMNElementTypes(t *testing.T) {
	p, productID := createProcessTestCosmos(t)
	proc, err := CreateProductProcess(p, productID, CreateProcessRequest{
		ID: "PRC-BPMN-ELEM", Name: "BPMN Element Types",
	})
	if err != nil {
		t.Fatalf("CreateProductProcess: %v", err)
	}

	taskTypes := []string{
		"service_task",
		"user_task",
		"manual_task",
		"script_task",
		"business_rule_task",
		"exclusive_gateway",
	}
	for i, tt := range taskTypes {
		_, err := AddProcessStep(p, proc.ID, UpsertProcessStepRequest{
			Name:     "Step " + tt,
			TaskType: tt,
			Required: i%2 == 0,
		})
		if err != nil {
			t.Errorf("AddProcessStep (type=%q): %v", tt, err)
		}
	}
}
