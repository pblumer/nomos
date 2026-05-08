package graph

import (
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/cosmosfs"
	"github.com/nomos/nomos/internal/model"
)

func TestMermaid_EmptyTree(t *testing.T) {
	tree := cosmosfs.Tree{
		Cosmos: model.Cosmos{Name: "Test Cosmos"},
	}
	out := Mermaid(tree)
	if !strings.Contains(out, "graph TD") {
		t.Fatal("expected mermaid graph directive")
	}
	if !strings.Contains(out, "Test Cosmos") {
		t.Fatal("expected cosmos name in output")
	}
}

func TestMermaid_WithBlueprint(t *testing.T) {
	tree := cosmosfs.Tree{
		Cosmos: model.Cosmos{Name: "My Cosmos"},
		Blueprints: []cosmosfs.BlueprintNode{
			{Metadata: model.Blueprint{
				ID:                        "PB-001",
				Name:                      "Product Blueprint",
				Type:                      "product_blueprint",
				RequiredServiceBlueprints: []string{"SB-001"},
			}},
			{Metadata: model.Blueprint{
				ID:   "SB-001",
				Name: "Service Blueprint",
				Type: "service_blueprint",
			}},
		},
	}
	out := Mermaid(tree)
	if !strings.Contains(out, "PB-001") {
		t.Fatalf("expected PB-001 in output: %s", out)
	}
	if !strings.Contains(out, "SB-001") {
		t.Fatalf("expected SB-001 in output: %s", out)
	}
	if !strings.Contains(out, "requires") {
		t.Fatalf("expected 'requires' edge in output: %s", out)
	}
}

func TestMermaid_WithRequiredServices(t *testing.T) {
	tree := cosmosfs.Tree{
		Cosmos: model.Cosmos{Name: "Cosmos"},
		Blueprints: []cosmosfs.BlueprintNode{
			{Metadata: model.Blueprint{
				ID:   "PB-002",
				Name: "Product",
				Type: "product_blueprint",
				RequiredServices: []model.RequiredServiceRef{
					{ServiceRef: "identity.example/user", ServiceBlueprintRef: "SB-002", Required: true},
				},
			}},
		},
	}
	out := Mermaid(tree)
	if !strings.Contains(out, "identity.example/user") {
		t.Fatalf("expected service ref in output: %s", out)
	}
	if !strings.Contains(out, "SB-002") {
		t.Fatalf("expected blueprint ref in output: %s", out)
	}
}

func TestMermaid_WithInstances(t *testing.T) {
	tree := cosmosfs.Tree{
		Cosmos: model.Cosmos{Name: "Cosmos"},
		Instances: []cosmosfs.InstanceNode{
			{Metadata: model.Instance{
				ID:                          "PI-001",
				Name:                        "Product One",
				Type:                        "product_instance",
				BlueprintRef:                "PB-001",
				ProvisionedServiceInstances: []string{"SI-001"},
			}},
			{Metadata: model.Instance{
				ID:           "SI-001",
				Name:         "Service One",
				Type:         "service_instance",
				BlueprintRef: "SB-001",
			}},
		},
	}
	out := Mermaid(tree)
	if !strings.Contains(out, "Product One") {
		t.Fatalf("expected instance name in output: %s", out)
	}
	if !strings.Contains(out, "Service One") {
		t.Fatalf("expected service instance name in output: %s", out)
	}
	if !strings.Contains(out, "Instance:") {
		t.Fatalf("expected Instance: label in output: %s", out)
	}
}

func TestMermaid_FallbackCosmosName(t *testing.T) {
	tree := cosmosfs.Tree{
		Cosmos: model.Cosmos{Name: ""},
		Path:   "/some/path",
	}
	out := Mermaid(tree)
	if !strings.Contains(out, "/some/path") {
		t.Fatalf("expected path as cosmos name fallback: %s", out)
	}
}
