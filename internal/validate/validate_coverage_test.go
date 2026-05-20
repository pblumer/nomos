package validate

import (
	"testing"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/model"
)

// ---------------------------------------------------------------------------
// requiredSeverity
// ---------------------------------------------------------------------------

func TestRequiredSeverity(t *testing.T) {
	if requiredSeverity(true) != "error" {
		t.Error("expected error for required=true")
	}
	if requiredSeverity(false) != "warning" {
		t.Error("expected warning for required=false")
	}
}

// ---------------------------------------------------------------------------
// splitServiceRef
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// validateBlueprint (unit — no disk I/O needed)
// ---------------------------------------------------------------------------

func newResult() *Result {
	return &Result{Status: "ok", Findings: []Finding{}}
}

func hasCode(res *Result, code string) bool {
	for _, f := range res.Findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

func TestValidateBlueprintServiceBlueprintEmpty(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-1", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
	}, "test.yaml", res)
	if !hasCode(res, "SERVICE_BLUEPRINT_CAPABILITY_EMPTY") {
		t.Error("expected SERVICE_BLUEPRINT_CAPABILITY_EMPTY finding")
	}
}

func TestValidateBlueprintServiceBlueprintWithCapability(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-2", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		Capabilities: []string{"create_user"},
	}, "test.yaml", res)
	if hasCode(res, "SERVICE_BLUEPRINT_CAPABILITY_EMPTY") {
		t.Error("unexpected SERVICE_BLUEPRINT_CAPABILITY_EMPTY finding for non-empty blueprint")
	}
}

func TestValidateBlueprintAttributeServiceRefShape(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-3", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		Attributes: []model.BlueprintAttribute{
			{ID: "attr1", Label: "Attr1", Type: "service_ref", ServiceRef: "nodomain"},
		},
		Capabilities: []string{"cap"},
	}, "test.yaml", res)
	if !hasCode(res, "BLUEPRINT_ATTRIBUTE_SERVICE_REF_SHAPE") {
		t.Error("expected BLUEPRINT_ATTRIBUTE_SERVICE_REF_SHAPE warning")
	}
}

func TestValidateBlueprintAttributeServiceRefWrongType(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-4", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		Attributes: []model.BlueprintAttribute{
			{ID: "attr1", Label: "Attr1", Type: "text", ServiceRef: "identity.blumer.cloud/svc"},
		},
		Capabilities: []string{"cap"},
	}, "test.yaml", res)
	if !hasCode(res, "BLUEPRINT_ATTRIBUTE_SERVICE_REF_TYPE") {
		t.Error("expected BLUEPRINT_ATTRIBUTE_SERVICE_REF_TYPE error")
	}
}

func TestValidateBlueprintAttributeLabelEmpty(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-5", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		Attributes: []model.BlueprintAttribute{
			{ID: "attr1", Label: "", Type: "text"},
		},
		Capabilities: []string{"cap"},
	}, "test.yaml", res)
	if !hasCode(res, "BLUEPRINT_ATTRIBUTE_LABEL_EMPTY") {
		t.Error("expected BLUEPRINT_ATTRIBUTE_LABEL_EMPTY error")
	}
}

func TestValidateBlueprintRequirementUnknownAttrRef(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-6", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		Attributes: []model.BlueprintAttribute{
			{ID: "known-attr", Label: "Known", Type: "text"},
		},
		Requirements: []model.BlueprintRequirement{
			{AttributeRefs: []string{"unknown-attr"}},
		},
		Capabilities: []string{"cap"},
	}, "test.yaml", res)
	if !hasCode(res, "BLUEPRINT_REQUIREMENT_ATTRIBUTE_REF_UNKNOWN") {
		t.Error("expected BLUEPRINT_REQUIREMENT_ATTRIBUTE_REF_UNKNOWN error")
	}
}

func TestValidateBlueprintRequirementEmptyAttrRef(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-7", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		Requirements: []model.BlueprintRequirement{
			{AttributeRefs: []string{""}},
		},
		Capabilities: []string{"cap"},
	}, "test.yaml", res)
	if !hasCode(res, "BLUEPRINT_REQUIREMENT_ATTRIBUTE_REF_EMPTY") {
		t.Error("expected BLUEPRINT_REQUIREMENT_ATTRIBUTE_REF_EMPTY warning")
	}
}

func TestValidateBlueprintNamespaceServiceRefValid(t *testing.T) {
	res := newResult()
	validateBlueprint(model.Blueprint{
		ID: "SB-8", Type: "service_blueprint", Name: "Test", Version: "0.1.0", Status: "draft", Owner: "Team",
		NamespaceServiceRef: "identity.blumer.cloud/user-account",
		Capabilities:        []string{"cap"},
	}, "test.yaml", res)
	// shape warning should NOT appear when ref contains "/"
	if hasCode(res, "SERVICE_BLUEPRINT_NAMESPACE_SERVICE_REF_SHAPE") {
		t.Error("unexpected shape warning for valid service ref")
	}
}

// ---------------------------------------------------------------------------
// validateCrossArtifacts (unit — in-memory Tree)
// ---------------------------------------------------------------------------

func TestValidateCrossArtifactsBlueprintServiceRefMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/bp.yaml", Metadata: model.Blueprint{
				ID: "SB-X", Type: "service_blueprint",
				NamespaceServiceRef: "identity.blumer.cloud/missing-svc",
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "BLUEPRINT_SERVICE_REF_MISSING") {
		t.Errorf("expected BLUEPRINT_SERVICE_REF_MISSING, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsBlueprintAttributeServiceRefMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/bp.yaml", Metadata: model.Blueprint{
				ID: "SB-Y", Type: "service_blueprint",
				Attributes: []model.BlueprintAttribute{
					{ID: "svc-attr", Type: "service_ref", ServiceRef: "identity.blumer.cloud/nonexistent"},
				},
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "BLUEPRINT_ATTRIBUTE_SERVICE_REF_MISSING") {
		t.Errorf("expected BLUEPRINT_ATTRIBUTE_SERVICE_REF_MISSING, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsBlueprintRequiredServiceRefMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/bp.yaml", Metadata: model.Blueprint{
				ID: "PB-Z", Type: "product_blueprint",
				RequiredServices: []model.RequiredServiceRef{
					{ServiceRef: "identity.blumer.cloud/ghost-svc"},
				},
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "BLUEPRINT_REQUIRED_SERVICE_REF_MISSING") {
		t.Errorf("expected BLUEPRINT_REQUIRED_SERVICE_REF_MISSING, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsBlueprintRequiredServiceBlueprintRefMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/bp.yaml", Metadata: model.Blueprint{
				ID: "PB-W", Type: "product_blueprint",
				RequiredServices: []model.RequiredServiceRef{
					{ServiceBlueprintRef: "SB-NONEXISTENT"},
				},
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "BLUEPRINT_SERVICE_BLUEPRINT_REF_MISSING") {
		t.Errorf("expected BLUEPRINT_SERVICE_BLUEPRINT_REF_MISSING, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsInstanceBlueprintRefMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Instances: []cosmosfs.InstanceNode{
			{Path: "/test/inst.yaml", Metadata: model.Instance{
				ID: "INST-1", Type: "product_instance",
				BlueprintRef: "PB-NONEXISTENT",
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "INSTANCE_BLUEPRINT_REF_MISSING") {
		t.Errorf("expected INSTANCE_BLUEPRINT_REF_MISSING, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsInstanceTypeMismatch(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/sb.yaml", Metadata: model.Blueprint{ID: "SB-M", Type: "service_blueprint"}},
		},
		Instances: []cosmosfs.InstanceNode{
			{Path: "/test/inst.yaml", Metadata: model.Instance{
				ID: "INST-2", Type: "product_instance",
				BlueprintRef: "SB-M", BlueprintVersion: "0.1.0",
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "INSTANCE_TYPE_MISMATCH") {
		t.Errorf("expected INSTANCE_TYPE_MISMATCH, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsInstanceRequiredInputMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/pb.yaml", Metadata: model.Blueprint{
				ID: "PB-INP", Type: "product_blueprint",
				RequiredInputs: []string{"person_reference"},
			}},
		},
		Instances: []cosmosfs.InstanceNode{
			{Path: "/test/inst.yaml", Metadata: model.Instance{
				ID: "INST-3", Type: "product_instance",
				BlueprintRef: "PB-INP", BlueprintVersion: "0.1.0",
				ComplianceStatus: "compliant",
				Inputs:           map[string]string{},
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "INSTANCE_REQUIRED_INPUT_MISSING") {
		t.Errorf("expected INSTANCE_REQUIRED_INPUT_MISSING, got %v", res.Findings)
	}
}

func TestValidateCrossArtifactsInstanceEvidenceMissing(t *testing.T) {
	res := newResult()
	tree := cosmosfs.Tree{
		Path: "/test",
		Blueprints: []cosmosfs.BlueprintNode{
			{Path: "/test/pb.yaml", Metadata: model.Blueprint{
				ID: "PB-EV", Type: "product_blueprint",
				EvidenceRequirements: []string{"EV-001"},
			}},
		},
		Instances: []cosmosfs.InstanceNode{
			{Path: "/test/inst.yaml", Metadata: model.Instance{
				ID: "INST-4", Type: "product_instance",
				BlueprintRef: "PB-EV", BlueprintVersion: "0.1.0",
				ComplianceStatus: "compliant",
			}},
		},
	}
	validateCrossArtifacts(tree, res)
	if !hasCode(res, "INSTANCE_EVIDENCE_MISSING") {
		t.Errorf("expected INSTANCE_EVIDENCE_MISSING, got %v", res.Findings)
	}
}

// ---------------------------------------------------------------------------
// isYAML / relPath helpers
// ---------------------------------------------------------------------------

func TestIsYAML(t *testing.T) {
	cases := map[string]bool{
		"foo.yaml": true, "foo.yml": true, "foo.YAML": true,
		"foo.json": false, "foo.txt": false, "noeext": false,
	}
	for path, want := range cases {
		if isYAML(path) != want {
			t.Errorf("isYAML(%q) = %v, want %v", path, !want, want)
		}
	}
}

func TestRelPath(t *testing.T) {
	got := relPath("/root", "/root/sub/file.yaml")
	if got != "sub/file.yaml" {
		t.Errorf("relPath = %q", got)
	}
	// error path: relPath returns the path as-is
	got = relPath("", "absolute/path.yaml")
	if got == "" {
		t.Error("expected non-empty relPath on error")
	}
}
