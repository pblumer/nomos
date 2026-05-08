package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

func createTestCosmosWithCatalog(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "services"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Platform Team\n"), 0o644))
	return p
}

func TestCreateBlueprint_ProductBlueprint(t *testing.T) {
	p := createTestCosmosWithCatalog(t)

	bp := model.Blueprint{
		ID:                        "PB-TEST-001",
		Type:                      "product_blueprint",
		Name:                      "Test Product Blueprint",
		Version:                   "0.1.0",
		Status:                    "draft",
		Owner:                     "Product Team",
		Summary:                   "A test product blueprint for the test.",
		RequiredServiceBlueprints: []string{"SB-ACC-001"},
		RequiredServices: []model.RequiredServiceRef{
			{ServiceRef: "identity.blumer.cloud/user-account", ServiceBlueprintRef: "SB-ACC-001", Purpose: "Account creation", Required: true},
		},
	}

	err := CreateBlueprint(p, bp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was written
	path := filepath.Join(storage.CatalogDir(p), "blueprints", "products", "PB-TEST-001.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected blueprint file to exist at %s", path)
	}

	// Verify loading via existing API
	loaded, err := GetBlueprint(p, "PB-TEST-001")
	if err != nil {
		t.Fatalf("could not load created blueprint: %v", err)
	}
	if loaded.ID != "PB-TEST-001" {
		t.Fatalf("expected ID PB-TEST-001, got %s", loaded.ID)
	}
	if loaded.Name != "Test Product Blueprint" {
		t.Fatalf("expected name mismatch")
	}
	if len(loaded.RequiredServices) != 1 {
		t.Fatalf("expected 1 required service, got %d", len(loaded.RequiredServices))
	}
}

func TestCreateBlueprint_ServiceBlueprint(t *testing.T) {
	p := createTestCosmosWithCatalog(t)

	bp := model.Blueprint{
		ID:                  "SB-TEST-001",
		Type:                "service_blueprint",
		Name:                "Test Service Blueprint",
		Version:             "0.1.0",
		Status:              "draft",
		Owner:               "Platform Team",
		Summary:             "A test service blueprint.",
		NamespaceServiceRef: "identity.blumer.cloud/user-account",
	}

	err := CreateBlueprint(p, bp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(storage.CatalogDir(p), "blueprints", "services", "SB-TEST-001.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected blueprint file to exist at %s", path)
	}

	loaded, err := GetBlueprint(p, "SB-TEST-001")
	if err != nil {
		t.Fatalf("could not load created blueprint: %v", err)
	}
	if loaded.Type != "service_blueprint" {
		t.Fatalf("expected service_blueprint, got %s", loaded.Type)
	}
}

func TestCreateBlueprint_DuplicateID(t *testing.T) {
	p := createTestCosmosWithCatalog(t)

	bp := model.Blueprint{ID: "PB-DUP-001", Type: "product_blueprint", Name: "First", Version: "0.1.0", Status: "draft", Owner: "Team"}
	if err := CreateBlueprint(p, bp); err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	bp2 := model.Blueprint{ID: "PB-DUP-001", Type: "product_blueprint", Name: "Second", Version: "0.1.0", Status: "draft", Owner: "Team"}
	err := CreateBlueprint(p, bp2)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestCreateBlueprint_MissingCosmos(t *testing.T) {
	p := t.TempDir()
	bp := model.Blueprint{ID: "PB-TEST", Type: "product_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team"}
	err := CreateBlueprint(p, bp)
	if err == nil {
		t.Fatal("expected error for missing cosmos")
	}
}

func TestCreateBlueprint_EmptyID(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	bp := model.Blueprint{ID: "", Type: "product_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team"}
	err := CreateBlueprint(p, bp)
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestCreateBlueprint_InvalidType(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	bp := model.Blueprint{ID: "PB-BAD", Type: "invalid_type", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team"}
	err := CreateBlueprint(p, bp)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestDeleteBlueprint_Success(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	bp := model.Blueprint{ID: "PB-DEL-001", Type: "product_blueprint", Name: "To Delete", Version: "0.1.0", Status: "draft", Owner: "Team"}
	if err := CreateBlueprint(p, bp); err != nil {
		t.Fatal(err)
	}

	err := DeleteBlueprint(p, "PB-DEL-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(storage.CatalogDir(p), "blueprints", "products", "PB-DEL-001.yaml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted")
	}

	_, err = GetBlueprint(p, "PB-DEL-001")
	if err == nil {
		t.Fatal("expected blueprint not found after delete")
	}
}

func TestDeleteBlueprint_NotFound(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	err := DeleteBlueprint(p, "PB-NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
	if ae, ok := AsAppError(err); !ok || ae.Code != CodeBlueprintNotFound {
		t.Fatalf("expected BLUEPRINT_NOT_FOUND, got %v", err)
	}
}

// helpers shared by service-management and requirement tests

func createProductAndService(t *testing.T) (cosmosPath, productID, serviceID string) {
	t.Helper()
	p := createTestCosmosWithCatalog(t)
	prod := model.Blueprint{ID: "PB-PROD-001", Type: "product_blueprint", Name: "Product", Version: "0.1.0", Status: "draft", Owner: "Team"}
	svc := model.Blueprint{ID: "SB-SVC-001", Type: "service_blueprint", Name: "Service", Version: "0.1.0", Status: "draft", Owner: "Team"}
	if err := CreateBlueprint(p, prod); err != nil {
		t.Fatal(err)
	}
	if err := CreateBlueprint(p, svc); err != nil {
		t.Fatal(err)
	}
	return p, prod.ID, svc.ID
}

// ── AddServiceBlueprintToProduct ────────────────────────────────────────────

func TestAddServiceBlueprintToProduct_Success(t *testing.T) {
	p, productID, serviceID := createProductAndService(t)

	got, err := AddServiceBlueprintToProduct(p, productID, serviceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.RequiredServiceBlueprints) != 1 || got.RequiredServiceBlueprints[0] != serviceID {
		t.Fatalf("expected %s in required_service_blueprints, got %v", serviceID, got.RequiredServiceBlueprints)
	}
}

func TestAddServiceBlueprintToProduct_Idempotent(t *testing.T) {
	p, productID, serviceID := createProductAndService(t)

	if _, err := AddServiceBlueprintToProduct(p, productID, serviceID); err != nil {
		t.Fatal(err)
	}
	got, err := AddServiceBlueprintToProduct(p, productID, serviceID)
	if err != nil {
		t.Fatalf("second add should be idempotent, got error: %v", err)
	}
	if len(got.RequiredServiceBlueprints) != 1 {
		t.Fatalf("expected exactly 1 entry after idempotent add, got %v", got.RequiredServiceBlueprints)
	}
}

func TestAddServiceBlueprintToProduct_TargetNotProduct(t *testing.T) {
	p, _, serviceID := createProductAndService(t)

	_, err := AddServiceBlueprintToProduct(p, serviceID, serviceID)
	if err == nil {
		t.Fatal("expected error when target is not a product_blueprint")
	}
}

func TestAddServiceBlueprintToProduct_ServiceNotFound(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	_, err := AddServiceBlueprintToProduct(p, productID, "SB-NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent service blueprint")
	}
}

func TestAddServiceBlueprintToProduct_ServiceIsNotServiceBlueprint(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	_, err := AddServiceBlueprintToProduct(p, productID, productID)
	if err == nil {
		t.Fatal("expected error when referenced blueprint is not a service_blueprint")
	}
}

// ── RemoveServiceBlueprintFromProduct ───────────────────────────────────────

func TestRemoveServiceBlueprintFromProduct_Success(t *testing.T) {
	p, productID, serviceID := createProductAndService(t)

	if _, err := AddServiceBlueprintToProduct(p, productID, serviceID); err != nil {
		t.Fatal(err)
	}
	got, err := RemoveServiceBlueprintFromProduct(p, productID, serviceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.RequiredServiceBlueprints) != 0 {
		t.Fatalf("expected empty required_service_blueprints after remove, got %v", got.RequiredServiceBlueprints)
	}
}

func TestRemoveServiceBlueprintFromProduct_NonExistentServiceIsNoop(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := RemoveServiceBlueprintFromProduct(p, productID, "SB-DOES-NOT-EXIST")
	if err != nil {
		t.Fatalf("removing non-existent service should be a noop, got error: %v", err)
	}
	if len(got.RequiredServiceBlueprints) != 0 {
		t.Fatalf("expected empty list, got %v", got.RequiredServiceBlueprints)
	}
}

func TestRemoveServiceBlueprintFromProduct_TargetNotProduct(t *testing.T) {
	p, _, serviceID := createProductAndService(t)

	_, err := RemoveServiceBlueprintFromProduct(p, serviceID, "anything")
	if err == nil {
		t.Fatal("expected error when target is not a product_blueprint")
	}
}

// ── AddBlueprintRequirement ─────────────────────────────────────────────────

func TestAddBlueprintRequirement_Success(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := AddBlueprintRequirement(p, productID, "Datenschutz-Konzept liegt vor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(got.Requirements))
	}
	req := got.Requirements[0]
	if req.Label != "Datenschutz-Konzept liegt vor" {
		t.Fatalf("unexpected label: %s", req.Label)
	}
	if req.Fulfilled {
		t.Fatal("new requirement should not be fulfilled")
	}
	if req.ID == "" {
		t.Fatal("requirement should have a non-empty ID")
	}
}

func TestAddBlueprintRequirement_MultipleRequirements(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	if _, err := AddBlueprintRequirement(p, productID, "Anforderung A"); err != nil {
		t.Fatal(err)
	}
	got, err := AddBlueprintRequirement(p, productID, "Anforderung B")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Requirements) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(got.Requirements))
	}
}

func TestAddBlueprintRequirement_EmptyLabel(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	_, err := AddBlueprintRequirement(p, productID, "")
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestAddBlueprintRequirement_BlueprintNotFound(t *testing.T) {
	p := createTestCosmosWithCatalog(t)

	_, err := AddBlueprintRequirement(p, "PB-NONEXISTENT", "some label")
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
}

// ── SetBlueprintRequirementFulfilled ────────────────────────────────────────

func TestSetBlueprintRequirementFulfilled_ToggleTrue(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	added, err := AddBlueprintRequirement(p, productID, "Sicherheitstest bestanden")
	if err != nil {
		t.Fatal(err)
	}
	reqID := added.Requirements[0].ID

	got, err := SetBlueprintRequirementFulfilled(p, productID, reqID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Requirements[0].Fulfilled {
		t.Fatal("requirement should be fulfilled after setting to true")
	}
}

func TestSetBlueprintRequirementFulfilled_ToggleFalse(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	added, _ := AddBlueprintRequirement(p, productID, "Test")
	reqID := added.Requirements[0].ID

	fulfilled, _ := SetBlueprintRequirementFulfilled(p, productID, reqID, true)
	if !fulfilled.Requirements[0].Fulfilled {
		t.Fatal("should be fulfilled")
	}

	got, err := SetBlueprintRequirementFulfilled(p, productID, reqID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Requirements[0].Fulfilled {
		t.Fatal("requirement should be open after toggling back to false")
	}
}

func TestSetBlueprintRequirementFulfilled_RequirementNotFound(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	_, err := SetBlueprintRequirementFulfilled(p, productID, "req-nonexistent", true)
	if err == nil {
		t.Fatal("expected error for non-existent requirement ID")
	}
}

// ── DeleteBlueprintRequirement ──────────────────────────────────────────────

func TestDeleteBlueprintRequirement_Success(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	added, err := AddBlueprintRequirement(p, productID, "Zu löschende Anforderung")
	if err != nil {
		t.Fatal(err)
	}
	reqID := added.Requirements[0].ID

	got, err := DeleteBlueprintRequirement(p, productID, reqID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Requirements) != 0 {
		t.Fatalf("expected 0 requirements after delete, got %d", len(got.Requirements))
	}
}

func TestDeleteBlueprintRequirement_OnlyDeletesTargeted(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	r1, _ := AddBlueprintRequirement(p, productID, "Anforderung 1")
	r2, _ := AddBlueprintRequirement(p, productID, "Anforderung 2")
	_ = r1
	reqID := r2.Requirements[1].ID

	got, err := DeleteBlueprintRequirement(p, productID, reqID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Requirements) != 1 {
		t.Fatalf("expected 1 requirement remaining, got %d", len(got.Requirements))
	}
	if got.Requirements[0].Label != "Anforderung 1" {
		t.Fatalf("wrong requirement remaining: %s", got.Requirements[0].Label)
	}
}

func TestDeleteBlueprintRequirement_NonExistentIsNoop(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := DeleteBlueprintRequirement(p, productID, "req-does-not-exist")
	if err != nil {
		t.Fatalf("deleting non-existent requirement should be noop, got: %v", err)
	}
	if len(got.Requirements) != 0 {
		t.Fatalf("expected empty list, got %v", got.Requirements)
	}
}

// ── PatchBlueprint ───────────────────────────────────────────────────────────

func TestPatchBlueprint_Fields(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := PatchBlueprint(p, productID, map[string]string{
		"name":    "Updated Name",
		"owner":   "New Owner",
		"status":  "active",
		"version": "1.0.0",
		"summary": "Updated summary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated Name" {
		t.Fatalf("name not updated: %s", got.Name)
	}
	if got.Owner != "New Owner" {
		t.Fatalf("owner not updated: %s", got.Owner)
	}
	if got.Status != "active" {
		t.Fatalf("status not updated: %s", got.Status)
	}
	if got.Version != "1.0.0" {
		t.Fatalf("version not updated: %s", got.Version)
	}
	if got.Summary != "Updated summary" {
		t.Fatalf("summary not updated: %s", got.Summary)
	}
}

func TestPatchBlueprint_NotFound(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	_, err := PatchBlueprint(p, "PB-NONEXISTENT", map[string]string{"name": "x"})
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
}

// ── PublishBlueprint ─────────────────────────────────────────────────────────

func TestPublishBlueprint_SetsStatusToPublished(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := PublishBlueprint(p, productID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != "published" {
		t.Fatalf("expected status published, got %s", got.Status)
	}
}

func TestPublishBlueprint_NotFound(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	_, err := PublishBlueprint(p, "PB-NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
}

// ── ValidateBlueprint ────────────────────────────────────────────────────────

func TestValidateBlueprint_ValidBlueprint(t *testing.T) {
	p, productID, _ := createProductAndService(t)

	got, err := ValidateBlueprint(p, productID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status == "" {
		t.Fatal("expected a non-empty validation status")
	}
}

func TestValidateBlueprint_NotFound(t *testing.T) {
	p := createTestCosmosWithCatalog(t)
	_, err := ValidateBlueprint(p, "PB-NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
}
