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
}

func TestSaveAndLoadTypeForm(t *testing.T) {
	p := createAppTestCosmos(t)

	// No form authored yet.
	if _, ok, err := LoadTypeForm(p, "task"); err != nil || ok {
		t.Fatalf("expected no form, got ok=%v err=%v", ok, err)
	}

	schema := map[string]any{"type": "default", "components": []any{map[string]any{"type": "textfield", "key": "title"}}}
	if _, err := SaveTypeForm(p, "task", model.View{Schema: schema}); err != nil {
		t.Fatalf("SaveTypeForm: %v", err)
	}
	// Written to .nomos/views/<id>_new.frm with the default engine.
	if _, err := os.Stat(filepath.Join(storage.ViewsDir(p), "task_new.frm")); err != nil {
		t.Fatalf("form file not created: %v", err)
	}
	v, ok, err := LoadTypeForm(p, "task")
	if err != nil || !ok {
		t.Fatalf("LoadTypeForm: ok=%v err=%v", ok, err)
	}
	if v.Engine != "form-js" {
		t.Fatalf("expected default engine form-js, got %q", v.Engine)
	}
	if v.Schema["type"] != "default" {
		t.Fatalf("schema not round-tripped: %+v", v.Schema)
	}

	// Invalid ids are rejected.
	if _, err := SaveTypeForm(p, "Not Valid", model.View{}); err == nil {
		t.Fatal("expected rejection of invalid type id")
	}
}

func TestSaveTypeDefSeedsStandardForms(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := SaveTypeDef(p, model.TypeDef{
		ID:    "risk",
		Label: "Risk",
		Properties: []model.TypeProperty{
			{Name: "title", Type: "string", Required: true},
		},
	}); err != nil {
		t.Fatalf("SaveTypeDef: %v", err)
	}
	// All four standard view forms are generated.
	for _, v := range []string{"new", "edit", "list", "short"} {
		if _, err := os.Stat(filepath.Join(storage.ViewsDir(p), "risk_"+v+".frm")); err != nil {
			t.Errorf("expected seeded risk_%s.frm: %v", v, err)
		}
		view, ok, err := LoadTypeFormVariant(p, "risk", v)
		if err != nil || !ok {
			t.Errorf("LoadTypeFormVariant(%s): ok=%v err=%v", v, ok, err)
			continue
		}
		if view.Engine != "form-js" {
			t.Errorf("variant %s engine = %q, want form-js", v, view.Engine)
		}
	}

	// Re-saving must not clobber a form the user has edited.
	custom := map[string]any{"type": "default", "components": []any{map[string]any{"type": "textfield", "key": "custom"}}}
	if _, err := SaveTypeFormVariant(p, "risk", "edit", model.View{Schema: custom}); err != nil {
		t.Fatalf("SaveTypeFormVariant: %v", err)
	}
	if _, err := SaveTypeDef(p, model.TypeDef{ID: "risk", Label: "Risk v2"}); err != nil {
		t.Fatalf("SaveTypeDef (re-save): %v", err)
	}
	v, _, err := LoadTypeFormVariant(p, "risk", "edit")
	if err != nil {
		t.Fatalf("LoadTypeFormVariant after re-save: %v", err)
	}
	comps, _ := v.Schema["components"].([]any)
	if len(comps) != 1 {
		t.Fatalf("authored edit form was clobbered on re-save: %+v", v.Schema)
	}

	// Unknown variants are rejected.
	if _, _, err := LoadTypeFormVariant(p, "risk", "bogus"); err == nil {
		t.Error("expected rejection of unknown form variant")
	}

	// Deleting the type removes all four forms.
	if err := DeleteTypeDef(p, "risk"); err != nil {
		t.Fatalf("DeleteTypeDef: %v", err)
	}
	for _, v := range []string{"new", "edit", "list", "short"} {
		if _, err := os.Stat(filepath.Join(storage.ViewsDir(p), "risk_"+v+".frm")); !os.IsNotExist(err) {
			t.Errorf("expected risk_%s.frm removed, stat err = %v", v, err)
		}
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
	types := findChild(nomos.Children, "folder", "types")
	if types == nil {
		t.Fatal("expected types/ folder under .nomos")
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
