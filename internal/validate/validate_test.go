package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestValidateProductBlueprintMissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "blueprints", "products", "broken.yaml"), `
type: product_blueprint
name: Broken
version: 0.1.0
status: draft
owner: Team
required_inputs: []
required_service_blueprints: []
`)
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if res.Status != "failed" {
		t.Fatalf("expected failed validation, got %#v", res)
	}
	assertFinding(t, res, "BLUEPRINT_REQUIRED_FIELD")
	assertFinding(t, res, "PRODUCT_BLUEPRINT_INPUTS_EMPTY")
	assertFinding(t, res, "PRODUCT_BLUEPRINT_SERVICES_EMPTY")
}

func TestValidateInstanceComplianceStatus(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "instances", "services", "broken.yaml"), `
id: SI-1
type: service_instance
blueprint_ref: SB-1
blueprint_version: 0.1.0
compliance_status: strange
findings: []
`)
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	assertFinding(t, res, "INSTANCE_COMPLIANCE_STATUS_INVALID")
}

func TestValidateLegacyProductIsBackwardCompatible(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "products", "legacy.yaml"), `
id: PROD-1
type: product
name: Legacy Product
`)
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("legacy product should not create findings: %#v", res.Findings)
	}
}

func TestValidateProductBlueprintRequiredServices(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "blueprints", "products", "broken-required-services.yaml"), `
id: PB-1
type: product_blueprint
name: Broken Required Services
version: 0.1.0
status: draft
owner: Team
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-1
required_services:
  - service_ref: ""
    service_blueprint_ref: SB-1
    required: true
  - service_ref: identity.blumer.cloud/user-account
    service_blueprint_ref: ""
    required: true
`)
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if res.Status != "failed" {
		t.Fatalf("expected failed validation, got %#v", res)
	}
	assertFinding(t, res, "PRODUCT_BLUEPRINT_REQUIRED_SERVICE_REF_EMPTY")
	assertFinding(t, res, "PRODUCT_BLUEPRINT_REQUIRED_SERVICE_BLUEPRINT_REF_EMPTY")
}

func TestValidateProductBlueprintRequiredServicesRecommendedWarning(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "blueprints", "products", "legacy-blueprint.yaml"), `
id: PB-LEGACY
type: product_blueprint
name: Legacy Blueprint
version: 0.1.0
status: draft
owner: Team
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-1
`)
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	assertFinding(t, res, "PRODUCT_BLUEPRINT_REQUIRED_SERVICES_RECOMMENDED")
	assertFinding(t, res, "PRODUCT_OFFERED_BY_MISSING")
}

func TestValidateServiceBlueprintNamespaceServiceRefShapeWarning(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "blueprints", "services", "mailbox.yaml"), `
id: SB-MAILBOX-001
type: service_blueprint
name: Mailbox Service
version: 0.1.0
status: draft
owner: Team
namespace_service_ref: collaboration.blumer.cloud
capabilities:
  - create_mailbox
`)
	res, err := Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if res.Status != "ok" {
		t.Fatalf("warning-only validation should stay ok, got %#v", res)
	}
	assertFinding(t, res, "SERVICE_BLUEPRINT_NAMESPACE_SERVICE_REF_SHAPE")
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertFinding(t *testing.T, res Result, code string) {
	t.Helper()
	for _, f := range res.Findings {
		if f.Code == code {
			return
		}
	}
	t.Fatalf("finding %s not found in %#v", code, res.Findings)
}

func TestValidateDomainOwnedProductFulfillment(t *testing.T) {
	t.Run("valid cross-domain fulfillment", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		res, err := Validate(dir)
		if err != nil {
			t.Fatalf("validate: %v", err)
		}
		if res.Status != "ok" {
			t.Fatalf("expected valid cosmos, got %#v", res)
		}
	})
	t.Run("product without offered_by", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		rewriteBlueprint(t, dir, "offered_by: identity.blumer.cloud\n", "")
		res, _ := Validate(dir)
		assertFinding(t, res, "PRODUCT_OFFERED_BY_MISSING")
	})
	t.Run("product with unresolved offered_by", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		rewriteBlueprint(t, dir, "offered_by: identity.blumer.cloud", "offered_by: missing.blumer.cloud")
		res, _ := Validate(dir)
		assertFinding(t, res, "PRODUCT_OFFERED_BY_UNRESOLVED")
	})
	t.Run("missing required service domain", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		rewriteBlueprint(t, dir, "collaboration.blumer.cloud/mailbox", "missing.blumer.cloud/mailbox")
		res, _ := Validate(dir)
		assertFinding(t, res, "FULFILLMENT_SERVICE_DOMAIN_UNRESOLVED")
	})
	t.Run("missing required service below existing domain", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		rewriteBlueprint(t, dir, "collaboration.blumer.cloud/mailbox", "collaboration.blumer.cloud/missing")
		res, _ := Validate(dir)
		assertFinding(t, res, "FULFILLMENT_SERVICE_UNRESOLVED")
	})
	t.Run("service without owned_by", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		servicePath := filepath.Join(storage.DomainsDir(dir), "cloud", "blumer", "identity", "services", "user-account", "service.yaml")
		b, _ := os.ReadFile(servicePath)
		mustWrite(t, servicePath, strings.Replace(string(b), "owned_by: identity.blumer.cloud\n", "", 1))
		res, _ := Validate(dir)
		assertFinding(t, res, "SERVICE_OWNED_BY_MISSING")
	})
	t.Run("service with unresolved owned_by", func(t *testing.T) {
		dir := t.TempDir()
		writeValidFulfillmentCosmos(t, dir)
		servicePath := filepath.Join(storage.DomainsDir(dir), "cloud", "blumer", "identity", "services", "user-account", "service.yaml")
		b, _ := os.ReadFile(servicePath)
		mustWrite(t, servicePath, strings.Replace(string(b), "owned_by: identity.blumer.cloud", "owned_by: missing.blumer.cloud", 1))
		res, _ := Validate(dir)
		assertFinding(t, res, "SERVICE_OWNED_BY_UNRESOLVED")
	})
}

func writeValidFulfillmentCosmos(t *testing.T, dir string) {
	t.Helper()
	mustWrite(t, storage.CosmosFile(dir), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(storage.DomainsDir(dir), "cloud", "blumer", "identity", "domain.yaml"), "name: identity.blumer.cloud\n")
	mustWrite(t, filepath.Join(storage.DomainsDir(dir), "cloud", "blumer", "collaboration", "domain.yaml"), "name: collaboration.blumer.cloud\n")
	mustWrite(t, filepath.Join(storage.DomainsDir(dir), "cloud", "blumer", "identity", "services", "user-account", "service.yaml"), "name: user-account\nowned_by: identity.blumer.cloud\n")
	mustWrite(t, filepath.Join(storage.DomainsDir(dir), "cloud", "blumer", "collaboration", "services", "mailbox", "service.yaml"), "name: mailbox\nowned_by: collaboration.blumer.cloud\n")
	mustWrite(t, filepath.Join(storage.CatalogDir(dir), "blueprints", "products", "account.yaml"), `id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity Team
offered_by: identity.blumer.cloud
required_inputs:
  - person_reference
fulfillment:
  required_services:
    - service_ref: identity.blumer.cloud/user-account
      required: true
    - service_ref: collaboration.blumer.cloud/mailbox
      required: true
`)
}

func rewriteBlueprint(t *testing.T, dir, old, new string) {
	t.Helper()
	path := filepath.Join(storage.CatalogDir(dir), "blueprints", "products", "account.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, path, strings.Replace(string(b), old, new, 1))
}
