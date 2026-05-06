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
required_services:
  - service_ref: identity.blumer.cloud/user-account
    service_blueprint_ref: SB-IDENTITY-ACCOUNT-001
    purpose: Erstellt und verwaltet das Benutzerkonto.
    required: true
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
	if len(bp.RequiredServices) != 1 || bp.RequiredServices[0].ServiceRef != "identity.blumer.cloud/user-account" || bp.RequiredServices[0].ServiceBlueprintRef != "SB-IDENTITY-ACCOUNT-001" || !bp.RequiredServices[0].Required {
		t.Fatalf("expected required services mapping: %#v", bp.RequiredServices)
	}
}

func TestServiceBlueprintNamespaceServiceRefYAMLParsing(t *testing.T) {
	raw := []byte(`
id: SB-MAILBOX-001
type: service_blueprint
name: Mailbox Service
version: 0.1.0
status: draft
owner: Identity & Collaboration
namespace_service_ref: collaboration.blumer.cloud/mailbox
capabilities:
  - create_mailbox
`)
	var bp Blueprint
	if err := yaml.Unmarshal(raw, &bp); err != nil {
		t.Fatalf("unmarshal blueprint: %v", err)
	}
	if bp.NamespaceServiceRef != "collaboration.blumer.cloud/mailbox" {
		t.Fatalf("expected namespace service ref, got %#v", bp)
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
