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
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "blumer.cloud"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "domain.yaml"), []byte("name: blumer.cloud\nowner: Cloud Team\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "account.blumer.cloud"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "account.blumer.cloud", "domain.yaml"), []byte("name: account.blumer.cloud\nowner: Account Team\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "services", "user-account"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "domain.yaml"), []byte("name: identity.blumer.cloud\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "identity.blumer.cloud", "services", "user-account", "service.yaml"), []byte("name: user-account\nowner: Identity Team\nowned_by: identity.blumer.cloud\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud", "services", "rule-validation-api"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud", "domain.yaml"), []byte("name: platform.blumer.cloud\nowner: Platform Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "platform.blumer.cloud", "services", "rule-validation-api", "service.yaml"), []byte("name: rule-validation-api\nowner: Platform Team\nowned_by: platform.blumer.cloud\nstatus: draft\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.WriteFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account-maintenance.yaml"), []byte(`id: PROD-MAINT-001
type: product_blueprint
name: Account Maintenance
offered_by: blumer.cloud
owning_domain: blumer.cloud
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      role: primary
      required: true
      description: Original identity.
      sla_ref: SLA-IDENTITY-STANDARD
`), 0o644))

	updated, err := UpdateProductFulfillmentService(p, "PROD-MAINT-001", 0, UpdateFulfillmentServiceRequest{ServiceRef: "platform.blumer.cloud/rule-validation-api", Role: "supporting", Required: false, Description: "Updated validation."})
	if err != nil {
		t.Fatal(err)
	}
	row := updated.Fulfillment.RequiredServices[0]
	if row.ServiceRef != "platform.blumer.cloud/rule-validation-api" || row.Role != "supporting" || row.Required || row.Description != "Updated validation." {
		t.Fatalf("expected updated fulfillment row: %+v", row)
	}
	var raw string
	b, err := os.ReadFile(filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account-maintenance.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	raw = string(b)
	if !strings.Contains(raw, "sla_ref: SLA-IDENTITY-STANDARD") {
		t.Fatalf("expected update to preserve SLA ref, got:\n%s", raw)
	}

	if _, err := AddProductFulfillmentService(p, "PROD-MAINT-001", AddFulfillmentServiceRequest{ServiceRef: "identity.blumer.cloud/user-account", Role: "primary", Required: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateProductFulfillmentService(p, "PROD-MAINT-001", 1, UpdateFulfillmentServiceRequest{ServiceRef: "platform.blumer.cloud/rule-validation-api", Role: "primary", Required: true}); err == nil {
		t.Fatal("expected duplicate fulfillment ref on update to be rejected")
	}
	updated, err = RemoveProductFulfillmentService(p, "PROD-MAINT-001", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Fulfillment.RequiredServices) != 1 || updated.Fulfillment.RequiredServices[0].ServiceRef != "identity.blumer.cloud/user-account" {
		t.Fatalf("expected first fulfillment removed: %+v", updated.Fulfillment.RequiredServices)
	}
	if _, err := RemoveProductFulfillmentService(p, "PROD-MAINT-001", 9); err == nil {
		t.Fatal("expected invalid fulfillment index to be rejected")
	}

	moved, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-MAINT-001", TargetDomain: `"account.blumer.cloud"`, UpdateOwningDomain: true})
	if err != nil {
		t.Fatalf("expected quoted canonical target to normalize, got %v", err)
	}
	if moved.OfferedBy != "account.blumer.cloud" || moved.OwningDomain != "account.blumer.cloud" {
		t.Fatalf("unexpected moved product: %+v", moved)
	}
	if _, err := MoveProductOffering(p, MoveProductOfferingRequest{ProductID: "PROD-MAINT-001", TargetDomain: `"account.blumer.cloud"`}); err == nil || !strings.Contains(err.Error(), "Bitte eine andere Zieldomäne wählen.") {
		t.Fatalf("expected same-domain move error, got %v", err)
	}
}
