package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

// createTestCosmosWithInstances sets up a cosmos with one product blueprint,
// one service blueprint, one product instance, and one service instance.
func createTestCosmosWithInstances(t *testing.T) string {
	t.Helper()
	p := createTestCosmosWithCatalog(t)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "instances", "products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "instances", "services"), 0o755))

	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "blueprints", "products", "PB-INST-001.yaml"),
		[]byte("id: PB-INST-001\ntype: product_blueprint\nname: Test Product\nversion: 0.1.0\nstatus: draft\nowner: Team\n"),
		0o644,
	))
	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "blueprints", "services", "SB-INST-001.yaml"),
		[]byte("id: SB-INST-001\ntype: service_blueprint\nname: Test Service\nversion: 0.1.0\nstatus: draft\nowner: Team\n"),
		0o644,
	))
	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "instances", "products", "PI-001.yaml"),
		[]byte("id: PI-001\ntype: product_instance\nname: Product Instance\nblueprint_ref: PB-INST-001\nblueprint_version: 0.1.0\nstatus: draft\nowner: Team\ncompliance_status: unknown\nfindings: []\n"),
		0o644,
	))
	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "instances", "services", "SI-001.yaml"),
		[]byte("id: SI-001\ntype: service_instance\nname: Service Instance\nblueprint_ref: SB-INST-001\nblueprint_version: 0.1.0\nstatus: draft\nowner: Team\nowning_product_instance: PI-001\ncompliance_status: unknown\nfindings: []\n"),
		0o644,
	))
	return p
}

// ── ListInstances ────────────────────────────────────────────────────────────

func TestListInstances_ReturnsAll(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	got, err := ListInstances(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 2 {
		t.Fatalf("expected 2 instances, got %d", got.Count)
	}
}

func TestListInstances_EmptyCosmos(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	got, err := ListInstances(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Count != 0 {
		t.Fatalf("expected 0 instances, got %d", got.Count)
	}
}

// ── GetInstance ──────────────────────────────────────────────────────────────

func TestGetInstance_Success(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	got, err := GetInstance(p, "PI-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "PI-001" || got.Type != "product_instance" {
		t.Fatalf("unexpected instance: %+v", got)
	}
	if got.BlueprintRef != "PB-INST-001" {
		t.Fatalf("expected blueprint_ref PB-INST-001, got %s", got.BlueprintRef)
	}
}

func TestGetInstance_NotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	_, err := GetInstance(p, "DOES-NOT-EXIST")
	if err == nil {
		t.Fatal("expected error for missing instance")
	}
	ae, ok := AsAppError(err)
	if !ok || ae.Code != CodeInstanceNotFound {
		t.Fatalf("expected INSTANCE_NOT_FOUND, got %v", err)
	}
}

// ── GetInstanceCompliance ────────────────────────────────────────────────────

func TestGetInstanceCompliance_Success(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	got, err := GetInstanceCompliance(p, "PI-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.InstanceID != "PI-001" {
		t.Fatalf("unexpected instance ID: %s", got.InstanceID)
	}
	if got.Status != "unknown" {
		t.Fatalf("expected status unknown, got %s", got.Status)
	}
}

func TestGetInstanceCompliance_NotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	_, err := GetInstanceCompliance(p, "NO-SUCH-INSTANCE")
	if err == nil {
		t.Fatal("expected error for missing instance")
	}
}

// ── CreateInstance ───────────────────────────────────────────────────────────

func TestCreateInstance_ProductInstance(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	inst := model.Instance{
		ID:           "PI-NEW-001",
		Type:         "product_instance",
		Name:         "New Product",
		BlueprintRef: "PB-INST-001",
		Status:       "draft",
		Owner:        "Team",
	}
	if err := CreateInstance(p, inst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := GetInstance(p, "PI-NEW-001")
	if err != nil {
		t.Fatalf("could not load created instance: %v", err)
	}
	if got.Type != "product_instance" {
		t.Fatalf("expected product_instance, got %s", got.Type)
	}
}

func TestCreateInstance_ServiceInstance(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	inst := model.Instance{
		ID:           "SI-NEW-001",
		Type:         "service_instance",
		Name:         "New Service",
		BlueprintRef: "SB-INST-001",
		Status:       "draft",
		Owner:        "Team",
	}
	if err := CreateInstance(p, inst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateInstance_DuplicateID(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	inst := model.Instance{ID: "PI-001", Type: "product_instance", BlueprintRef: "PB-INST-001", Status: "draft", Owner: "Team"}
	err := CreateInstance(p, inst)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestCreateInstance_EmptyID_AutoGenerates(t *testing.T) {
	// ADR-0020: leere ID erzeugt eine system-generierte ID (Präfix PRI_ für product_instance).
	p := createTestCosmosWithInstances(t)
	inst := model.Instance{ID: "", Type: "product_instance", BlueprintRef: "PB-INST-001", Status: "draft", Owner: "Team"}
	if err := CreateInstance(p, inst); err != nil {
		t.Fatalf("CreateInstance with empty ID should auto-generate, got error: %v", err)
	}
	list, err := ListInstances(p)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range list.Instances {
		if i.BlueprintRef == "PB-INST-001" && idgen.IsValidForType(i.ID, "product_instance") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected an instance with auto-generated PRI_ ID")
	}
}

func TestCreateInstance_InvalidType(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	inst := model.Instance{ID: "XI-001", Type: "bad_type", BlueprintRef: "PB-INST-001", Status: "draft", Owner: "Team"}
	if err := CreateInstance(p, inst); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestCreateInstance_MissingBlueprintRef(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	inst := model.Instance{ID: "PI-NOREF", Type: "product_instance", Status: "draft", Owner: "Team"}
	if err := CreateInstance(p, inst); err == nil {
		t.Fatal("expected error for missing blueprint_ref")
	}
}

func TestCreateInstance_MissingCosmos(t *testing.T) {
	p := t.TempDir()
	inst := model.Instance{ID: "PI-001", Type: "product_instance", BlueprintRef: "PB-001", Status: "draft", Owner: "Team"}
	if err := CreateInstance(p, inst); err == nil {
		t.Fatal("expected error for missing cosmos")
	}
}

// ── DeleteInstance ───────────────────────────────────────────────────────────

func TestDeleteInstance_Success(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	if err := DeleteInstance(p, "SI-001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := GetInstance(p, "SI-001"); err == nil {
		t.Fatal("expected instance to be gone after delete")
	}
}

func TestDeleteInstance_NotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	if err := DeleteInstance(p, "DOES-NOT-EXIST"); err == nil {
		t.Fatal("expected error for non-existent instance")
	}
}

// ── PatchInstance ────────────────────────────────────────────────────────────

func TestPatchInstance_Fields(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	got, err := PatchInstance(p, "PI-001", map[string]string{
		"name":              "Updated Product",
		"owner":             "New Owner",
		"status":            "active",
		"compliance_status": "compliant",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated Product" {
		t.Fatalf("name not updated: %s", got.Name)
	}
	if got.Owner != "New Owner" {
		t.Fatalf("owner not updated: %s", got.Owner)
	}
	if got.Status != "active" {
		t.Fatalf("status not updated: %s", got.Status)
	}
	if got.ComplianceStatus != "compliant" {
		t.Fatalf("compliance_status not updated: %s", got.ComplianceStatus)
	}
}

func TestPatchInstance_NotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	_, err := PatchInstance(p, "NO-SUCH", map[string]string{"name": "x"})
	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
}

// ── VerifyInstance ───────────────────────────────────────────────────────────

func TestVerifyInstance_SetsCompliantAndAddsEvidence(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	got, err := VerifyInstance(p, "PI-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ComplianceStatus != "compliant" {
		t.Fatalf("expected compliant, got %s", got.ComplianceStatus)
	}
	if len(got.Evidence) == 0 {
		t.Fatal("expected at least one evidence entry after verify")
	}
	// ADR-0020: Evidence-IDs sind system-generiert (EVD_xxxxxx).
	if !idgen.IsValidForType(got.Evidence[0].ID, "evidence") {
		t.Fatalf("unexpected evidence ID: %s", got.Evidence[0].ID)
	}
}

func TestVerifyInstance_NotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	_, err := VerifyInstance(p, "NO-SUCH")
	if err == nil {
		t.Fatal("expected error for non-existent instance")
	}
}

// ── ProvisionServiceInstance ─────────────────────────────────────────────────

func TestProvisionServiceInstance_Success(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	newSvc := model.Instance{
		ID:           "SI-PROVISIONED-001",
		Name:         "Provisioned Service",
		BlueprintRef: "SB-INST-001",
		Status:       "draft",
		Owner:        "Team",
	}
	got, err := ProvisionServiceInstance(p, "PI-001", newSvc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Type != "service_instance" {
		t.Fatalf("expected service_instance, got %s", got.Type)
	}
	if got.OwningProductInstance != "PI-001" {
		t.Fatalf("expected owning_product_instance PI-001, got %s", got.OwningProductInstance)
	}
	// Product should reference the new service
	product, err := GetInstance(p, "PI-001")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, sid := range product.ProvisionedServiceInstances {
		if sid == "SI-PROVISIONED-001" {
			found = true
		}
	}
	if !found {
		t.Fatalf("product instance does not reference provisioned service: %v", product.ProvisionedServiceInstances)
	}
}

func TestProvisionServiceInstance_TargetNotProduct(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	newSvc := model.Instance{ID: "SI-NEW", BlueprintRef: "SB-INST-001", Status: "draft", Owner: "Team"}
	_, err := ProvisionServiceInstance(p, "SI-001", newSvc)
	if err == nil {
		t.Fatal("expected error when target is not a product_instance")
	}
}

func TestProvisionServiceInstance_ProductNotFound(t *testing.T) {
	p := createTestCosmosWithInstances(t)
	newSvc := model.Instance{ID: "SI-NEW", BlueprintRef: "SB-INST-001", Status: "draft", Owner: "Team"}
	_, err := ProvisionServiceInstance(p, "PI-NONEXISTENT", newSvc)
	if err == nil {
		t.Fatal("expected error for non-existent product instance")
	}
}

// ── ListServices ─────────────────────────────────────────────────────────────

func TestListServices_Success(t *testing.T) {
	p := createAppTestCosmos(t)
	got, err := ListServices(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Domain != "identity.blumer.cloud" {
		t.Fatalf("unexpected domain: %s", got.Domain)
	}
	if len(got.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(got.Services))
	}
}

func TestListServices_DomainNotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	_, err := ListServices(p, "does-not-exist.example")
	if err == nil {
		t.Fatal("expected error for non-existent domain")
	}
}
