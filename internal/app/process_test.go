package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

func TestProcessArtifactCreateTasksMappingsAndValidation(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{ID: "PROD-PROC-001", Name: "Process Product", Summary: "User-facing summary"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddProductFulfillmentService(p, product.ID, AddFulfillmentServiceRequest{ServiceRef: "identity.blumer.cloud/user-account", Role: "primary", Required: true}); err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-PROC-001", Name: "Provision Process"})
	if err != nil {
		t.Fatal(err)
	}
	if proc.BPMN.File != "prc-proc-001.bpmn" || len(proc.Tasks) == 0 {
		t.Fatalf("unexpected process dto: %+v", proc)
	}
	updated, err := UpdateProcessTaskMappings(p, proc.ID, UpdateTaskMappingsRequest{TaskMappings: []ProcessTaskMappingDTO{{BPMNElementID: "Task_PerformFulfillment", ServiceRef: "identity.blumer.cloud/user-account", Role: "primary", Required: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.TaskMappings) != 1 {
		t.Fatalf("mapping not saved: %+v", updated.TaskMappings)
	}
	if updated.Validation.Status != "needs_attention" {
		t.Fatalf("expected warnings for unmapped tasks/missing SLA, got %s", updated.Validation.Status)
	}
	stored, err := os.ReadFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", product.ID+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stored), "PRC-PROC-001") {
		t.Fatalf("product did not reference process:\n%s", stored)
	}
}

func TestProcessInvalidBPMNAndUnknownMapping(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, "identity.blumer.cloud", CreateProductOfferingRequest{ID: "PROD-PROC-002", Name: "Old Product"})
	if err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-PROC-002", Name: "Broken Process"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateProcessBPMN(p, proc.ID, "<broken"); err == nil {
		t.Fatal("expected invalid BPMN error")
	}
	if _, err := UpdateProcessTaskMappings(p, proc.ID, UpdateTaskMappingsRequest{TaskMappings: []ProcessTaskMappingDTO{{BPMNElementID: "Task_DoesNotExist", ServiceRef: "missing.example/service", Role: "primary", Required: true}}}); err != nil {
		t.Fatal(err)
	}
	validated, err := GetProcess(p, proc.ID)
	if err != nil {
		t.Fatal(err)
	}
	foundUnknown := false
	for _, f := range validated.Validation.Findings {
		if f.Code == "TASK_MAPPING_UNKNOWN_TASK" {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Fatalf("expected unknown-task finding: %+v", validated.Validation.Findings)
	}
}

func TestBackwardCompatibleProductWithoutProcesses(t *testing.T) {
	p := createAppTestCosmos(t)
	dir := filepath.Join(storage.CatalogDir(p), "blueprints", "products")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	bp := model.Blueprint{ID: "PROD-OLD-001", Type: "product_blueprint", Name: "Legacy Product", Version: "0.1.0", Status: "draft", Owner: "Team", OfferedBy: "identity.blumer.cloud", Summary: "Legacy"}
	if err := fsx.WriteYAML(filepath.Join(dir, "legacy.yaml"), bp); err != nil {
		t.Fatal(err)
	}
	got, err := GetBlueprint(p, bp.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ProcessSummary.Count != 0 || len(got.Processes) != 0 {
		t.Fatalf("legacy product should have no processes: %+v", got.ProcessSummary)
	}
}
