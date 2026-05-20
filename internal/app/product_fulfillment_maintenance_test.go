package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestProductFulfillmentUpdateRemoveAndMoveInputNormalization(t *testing.T) {
	p := createAppTestCosmos(t)
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account-maintenance.yaml"), []byte(`id: PROD-MAINT-001
type: product_blueprint
name: Account Maintenance
fulfillment:
  required_services:
    - service_ref: user-account
      role: primary
      required: true
      description: Original identity.
      sla_ref: SLA-IDENTITY-STANDARD
`), 0o644))

	updated, err := UpdateProductFulfillmentService(p, "PROD-MAINT-001", 0, UpdateFulfillmentServiceRequest{ServiceRef: "rule-validation-api", Role: "supporting", Required: false, Description: "Updated validation."})
	if err != nil {
		t.Fatal(err)
	}
	row := updated.Fulfillment.RequiredServices[0]
	if row.ServiceRef != "rule-validation-api" || row.Role != "supporting" || row.Required || row.Description != "Updated validation." {
		t.Fatalf("expected updated fulfillment row: %+v", row)
	}
	b, err := os.ReadFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account-maintenance.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	raw := string(b)
	if !strings.Contains(raw, "sla_ref: SLA-IDENTITY-STANDARD") {
		t.Fatalf("expected update to preserve SLA ref, got:\n%s", raw)
	}

	if _, err := AddProductFulfillmentService(p, "PROD-MAINT-001", AddFulfillmentServiceRequest{ServiceRef: "user-account", Role: "primary", Required: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateProductFulfillmentService(p, "PROD-MAINT-001", 1, UpdateFulfillmentServiceRequest{ServiceRef: "rule-validation-api", Role: "primary", Required: true}); err == nil {
		t.Fatal("expected duplicate fulfillment ref on update to be rejected")
	}
	updated, err = RemoveProductFulfillmentService(p, "PROD-MAINT-001", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Fulfillment.RequiredServices) != 1 || updated.Fulfillment.RequiredServices[0].ServiceRef != "user-account" {
		t.Fatalf("expected first fulfillment removed: %+v", updated.Fulfillment.RequiredServices)
	}
	if _, err := RemoveProductFulfillmentService(p, "PROD-MAINT-001", 9); err == nil {
		t.Fatal("expected invalid fulfillment index to be rejected")
	}
}
