package idmigrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/nomos/nomos/internal/idgen"
)

// Plan ist eine Liste geplanter ID-Umbenennungen.
type Plan struct {
	Steps []PlanStep
}

// PlanStep beschreibt eine einzelne Umbenennung.
type PlanStep struct {
	Path         string
	ArtefactType string
	OldID        string
	NewID        string
	// NewPath ist der neue Dateipfad, falls die Datei umbenannt werden soll
	// (Dateiname enthielt die alte ID). Leer, wenn keine Umbenennung nötig.
	NewPath string
}

// BuildPlan erzeugt einen Migrationsplan aus den Scan-Kandidaten.
// Für jeden Kandidaten wird eine neue ID per idgen erzeugt.
func BuildPlan(candidates []Candidate) (Plan, error) {
	var plan Plan
	for _, c := range candidates {
		newID, err := generateForType(c.ArtefactType)
		if err != nil {
			return Plan{}, fmt.Errorf("%s: %w", c.Path, err)
		}
		step := PlanStep{
			Path:         c.Path,
			ArtefactType: c.ArtefactType,
			OldID:        c.OldID,
			NewID:        newID,
		}
		step.NewPath = renamedPath(c.Path, c.OldID, newID)
		plan.Steps = append(plan.Steps, step)
	}
	return plan, nil
}

// generateForType fällt auf ein Default-Präfix zurück, wenn der Artefakttyp
// unbekannt ist (z. B. ältere Cosmos-Versionen ohne type-Feld).
func generateForType(artefactType string) (string, error) {
	if artefactType == "" || artefactType == "cosmos" {
		// Cosmos wird nicht migriert; Default-Präfix BLP für unbekannte Typen.
		if artefactType == "cosmos" {
			return idgen.CosmosRootID, nil
		}
		return idgen.New("BLP")
	}
	id, err := idgen.NewForType(artefactType)
	if err == nil {
		return id, nil
	}
	// Unbekannter Typ → generischen Präfix vergeben, damit Migration nie abbricht.
	return idgen.New("BLP")
}

// renamedPath schlägt einen neuen Dateipfad vor, wenn der alte Dateiname
// die alte ID enthielt (verbreitetes Muster: <id>.yaml).
// Returns leer, wenn keine Umbenennung nötig ist.
func renamedPath(p, oldID, newID string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	// Exact-Match: stem == oldID → rename zu newID.
	if stem == oldID {
		return filepath.Join(filepath.Dir(p), newID+ext)
	}
	return ""
}

// Apply führt den Plan aus: schreibt die YAML-Dateien mit neuer ID,
// benennt Dateien um, speichert History-Einträge.
// dryRun=true: keine Änderungen am Dateisystem, nur Berichten.
func Apply(workspace string, plan Plan, dryRun bool) (History, error) {
	history, err := LoadHistory(workspace)
	if err != nil {
		return History{}, err
	}
	// Mapping old→new für Cross-Reference-Pass aufbauen.
	mapping := make(map[string]string, len(plan.Steps))
	for _, s := range plan.Steps {
		mapping[s.OldID] = s.NewID
	}

	// 1) Top-Level id:-Feld + Cross-References in jedem Artefakt rewriten.
	for _, s := range plan.Steps {
		if dryRun {
			continue
		}
		if err := rewriteFile(s.Path, s.OldID, s.NewID, mapping); err != nil {
			return history, fmt.Errorf("rewrite %s: %w", s.Path, err)
		}
		if s.NewPath != "" && s.NewPath != s.Path {
			if err := os.Rename(s.Path, s.NewPath); err != nil {
				return history, fmt.Errorf("rename %s → %s: %w", s.Path, s.NewPath, err)
			}
		}
	}

	// 2) Cross-References in *anderen* Dateien aktualisieren (Dateien, die
	// nicht selbst migriert wurden, aber auf migrierte IDs verweisen).
	if !dryRun {
		if err := rewriteCrossRefs(workspace, mapping, plan.Steps); err != nil {
			return history, err
		}
	}

	// 3) History-Einträge schreiben.
	now := time.Now().UTC().Format(time.RFC3339)
	for _, s := range plan.Steps {
		history.Entries = append(history.Entries, HistoryEntry{
			OldID:        s.OldID,
			NewID:        s.NewID,
			ArtefactType: s.ArtefactType,
			SourcePath:   relativeToWorkspace(workspace, s.Path),
			MigratedAt:   now,
			Reason:       "ADR-0020 migration",
		})
	}
	sort.SliceStable(history.Entries, func(i, j int) bool {
		return history.Entries[i].OldID < history.Entries[j].OldID
	})
	if dryRun {
		return history, nil
	}
	if err := SaveHistory(workspace, history); err != nil {
		return history, err
	}
	Invalidate(workspace)
	return history, nil
}

// rewriteFile rewriten die YAML-Datei: top-level id-Feld und alle Cross-
// References. Wir laden via yaml.Node, weil die Strukturen heterogen sind
// und ein typsicheres Round-Trip alle Typen kennen müsste.
func rewriteFile(path, oldID, newID string, mapping map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}
	if root.Kind == 0 {
		return nil // leere Datei
	}
	rewriteNode(&root, oldID, newID, mapping)
	out, err := yaml.Marshal(&root)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// rewriteNode geht den YAML-Baum durch und ersetzt jeden skalaren Wert,
// dessen Inhalt exakt einer alten ID im mapping entspricht. Da Legacy-IDs
// (z. B. "PROD-ACC-MBX-001", "SB-MAILBOX-001") syntaktisch sehr distinct
// sind, ist das Risiko von Fehltreffern in Freitext praktisch null —
// und wenn doch ein Beschreibungstext auf eine ID verweist, ist die
// Ersetzung sogar gewollt.
//
// oldID/newID sind keine zusätzlichen Eingaben mehr (im mapping enthalten);
// die Parameter bleiben für API-Stabilität.
func rewriteNode(n *yaml.Node, _, _ string, mapping map[string]string) {
	if n == nil {
		return
	}
	switch n.Kind {
	case yaml.DocumentNode, yaml.MappingNode, yaml.SequenceNode:
		for _, c := range n.Content {
			rewriteNode(c, "", "", mapping)
		}
	case yaml.ScalarNode:
		if newID, ok := mapping[strings.TrimSpace(n.Value)]; ok {
			n.Value = newID
		}
	}
}

// rewriteCrossRefs läuft über alle YAML-Dateien im Cosmos und ersetzt
// Cross-References anhand des mapping. Dateien aus dem Plan werden
// uebersprungen, weil sie schon rewritten wurden — wir muessen aber
// pruefen ob die neue Datei (NewPath) existiert.
func rewriteCrossRefs(workspace string, mapping map[string]string, skipSteps []PlanStep) error {
	skip := make(map[string]bool, len(skipSteps)*2)
	for _, s := range skipSteps {
		skip[s.Path] = true
		if s.NewPath != "" {
			skip[s.NewPath] = true
		}
	}
	return filepath.WalkDir(workspace, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// `.nomos/cache` und `.nomos/index` koennen lokal generiert sein —
			// nicht migrieren, sonst kollidieren wir mit Regenerationen.
			base := d.Name()
			if base == "cache" || base == "index" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isYAML(p) {
			return nil
		}
		if skip[p] {
			return nil
		}
		if err := rewriteFile(p, "", "", mapping); err != nil {
			// Defekte / nicht-Map-YAML-Dateien (z.B. fragmentartige Daten)
			// stoppen die Migration nicht.
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return nil
		}
		return nil
	})
}

func relativeToWorkspace(workspace, p string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return p
	}
	rel, err := filepath.Rel(abs, p)
	if err != nil {
		return p
	}
	return rel
}
