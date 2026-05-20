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
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-PROC-001", Name: "Process Product", Summary: "User-facing summary"})
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
	if proc.BPMN.File != "prc-proc-001.bpmn" {
		t.Fatalf("unexpected process dto: %+v", proc)
	}
	// New processes start with a blank BPMN (no placeholder tasks).
	// Add a real task via BPMN update before mapping.
	customBPMN := DefaultBPMNTemplate(proc.BPMN.ProcessID, product.Name)
	customBPMN = strings.ReplaceAll(customBPMN, "<bpmn:sequenceFlow id=\"Flow_Start_End\"", "<bpmn:task id=\"Task_Custom\" name=\"Custom task\"><bpmn:incoming>Flow_Start_Custom</bpmn:incoming><bpmn:outgoing>Flow_Custom_End</bpmn:outgoing></bpmn:task><bpmn:sequenceFlow id=\"Flow_Start_Custom\" sourceRef=\"StartEvent_Begin\" targetRef=\"Task_Custom\" /><bpmn:sequenceFlow id=\"Flow_Custom_End\" sourceRef=\"Task_Custom\" targetRef=\"EndEvent_Done\" /><bpmn:sequenceFlow id=\"Flow_Start_End\"")
	if _, err := UpdateProcessBPMN(p, proc.ID, customBPMN); err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateProcessTaskMappings(p, proc.ID, UpdateTaskMappingsRequest{TaskMappings: []ProcessTaskMappingDTO{{BPMNElementID: "Task_Custom", ServiceRef: "identity.blumer.cloud/user-account", Role: "primary", Required: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.TaskMappings) != 1 {
		t.Fatalf("mapping not saved: %+v", updated.TaskMappings)
	}
	if updated.Validation.Status != "needs_attention" {
		t.Fatalf("expected warnings for missing SLA, got %s", updated.Validation.Status)
	}
	stored, err := os.ReadFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", product.ID+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(stored), "PRC-PROC-001") {
		t.Fatalf("product did not reference process:\n%s", stored)
	}
}

func TestNewProcessStartsWithBlankBPMN(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-PROC-STD", Name: "Standard Process Product"})
	if err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-PROC-STD", Name: "Standard Process"})
	if err != nil {
		t.Fatal(err)
	}
	// New processes have a blank BPMN with no mappable tasks.
	if len(proc.Tasks) != 0 {
		t.Fatalf("expected no tasks in blank BPMN, got %d: %+v", len(proc.Tasks), proc.Tasks)
	}
	// No BPMN_TASK_UNMAPPED warnings since there are no tasks to map.
	for _, f := range proc.Validation.Findings {
		if f.Code == "BPMN_TASK_UNMAPPED" {
			t.Fatalf("blank BPMN should have no unmapped-task warnings: %+v", proc.Validation.Findings)
		}
	}
}

func TestProcessInvalidBPMNAndUnknownMapping(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-PROC-002", Name: "Old Product"})
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

func TestUpdateProcessBPMNMirrorsLaneServiceBindings(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-LANE-001", Name: "Lane Product"})
	if err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-LANE-001", Name: "Lane Process"})
	if err != nil {
		t.Fatal(err)
	}
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" id="Defs">
  <bpmn:process id="Process_Lane" isExecutable="false">
    <bpmn:laneSet id="LaneSet_1">
      <bpmn:lane id="Lane_Identity" name="Identity">
        <bpmn:documentation>nomos-service-ref:identity.blumer.cloud/user-account</bpmn:documentation>
      </bpmn:lane>
      <bpmn:lane id="Lane_Unbound" name="Other"/>
    </bpmn:laneSet>
    <bpmn:startEvent id="Start"/>
    <bpmn:endEvent id="End"/>
  </bpmn:process>
</bpmn:definitions>`
	updated, err := UpdateProcessBPMN(p, proc.ID, xml)
	if err != nil {
		t.Fatalf("update bpmn: %v", err)
	}
	if len(updated.Lanes) != 2 {
		t.Fatalf("expected 2 lanes mirrored, got %d: %+v", len(updated.Lanes), updated.Lanes)
	}
	byID := map[string]ProcessLaneDTO{}
	for _, l := range updated.Lanes {
		byID[l.BPMNLaneID] = l
	}
	if got := byID["Lane_Identity"]; got.ServiceRef != "identity.blumer.cloud/user-account" || got.Name != "Identity" {
		t.Fatalf("Lane_Identity mirror wrong: %+v", got)
	}
	if got := byID["Lane_Unbound"]; got.ServiceRef != "" || got.Name != "Other" {
		t.Fatalf("Lane_Unbound mirror wrong: %+v", got)
	}
}

func TestBackwardCompatibleProductWithoutProcesses(t *testing.T) {
	p := createAppTestCosmos(t)
	dir := filepath.Join(storage.CatalogDir(p), "blueprints", "products")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	bp := model.Blueprint{ID: "PROD-OLD-001", Type: "product_blueprint", Name: "Legacy Product", Version: "0.1.0", Status: "draft", Owner: "Team", Summary: "Legacy"}
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
