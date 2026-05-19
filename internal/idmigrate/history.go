// Package idmigrate verwaltet die Migration von Legacy-IDs auf das
// ADR-0020-Format und bietet eine transparente Alt→Neu-Auflösung.
//
// Persistenz: `.nomos/id-history.yaml` als append-only Mapping.
package idmigrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/storage"
)

// HistoryEntry beschreibt eine einzelne ID-Migration.
type HistoryEntry struct {
	OldID        string `yaml:"old_id"                  json:"old_id"`
	NewID        string `yaml:"new_id"                  json:"new_id"`
	ArtefactType string `yaml:"artefact_type,omitempty" json:"artefact_type,omitempty"`
	SourcePath   string `yaml:"source_path,omitempty"   json:"source_path,omitempty"`
	MigratedAt   string `yaml:"migrated_at"             json:"migrated_at"`
	Reason       string `yaml:"reason,omitempty"        json:"reason,omitempty"`
}

// History ist das persistierte Mapping. Es ist append-only — Einträge
// werden nie entfernt, damit externe Konsumenten Legacy-IDs weiter auflösen
// können.
type History struct {
	Version int            `yaml:"version"         json:"version"`
	Entries []HistoryEntry `yaml:"entries"         json:"entries"`
}

// HistoryFile liefert den kanonischen Pfad zur History-Datei.
func HistoryFile(workspace string) string {
	return filepath.Join(storage.NomosDir(workspace), "id-history.yaml")
}

// LoadHistory liest die History. Existiert die Datei nicht, wird eine
// leere History zurückgegeben (kein Fehler).
func LoadHistory(workspace string) (History, error) {
	path := HistoryFile(workspace)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return History{Version: 1, Entries: []HistoryEntry{}}, nil
		}
		return History{}, err
	}
	var h History
	if err := fsx.ReadYAML(path, &h); err != nil {
		return History{}, fmt.Errorf("read id-history: %w", err)
	}
	if h.Version == 0 {
		h.Version = 1
	}
	return h, nil
}

// SaveHistory schreibt die History atomar zurück.
func SaveHistory(workspace string, h History) error {
	if h.Version == 0 {
		h.Version = 1
	}
	dir := storage.NomosDir(workspace)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create .nomos dir: %w", err)
	}
	return fsx.WriteYAML(HistoryFile(workspace), h)
}

// AppendEntry fügt einen Eintrag hinzu und speichert sofort.
// Doppelte old_id→new_id Paare werden ignoriert.
func AppendEntry(workspace string, e HistoryEntry) error {
	h, err := LoadHistory(workspace)
	if err != nil {
		return err
	}
	if e.MigratedAt == "" {
		e.MigratedAt = time.Now().UTC().Format(time.RFC3339)
	}
	for _, ex := range h.Entries {
		if ex.OldID == e.OldID && ex.NewID == e.NewID {
			return nil
		}
	}
	h.Entries = append(h.Entries, e)
	sort.SliceStable(h.Entries, func(i, j int) bool {
		return h.Entries[i].OldID < h.Entries[j].OldID
	})
	return SaveHistory(workspace, h)
}

// ── Transparente Auflösung ──────────────────────────────────────────────────

// Cache vermeidet wiederholtes Einlesen der History je Cosmos-Pfad.
type resolverCache struct {
	mu      sync.RWMutex
	loaded  map[string]map[string]string // workspace → (oldID → newID)
	loadErr map[string]error
}

var globalResolver = &resolverCache{
	loaded:  map[string]map[string]string{},
	loadErr: map[string]error{},
}

// Invalidate entfernt den Cache für einen Workspace. Nach Schreibvorgängen
// auf die History aufrufen.
func Invalidate(workspace string) {
	abs, _ := filepath.Abs(workspace)
	globalResolver.mu.Lock()
	delete(globalResolver.loaded, abs)
	delete(globalResolver.loadErr, abs)
	globalResolver.mu.Unlock()
}

func loadInto(workspace string) (map[string]string, error) {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return nil, err
	}
	globalResolver.mu.RLock()
	if m, ok := globalResolver.loaded[abs]; ok {
		globalResolver.mu.RUnlock()
		return m, nil
	}
	if cerr, ok := globalResolver.loadErr[abs]; ok {
		globalResolver.mu.RUnlock()
		return nil, cerr
	}
	globalResolver.mu.RUnlock()

	h, err := LoadHistory(workspace)
	globalResolver.mu.Lock()
	defer globalResolver.mu.Unlock()
	if err != nil {
		globalResolver.loadErr[abs] = err
		return nil, err
	}
	m := make(map[string]string, len(h.Entries))
	for _, e := range h.Entries {
		// Wenn ein old_id mehrfach migriert wurde (selten), gewinnt der
		// letzte Eintrag; sort.SliceStable in AppendEntry sortiert nach OldID,
		// also iterieren wir und überschreiben — das ist OK.
		m[e.OldID] = e.NewID
	}
	globalResolver.loaded[abs] = m
	return m, nil
}

// Resolve gibt für eine beliebige ID die kanonische Neu-ID zurück.
// Ist die ID bereits im neuen Format oder unbekannt, wird sie unverändert
// zurückgegeben. wasLegacy zeigt an, ob aufgelöst wurde.
func Resolve(workspace, anyID string) (string, bool) {
	anyID = strings.TrimSpace(anyID)
	if anyID == "" {
		return "", false
	}
	m, err := loadInto(workspace)
	if err != nil {
		return anyID, false
	}
	if newID, ok := m[anyID]; ok {
		return newID, true
	}
	return anyID, false
}

// AllLegacy gibt alle bekannten Legacy-IDs zurück (für Listings).
func AllLegacy(workspace string) []string {
	m, err := loadInto(workspace)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
