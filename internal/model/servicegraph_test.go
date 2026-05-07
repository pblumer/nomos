package model

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestServicegraphYAMLRoundTrip(t *testing.T) {
	sg := Servicegraph{
		ID:             "SG-TEST-001",
		Type:           "servicegraph",
		Name:           "Test Graph",
		Version:        "0.1.0",
		Status:         "draft",
		Owner:          "test-team",
		Summary:        "A test servicegraph",
		RelatedProduct: "PROD-TEST-001",
		Variants: []GraphVariant{
			{ID: "V-001", Name: "Standard", Context: "default"},
		},
		Nodes: []GraphNode{
			{ID: "N-001", Type: "service", Name: "Main Service", Mandatory: true, Reusable: false},
			{ID: "N-002", Type: "activity", Name: "Step A", Mandatory: true, Reusable: true, SkillRef: "SKILL-001"},
		},
		Edges: []GraphEdge{
			{ID: "E-001", Source: "N-001", Target: "N-002", Type: "composed_of", Binding: "hard"},
		},
		Rules: []GraphRule{
			{ID: "G-001", Name: "Must be active", Type: "activation", Scope: "N-002", Binding: "must", Expression: "always"},
		},
	}

	data, err := yaml.Marshal(sg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Servicegraph
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != sg.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, sg.ID)
	}
	if decoded.Type != "servicegraph" {
		t.Errorf("Type mismatch: got %q", decoded.Type)
	}
	if len(decoded.Nodes) != 2 {
		t.Errorf("Nodes count: got %d, want 2", len(decoded.Nodes))
	}
	if len(decoded.Edges) != 1 {
		t.Errorf("Edges count: got %d, want 1", len(decoded.Edges))
	}
	if len(decoded.Rules) != 1 {
		t.Errorf("Rules count: got %d, want 1", len(decoded.Rules))
	}
	if len(decoded.Variants) != 1 {
		t.Errorf("Variants count: got %d, want 1", len(decoded.Variants))
	}
	if decoded.Nodes[1].SkillRef != "SKILL-001" {
		t.Errorf("SkillRef: got %q", decoded.Nodes[1].SkillRef)
	}
	if decoded.RelatedProduct != "PROD-TEST-001" {
		t.Errorf("RelatedProduct: got %q", decoded.RelatedProduct)
	}
}
