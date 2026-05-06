package model

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBlueprintYAMLParsing(t *testing.T) {
	raw := []byte(`
id: PB-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity & Collaboration
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-IDENTITY-ACCOUNT-001
quality_criteria:
  - account_exists
evidence_requirements:
  - account_creation_response
`)
	var bp Blueprint
	if err := yaml.Unmarshal(raw, &bp); err != nil {
		t.Fatalf("unmarshal blueprint: %v", err)
	}
	if bp.ID != "PB-ACC-MBX-001" || bp.Type != "product_blueprint" {
		t.Fatalf("unexpected blueprint: %#v", bp)
	}
	if len(bp.RequiredInputs) != 1 || len(bp.RequiredServiceBlueprints) != 1 {
		t.Fatalf("expected required inputs and service blueprints: %#v", bp)
	}
}

func TestInstanceYAMLParsing(t *testing.T) {
	raw := []byte(`
id: SI-MAILBOX-EXAMPLE-001
type: service_instance
name: Beispiel Mailbox Service Instance
blueprint_ref: SB-MAILBOX-001
blueprint_version: 0.1.0
status: active
observed_state:
  mailbox_status: active
compliance_status: warning
evidence:
  - id: EV-MAILBOX-001
    type: target_system_snapshot
    summary: Mailbox wurde gefunden.
findings:
  - id: FIND-MAILBOX-001
    severity: warning
    category: naming
    summary: Alte Namenskonvention.
`)
	var inst Instance
	if err := yaml.Unmarshal(raw, &inst); err != nil {
		t.Fatalf("unmarshal instance: %v", err)
	}
	if inst.ComplianceStatus != "warning" || inst.ObservedState["mailbox_status"] != "active" {
		t.Fatalf("unexpected instance: %#v", inst)
	}
	if len(inst.Evidence) != 1 || len(inst.Findings) != 1 {
		t.Fatalf("expected evidence and finding: %#v", inst)
	}
}
