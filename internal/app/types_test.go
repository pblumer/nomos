package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

func TestSaveAndListTypeDefs(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := SaveTypeDef(p, model.TypeDef{
		ID:    "widget",
		Label: "Widget",
		File:  "widget.yaml",
		Properties: []model.TypeProperty{
			{Name: "name", Type: "string", Required: true},
		},
		Dependencies: []model.TypeDependency{{Type: "service", Relation: "uses"}},
	}); err != nil {
		t.Fatalf("SaveTypeDef: %v", err)
	}
	// Written to .nomos/types/<id>.yaml.
	if _, err := os.Stat(filepath.Join(storage.TypesDir(p), "widget.yaml")); err != nil {
		t.Fatalf("type def file not created: %v", err)
	}
	defs, err := ListTypeDefs(p)
	if err != nil {
		t.Fatalf("ListTypeDefs: %v", err)
	}
	if len(defs) != 1 || defs[0].ID != "widget" || len(defs[0].Properties) != 1 {
		t.Fatalf("unexpected defs: %+v", defs)
	}

	// Invalid ids are rejected.
	if _, err := SaveTypeDef(p, model.TypeDef{ID: "Not Valid"}); err == nil {
		t.Fatal("expected rejection of invalid type id")
	}

	// Delete removes the definition; a second delete reports not found.
	if err := DeleteTypeDef(p, "widget"); err != nil {
		t.Fatalf("DeleteTypeDef: %v", err)
	}
	if _, err := os.Stat(filepath.Join(storage.TypesDir(p), "widget.yaml")); !os.IsNotExist(err) {
		t.Fatal("expected type def file to be removed")
	}
	if err := DeleteTypeDef(p, "widget"); err == nil {
		t.Fatal("expected not-found error deleting a missing type")
	}
}

func TestRepoTreeShowsTypeDefs(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := SaveTypeDef(p, model.TypeDef{ID: "widget", Label: "Widget", Viewer: "form"}); err != nil {
		t.Fatal(err)
	}
	nodes := BuildRepoTree(p)
	nomos := findChild(nodes, "folder", ".nomos")
	if nomos == nil {
		t.Fatal("expected .nomos folder")
	}
	types := findChild(nomos.Children, "types-folder", "types")
	if types == nil {
		t.Fatal("expected types/ folder under .nomos as a types-folder node")
	}
	td := findChild(types.Children, "type-def", "Widget")
	if td == nil {
		t.Fatal("expected Widget type-def node")
	}
	if td.TypeDef == nil || td.TypeDef.Viewer != "form" {
		t.Fatalf("expected parsed TypeDef carried on the node, got %+v", td.TypeDef)
	}
	if td.Canonical != "widget" {
		t.Fatalf("expected canonical 'widget', got %q", td.Canonical)
	}
}
