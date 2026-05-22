package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nomos/nomos/internal/app"
	"github.com/nomos/nomos/internal/storage"
)

func newCosmosForTools(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.ServicesDir(p), "rule-validation-api"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Platform Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.ServicesDir(p), "rule-validation-api/service.yaml"), []byte("name: rule-validation-api\nowner: Platform Team\nstatus: draft\n"), 0o644))
	return p
}

func TestToolCapabilityAdd_WithOperationRefs(t *testing.T) {
	p := newCosmosForTools(t)
	if _, err := app.AddServiceOperation(p, "rule-validation-api", "validate"); err != nil {
		t.Fatal(err)
	}

	add := toolCapabilityAdd(p)
	if _, _, err := add(context.Background(), nil, capabilityAddIn{
		Service:       "rule-validation-api",
		Name:          "Rule validation",
		ID:            "cap-rule-validation",
		Stability:     "stable",
		OperationRefs: []string{"validate"},
	}); err != nil {
		t.Fatal(err)
	}

	svc, err := app.GetService(p, "rule-validation-api")
	if err != nil {
		t.Fatal(err)
	}
	cap := findCapability(svc.CapabilityDefs, "cap-rule-validation", "")
	if cap.ID == "" {
		t.Fatalf("capability not persisted: %+v", svc.CapabilityDefs)
	}
	if !reflect.DeepEqual(cap.OperationRefs, []string{"validate"}) {
		t.Fatalf("expected operation_refs=[validate], got %v", cap.OperationRefs)
	}
}

func TestToolCapabilityUpdate_OmittedOperationRefsArePreserved(t *testing.T) {
	p := newCosmosForTools(t)
	if _, err := app.AddServiceOperation(p, "rule-validation-api", "validate"); err != nil {
		t.Fatal(err)
	}
	add := toolCapabilityAdd(p)
	if _, _, err := add(context.Background(), nil, capabilityAddIn{
		Service:       "rule-validation-api",
		Name:          "Rule validation",
		ID:            "cap-rule-validation",
		OperationRefs: []string{"validate"},
	}); err != nil {
		t.Fatal(err)
	}

	// Update only summary; operation_refs must remain intact.
	newSummary := "Validates incoming rule documents."
	update := toolCapabilityUpdate(p)
	if _, _, err := update(context.Background(), nil, capabilityUpdateIn{
		Service:      "rule-validation-api",
		CapabilityID: "cap-rule-validation",
		Summary:      &newSummary,
	}); err != nil {
		t.Fatal(err)
	}

	svc, _ := app.GetService(p, "rule-validation-api")
	cap := findCapability(svc.CapabilityDefs, "cap-rule-validation", "")
	if cap.Summary != newSummary {
		t.Fatalf("summary not updated: %q", cap.Summary)
	}
	if !reflect.DeepEqual(cap.OperationRefs, []string{"validate"}) {
		t.Fatalf("operation_refs lost on partial update: %v", cap.OperationRefs)
	}
}

func TestToolCapabilityUpdate_EmptyOperationRefsClears(t *testing.T) {
	p := newCosmosForTools(t)
	if _, err := app.AddServiceOperation(p, "rule-validation-api", "validate"); err != nil {
		t.Fatal(err)
	}
	add := toolCapabilityAdd(p)
	if _, _, err := add(context.Background(), nil, capabilityAddIn{
		Service:       "rule-validation-api",
		Name:          "Rule validation",
		ID:            "cap-rule-validation",
		OperationRefs: []string{"validate"},
	}); err != nil {
		t.Fatal(err)
	}

	empty := []string{}
	update := toolCapabilityUpdate(p)
	if _, _, err := update(context.Background(), nil, capabilityUpdateIn{
		Service:       "rule-validation-api",
		CapabilityID:  "cap-rule-validation",
		OperationRefs: &empty,
	}); err != nil {
		t.Fatal(err)
	}

	svc, _ := app.GetService(p, "rule-validation-api")
	cap := findCapability(svc.CapabilityDefs, "cap-rule-validation", "")
	if len(cap.OperationRefs) != 0 {
		t.Fatalf("expected empty operation_refs, got %v", cap.OperationRefs)
	}
}

func TestToolMethodDelete_CleansCapabilityRefs(t *testing.T) {
	p := newCosmosForTools(t)
	if _, err := app.AddServiceOperation(p, "rule-validation-api", "validate"); err != nil {
		t.Fatal(err)
	}
	add := toolCapabilityAdd(p)
	if _, _, err := add(context.Background(), nil, capabilityAddIn{
		Service:       "rule-validation-api",
		Name:          "Rule validation",
		ID:            "cap-rule-validation",
		OperationRefs: []string{"validate"},
	}); err != nil {
		t.Fatal(err)
	}

	del := toolMethodDelete(p)
	if _, _, err := del(context.Background(), nil, methodDeleteIn{
		Service: "rule-validation-api",
		Method:  "validate",
	}); err != nil {
		t.Fatal(err)
	}

	svc, _ := app.GetService(p, "rule-validation-api")
	cap := findCapability(svc.CapabilityDefs, "cap-rule-validation", "")
	if len(cap.OperationRefs) != 0 {
		t.Fatalf("expected operation_refs to be cleaned, got %v", cap.OperationRefs)
	}
}
