package idmigrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/idgen"
	"github.com/nomos/nomos/internal/storage"
)

// setupCosmos erzeugt ein minimales Cosmos-Layout mit Legacy-Artefakten.
func setupCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "products"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "blueprints", "services"), 0o755))
	must(os.MkdirAll(filepath.Join(storage.CatalogDir(p), "instances", "products"), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: COS_root\ntype: cosmos\nname: T\n"), 0o644))

	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "blueprints", "services", "SB-MAIL-001.yaml"),
		[]byte("id: SB-MAIL-001\ntype: service_blueprint\nname: Mailbox\nversion: 0.1.0\nstatus: draft\nowner: T\n"),
		0o644,
	))
	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "blueprints", "products", "PROD-MAIL-001.yaml"),
		[]byte("id: PROD-MAIL-001\ntype: product_blueprint\nname: Mail Product\nversion: 0.1.0\nstatus: draft\nowner: T\nrequired_service_blueprints:\n  - SB-MAIL-001\n"),
		0o644,
	))
	must(os.WriteFile(
		filepath.Join(storage.CatalogDir(p), "instances", "products", "PI-MAIL-001.yaml"),
		[]byte("id: PI-MAIL-001\ntype: product_instance\nname: My Mail\nblueprint_ref: PROD-MAIL-001\nblueprint_version: 0.1.0\ncompliance_status: unknown\n"),
		0o644,
	))
	return p
}

func TestScanFindsLegacyIDs(t *testing.T) {
	p := setupCosmos(t)
	cands, err := Scan(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 3 {
		t.Fatalf("expected 3 candidates, got %d: %#v", len(cands), cands)
	}
	wantTypes := map[string]bool{"service_blueprint": false, "product_blueprint": false, "product_instance": false}
	for _, c := range cands {
		if !idgen.IsLegacy(c.OldID) {
			t.Errorf("expected %q to be legacy", c.OldID)
		}
		if _, ok := wantTypes[c.ArtefactType]; ok {
			wantTypes[c.ArtefactType] = true
		}
	}
	for typ, seen := range wantTypes {
		if !seen {
			t.Errorf("expected to see type %s in candidates", typ)
		}
	}
}

func TestBuildPlanGeneratesNewIDs(t *testing.T) {
	p := setupCosmos(t)
	cands, _ := Scan(p)
	plan, err := BuildPlan(cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != len(cands) {
		t.Fatalf("plan steps mismatch: %d vs %d", len(plan.Steps), len(cands))
	}
	for _, s := range plan.Steps {
		if !idgen.IsValidForType(s.NewID, s.ArtefactType) {
			t.Errorf("step %s -> %s: new ID not valid for type %s", s.OldID, s.NewID, s.ArtefactType)
		}
		if s.NewPath == "" {
			t.Errorf("step %s: expected file rename plan since stem == OldID", s.OldID)
		}
	}
}

func TestApplyRewritesIDsAndReferences(t *testing.T) {
	p := setupCosmos(t)
	cands, _ := Scan(p)
	plan, _ := BuildPlan(cands)

	if _, err := Apply(p, plan, false); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Alte Dateinamen duerfen nicht mehr existieren, neue Dateinamen schon.
	for _, s := range plan.Steps {
		if _, err := os.Stat(s.Path); !os.IsNotExist(err) {
			t.Errorf("expected old file %s to be removed", s.Path)
		}
		if _, err := os.Stat(s.NewPath); err != nil {
			t.Errorf("expected new file %s to exist: %v", s.NewPath, err)
		}
	}

	// In jeder neuen Datei muss die ID dem ADR-0020 entsprechen.
	for _, s := range plan.Steps {
		data, err := os.ReadFile(s.NewPath)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.Contains(content, "id: "+s.NewID) {
			t.Errorf("file %s missing new id %s; content:\n%s", s.NewPath, s.NewID, content)
		}
		if strings.Contains(content, s.OldID) {
			t.Errorf("file %s still references old id %s; content:\n%s", s.NewPath, s.OldID, content)
		}
	}

	// History muss alle Eintraege enthalten.
	hist, err := LoadHistory(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist.Entries) != len(plan.Steps) {
		t.Fatalf("history len = %d; want %d", len(hist.Entries), len(plan.Steps))
	}

	// Resolve sollte alte IDs auf neue mappen.
	for _, s := range plan.Steps {
		got, was := Resolve(p, s.OldID)
		if !was {
			t.Errorf("Resolve(%s) wasLegacy=false; want true", s.OldID)
		}
		if got != s.NewID {
			t.Errorf("Resolve(%s) = %s; want %s", s.OldID, got, s.NewID)
		}
	}
}

func TestResolveLeavesUnknownIDsUnchanged(t *testing.T) {
	p := setupCosmos(t)
	got, was := Resolve(p, "PRD_A7K3M2")
	if got != "PRD_A7K3M2" || was {
		t.Errorf("Resolve unknown new-format ID = %s,%v; want PRD_A7K3M2,false", got, was)
	}
	got, was = Resolve(p, "")
	if got != "" || was {
		t.Errorf("Resolve empty = %s,%v; want \"\",false", got, was)
	}
}

func TestApplyDryRunNoChanges(t *testing.T) {
	p := setupCosmos(t)
	cands, _ := Scan(p)
	plan, _ := BuildPlan(cands)
	if _, err := Apply(p, plan, true); err != nil {
		t.Fatal(err)
	}
	// Originale Dateien sollten unangetastet sein.
	for _, c := range cands {
		if _, err := os.Stat(c.Path); err != nil {
			t.Errorf("dry-run removed file %s", c.Path)
		}
	}
	// Keine History geschrieben.
	if _, err := os.Stat(HistoryFile(p)); !os.IsNotExist(err) {
		t.Errorf("dry-run wrote history file")
	}
}

func TestHistoryAppendIdempotent(t *testing.T) {
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(storage.NomosDir(p), 0o755))

	e := HistoryEntry{OldID: "PROD-1", NewID: "PRD_AAAAAA"}
	must(AppendEntry(p, e))
	must(AppendEntry(p, e))
	h, err := LoadHistory(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Entries) != 1 {
		t.Errorf("expected idempotent append, got %d entries", len(h.Entries))
	}
}
