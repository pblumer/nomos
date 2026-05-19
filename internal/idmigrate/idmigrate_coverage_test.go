package idmigrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

// ---------------------------------------------------------------------------
// AllLegacy
// ---------------------------------------------------------------------------

func TestAllLegacy(t *testing.T) {
	p := setupCosmos(t)
	cands, _ := Scan(p)
	plan, _ := BuildPlan(cands)
	Apply(p, plan, false) //nolint

	all := AllLegacy(p)
	if len(all) == 0 {
		t.Error("AllLegacy returned empty list after migration")
	}
}

func TestAllLegacy_NoHistory(t *testing.T) {
	p := t.TempDir()
	// No history file → should return empty (nil or zero-length)
	all := AllLegacy(p)
	if len(all) != 0 {
		t.Errorf("expected empty, got %v", all)
	}
}

// ---------------------------------------------------------------------------
// inferTypeFromPath
// ---------------------------------------------------------------------------

func TestInferTypeFromPath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"catalog/blueprints/products/my.yaml", "product_blueprint"},
		{"catalog/blueprints/services/my.yaml", "service_blueprint"},
		{"catalog/instances/products/my.yaml", "product_instance"},
		{"catalog/instances/services/my.yaml", "service_instance"},
		{"catalog/servicegraphs/my.yaml", "servicegraph"},
		{"domains/identity.blumer.cloud/decisions/DEC-001/decision.yaml", "decision"},
		{"catalog/requirements/req.yaml", "requirement"},
		{"catalog/rules/rule.yaml", "rule"},
		{"catalog/products/old.yaml", "product"},
		{"catalog/services/old.yaml", "service"},
		{"unknown/path/file.yaml", ""},
	}
	for _, c := range cases {
		got := inferTypeFromPath(c.path)
		if got != c.want {
			t.Errorf("inferTypeFromPath(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// generateForType — cosmos and unknown type
// ---------------------------------------------------------------------------

func TestGenerateForType(t *testing.T) {
	id, err := generateForType("cosmos")
	if err != nil {
		t.Fatalf("generateForType(cosmos): %v", err)
	}
	if id == "" {
		t.Error("expected non-empty ID for cosmos type")
	}

	id2, err := generateForType("")
	if err != nil {
		t.Fatalf("generateForType(''): %v", err)
	}
	if id2 == "" {
		t.Error("expected non-empty ID for empty type")
	}
}

// ---------------------------------------------------------------------------
// SaveHistory / LoadHistory round-trip
// ---------------------------------------------------------------------------

func TestSaveLoadHistory(t *testing.T) {
	p := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(storage.NomosDir(p), 0o755))

	entries := []HistoryEntry{
		{OldID: "PROD-1", NewID: "PRD_A00001", SourcePath: "products/p1.yaml", ArtefactType: "product_blueprint"},
		{OldID: "SB-1", NewID: "SBP_B00001", SourcePath: "services/s1.yaml", ArtefactType: "service_blueprint"},
	}
	h := History{Entries: entries}
	if err := SaveHistory(p, h); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadHistory(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
}

// ---------------------------------------------------------------------------
// rewriteCrossRefs via Apply
// ---------------------------------------------------------------------------

func TestApplyUpdatesProductBlueprintRef(t *testing.T) {
	// Setup a cosmos where an instance references a product_blueprint by legacy ID.
	// Apply should update the reference.
	p := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "instances", "products"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: COS_OK\ntype: cosmos\nname: T\n"), 0o644))

	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "blueprints", "products", "PROD-X.yaml"),
		[]byte("id: PROD-X\ntype: product_blueprint\nname: X\nversion: 0.1.0\nstatus: draft\nowner: T\nrequired_service_blueprints: []\n"),
		0o644,
	))
	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "instances", "products", "PI-Y.yaml"),
		[]byte("id: PI-Y\ntype: product_instance\nname: Y\nblueprint_ref: PROD-X\nblueprint_version: 0.1.0\ncompliance_status: unknown\n"),
		0o644,
	))

	cands, err := Scan(p)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(cands)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(p, plan, false); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Instance file should now reference new blueprint ID.
	for _, s := range plan.Steps {
		if s.ArtefactType == "product_instance" {
			data, err := os.ReadFile(s.NewPath)
			if err != nil {
				t.Fatal(err)
			}
			// Old ref should be gone
			if containsStr(string(data), "PROD-X") {
				t.Errorf("instance still references old ID PROD-X: %s", data)
			}
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > 0 && (s[:len(sub)] == sub || containsStr(s[1:], sub))))
}
