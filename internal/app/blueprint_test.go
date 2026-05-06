package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/model"
)

func createTestCosmosWithCatalog(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(p, "catalog", "blueprints", "products"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "catalog", "blueprints", "services"), 0o755))
	must(os.WriteFile(filepath.Join(p, "cosmos.yaml"), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Platform Team\n"), 0o644))
	return p
}

func TestCreateBlueprint_ProductBlueprint(t *testing.T) {
	p := createTestCosmosWithCatalog(t)

	bp := model.Blueprint{
		ID:      "PB-TEST-001",
		Type:    "product_blueprint",
		Name:    "Test Product Blueprint",
		Version: "0.1.0",
		Status:  "draft",
		Owner:   "Product Team",
		Summary: "A test product blueprint for the test.",
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
	path := filepath.Join(p, "catalog", "blueprints", "products", "PB-TEST-001.yaml")
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
		ID:      "SB-TEST-001",
		Type:    "service_blueprint",
		Name:    "Test Service Blueprint",
		Version: "0.1.0",
		Status:  "draft",
		Owner:   "Platform Team",
		Summary: "A test service blueprint.",
		NamespaceServiceRef: "identity.blumer.cloud/user-account",
	}

	err := CreateBlueprint(p, bp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(p, "catalog", "blueprints", "services", "SB-TEST-001.yaml")
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

	path := filepath.Join(p, "catalog", "blueprints", "products", "PB-DEL-001.yaml")
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
