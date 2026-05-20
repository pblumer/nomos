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

func TestBlueprintAttributeServiceRefYAMLParsing(t *testing.T) {
	raw := []byte(`id: PB-ATTR-SVC-001
type: product_blueprint
name: Attribute Service Ref
attributes:
  - id: attr-service
    label: Provisionierungsservice
    type: service_ref
    required: true
    service_ref: identity.blumer.cloud/user-account
`)
	var bp Blueprint
	if err := yaml.Unmarshal(raw, &bp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(bp.Attributes) != 1 || bp.Attributes[0].Type != "service_ref" || bp.Attributes[0].ServiceRef != "identity.blumer.cloud/user-account" {
		t.Fatalf("expected service_ref attribute mapping: %#v", bp.Attributes)
	}
}

func TestProductBlueprintOwnershipFulfillmentYAMLParsing(t *testing.T) {
	raw := []byte(`
id: PROD-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Identity & Collaboration
offered_by: cloud.blumer.identity
owning_domain: cloud.blumer.identity
fulfillment:
  required_services:
    - service_ref: cloud.blumer.identity/user-account
      role: primary
      required: true
      description: Creates or manages the identity account.
`)
	var bp Blueprint
	if err := yaml.Unmarshal(raw, &bp); err != nil {
		t.Fatalf("unmarshal blueprint: %v", err)
	}
	if len(bp.Fulfillment.RequiredServices) != 1 || bp.Fulfillment.RequiredServices[0].Role != "primary" || !bp.Fulfillment.RequiredServices[0].Required {
		t.Fatalf("expected fulfillment services: %#v", bp.Fulfillment)
	}
}

func TestServiceOwnershipYAMLParsing(t *testing.T) {
	raw := []byte(`
id: service-user-account
type: service
name: user-account
version: 0.1.0
status: draft
owner: Identity Team
owned_by: cloud.blumer.identity
operated_by:
  - cloud.blumer.identity
capabilities:
  - user-account-management
supported_products:
  - PROD-ACC-MBX-001
`)
	var svc Service
	if err := yaml.Unmarshal(raw, &svc); err != nil {
		t.Fatalf("unmarshal service: %v", err)
	}
	if len(svc.Capabilities) != 1 || len(svc.SupportedProducts) != 1 {
		t.Fatalf("expected service metadata: %#v", svc)
	}
}
