package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateProductBlueprintMissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "cosmos.yaml"), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(dir, "catalog", "blueprints", "products", "broken.yaml"), `
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
	mustWrite(t, filepath.Join(dir, "cosmos.yaml"), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(dir, "catalog", "instances", "services", "broken.yaml"), `
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
	mustWrite(t, filepath.Join(dir, "cosmos.yaml"), "id: test\ntype: cosmos\n")
	mustWrite(t, filepath.Join(dir, "catalog", "products", "legacy.yaml"), `
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
