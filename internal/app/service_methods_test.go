package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

// ---------------------------------------------------------------------------
// AddServiceMethod / RemoveServiceMethod
// ---------------------------------------------------------------------------

func TestAddServiceMethod_Success(t *testing.T) {
	p := createAppTestCosmos(t)

	svc, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "create")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Methods) != 1 || svc.Methods[0] != "create" {
		t.Fatalf("expected method 'create', got %v", svc.Methods)
	}
}

func TestAddServiceMethod_MultipleMethodsAccumulate(t *testing.T) {
	p := createAppTestCosmos(t)

	for _, m := range []string{"create", "delete", "update"} {
		if _, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", m); err != nil {
			t.Fatalf("AddServiceMethod(%s): %v", m, err)
		}
	}
	svc, err := GetService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Methods) != 3 {
		t.Fatalf("expected 3 methods, got %v", svc.Methods)
	}
}

func TestAddServiceMethod_Duplicate(t *testing.T) {
	p := createAppTestCosmos(t)

	if _, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "create"); err != nil {
		t.Fatal(err)
	}
	_, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "create")
	if err == nil {
		t.Fatal("expected conflict error for duplicate method")
	}
}

func TestAddServiceMethod_EmptyName(t *testing.T) {
	p := createAppTestCosmos(t)

	_, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "  ")
	if err == nil {
		t.Fatal("expected error for empty method name")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddServiceMethod_TrimmedWhitespace(t *testing.T) {
	p := createAppTestCosmos(t)

	svc, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "  create  ")
	if err != nil {
		t.Fatal(err)
	}
	if svc.Methods[0] != "create" {
		t.Fatalf("expected trimmed method name, got %q", svc.Methods[0])
	}
}

func TestAddServiceMethod_ServiceNotFound(t *testing.T) {
	p := createAppTestCosmos(t)

	_, err := AddServiceMethod(p, "identity.blumer.cloud", "does-not-exist", "create")
	if err == nil {
		t.Fatal("expected error for missing service")
	}
}

func TestRemoveServiceMethod_Success(t *testing.T) {
	p := createAppTestCosmos(t)

	if _, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "create"); err != nil {
		t.Fatal(err)
	}

	svc, err := RemoveServiceMethod(p, "identity.blumer.cloud", "user-account", "create")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Methods) != 0 {
		t.Fatalf("expected 0 methods after remove, got %v", svc.Methods)
	}
}

func TestRemoveServiceMethod_OnlyRemovesTarget(t *testing.T) {
	p := createAppTestCosmos(t)

	for _, m := range []string{"create", "delete"} {
		if _, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", m); err != nil {
			t.Fatal(err)
		}
	}
	svc, err := RemoveServiceMethod(p, "identity.blumer.cloud", "user-account", "delete")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Methods) != 1 || svc.Methods[0] != "create" {
		t.Fatalf("expected only 'create' remaining, got %v", svc.Methods)
	}
}

func TestRemoveServiceMethod_NonExistentIsNoOp(t *testing.T) {
	p := createAppTestCosmos(t)

	if _, err := AddServiceMethod(p, "identity.blumer.cloud", "user-account", "create"); err != nil {
		t.Fatal(err)
	}
	svc, err := RemoveServiceMethod(p, "identity.blumer.cloud", "user-account", "does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.Methods) != 1 {
		t.Fatalf("existing method should be untouched, got %v", svc.Methods)
	}
}

// ---------------------------------------------------------------------------
// productSummaryDTO — process grouping
// ---------------------------------------------------------------------------

func TestProductSummaryDTO_ProcessGroupsWithSteps(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "blumer.cloud"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "domain.yaml"), []byte("name: blumer.cloud\nowner: Cloud Team\nstatus: draft\n"), 0o644))

	prod, err := CreateProductOffering(p, "blumer.cloud", CreateProductOfferingRequest{
		ID: "PROD-GRP-001", Name: "Group Test Product", Summary: "test",
	})
	must(err)

	proc1, err := CreateProductProcess(p, prod.ID, CreateProcessRequest{
		ID: "PRC-GRP-001", Name: "Onboarding",
	})
	must(err)
	proc2, err := CreateProductProcess(p, prod.ID, CreateProcessRequest{
		ID: "PRC-GRP-002", Name: "Offboarding",
	})
	must(err)

	_, err = AddProcessStep(p, proc1.ID, UpsertProcessStepRequest{
		Name:       "Create Account",
		ServiceRef: "identity.blumer.cloud/user-account",
		Method:     "create",
		Inputs:     []StepInputBindingDTO{{Name: "username", Source: "instance.username", Required: true}},
		Outputs:    []StepOutputSchemaDTO{{Name: "accountID", Type: "string"}},
	})
	must(err)

	_, err = AddProcessStep(p, proc2.ID, UpsertProcessStepRequest{
		Name:       "Delete Account",
		ServiceRef: "identity.blumer.cloud/user-account",
		Method:     "delete",
	})
	must(err)

	summaries, err := ProductsOfferedBy(p, "blumer.cloud")
	must(err)

	var summary *ProductSummaryDTO
	for i := range summaries {
		if summaries[i].ID == "PROD-GRP-001" {
			summary = &summaries[i]
			break
		}
	}
	if summary == nil {
		t.Fatal("product not found in summaries")
	}
	if len(summary.Processes) != 2 {
		t.Fatalf("expected 2 process groups, got %d: %+v", len(summary.Processes), summary.Processes)
	}

	byID := map[string]ProcessGroupDTO{}
	for _, g := range summary.Processes {
		byID[g.ProcessID] = g
	}

	g1, ok := byID["PRC-GRP-001"]
	if !ok {
		t.Fatal("PRC-GRP-001 missing from process groups")
	}
	if g1.ProcessName != "Onboarding" {
		t.Fatalf("unexpected name: %s", g1.ProcessName)
	}
	if len(g1.Steps) != 1 {
		t.Fatalf("expected 1 step in PRC-GRP-001, got %d", len(g1.Steps))
	}
	step := g1.Steps[0]
	if step.Name != "Create Account" || step.Method != "create" {
		t.Fatalf("unexpected step: %+v", step)
	}
	if len(step.Inputs) != 1 || step.Inputs[0].Name != "username" || !step.Inputs[0].Required {
		t.Fatalf("inputs not propagated to summary: %+v", step.Inputs)
	}
	if len(step.Outputs) != 1 || step.Outputs[0].Name != "accountID" {
		t.Fatalf("outputs not propagated to summary: %+v", step.Outputs)
	}

	g2, ok := byID["PRC-GRP-002"]
	if !ok {
		t.Fatal("PRC-GRP-002 missing from process groups")
	}
	if g2.ProcessName != "Offboarding" || len(g2.Steps) != 1 {
		t.Fatalf("unexpected PRC-GRP-002: %+v", g2)
	}
}

func TestProductSummaryDTO_NoProcessGroupsWhenNoSteps(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "blumer.cloud"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "domain.yaml"), []byte("name: blumer.cloud\nowner: Cloud Team\nstatus: draft\n"), 0o644))

	prod, err := CreateProductOffering(p, "blumer.cloud", CreateProductOfferingRequest{
		ID: "PROD-GRP-002", Name: "No Steps Product", Summary: "test",
	})
	must(err)

	// Process exists but has no steps — should not appear in Processes groups
	_, err = CreateProductProcess(p, prod.ID, CreateProcessRequest{
		ID: "PRC-GRP-003", Name: "Empty Process",
	})
	must(err)

	summaries, err := ProductsOfferedBy(p, "blumer.cloud")
	must(err)

	var summary *ProductSummaryDTO
	for i := range summaries {
		if summaries[i].ID == "PROD-GRP-002" {
			summary = &summaries[i]
			break
		}
	}
	if summary == nil {
		t.Fatal("product not found")
	}
	// Process without steps still appears as a group (so it shows up in the tree)
	if len(summary.Processes) != 1 {
		t.Fatalf("expected 1 process group even without steps, got %d", len(summary.Processes))
	}
	if len(summary.Processes[0].Steps) != 0 {
		t.Fatalf("expected 0 steps in the group, got %d", len(summary.Processes[0].Steps))
	}
	if summary.ProcessCount != 1 {
		t.Fatalf("expected ProcessCount=1, got %d", summary.ProcessCount)
	}
}
