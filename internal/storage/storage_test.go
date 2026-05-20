package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func TestPathHelpers(t *testing.T) {
	ws := "/workspace"
	cases := []struct{ fn, want string }{
		{storage.NomosDir(ws), filepath.Join(ws, ".nomos")},
		{storage.CosmosFile(ws), filepath.Join(ws, ".nomos", "cosmos.yaml")},
		{storage.ServicesDir(ws), filepath.Join(ws, ".nomos", "services")},
		{storage.DecisionsDir(ws), filepath.Join(ws, ".nomos", "decisions")},
		{storage.CatalogDir(ws), filepath.Join(ws, ".nomos", "catalog")},
		{storage.ServicegraphsDir(ws), filepath.Join(ws, ".nomos", "servicegraphs")},
		{storage.EvidenceDir(ws), filepath.Join(ws, ".nomos", "evidence")},
		{storage.IndexDir(ws), filepath.Join(ws, ".nomos", "index")},
		{storage.CacheDir(ws), filepath.Join(ws, ".nomos", "cache")},
		{storage.UCIDir(ws), filepath.Join(ws, ".nomos", "uci")},
		{storage.KeysFile(ws), filepath.Join(ws, ".nomos", "keys.yaml")},
		{storage.SelfModelDir(ws), filepath.Join(ws, ".nomos", "services", "nomos-core")},
		{storage.LegacyCosmosFile(ws), filepath.Join(ws, "cosmos.yaml")},
		{storage.LegacyCatalogDir(ws), filepath.Join(ws, "catalog")},
		{storage.LegacyServicegraphsDir(ws), filepath.Join(ws, "servicegraphs")},
		{storage.CatalogServicegraphsDir(ws), filepath.Join(ws, ".nomos", "catalog", "servicegraphs")},
	}
	for _, c := range cases {
		if c.fn != c.want {
			t.Errorf("got %q, want %q", c.fn, c.want)
		}
	}
}

func TestCosmosFileForRead_Canonical(t *testing.T) {
	dir := t.TempDir()
	must(t, os.MkdirAll(storage.NomosDir(dir), 0o755))
	must(t, os.WriteFile(storage.CosmosFile(dir), []byte("id: test\n"), 0o644))
	got, legacy := storage.CosmosFileForRead(dir)
	if got != storage.CosmosFile(dir) {
		t.Errorf("expected canonical path, got %q", got)
	}
	if legacy {
		t.Error("expected legacy=false for canonical layout")
	}
}

func TestCosmosFileForRead_Legacy(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "cosmos.yaml"), []byte("id: test\n"), 0o644))
	got, legacy := storage.CosmosFileForRead(dir)
	if got != storage.LegacyCosmosFile(dir) {
		t.Errorf("expected legacy path, got %q", got)
	}
	if !legacy {
		t.Error("expected legacy=true for legacy layout")
	}
}

func TestCosmosFileForRead_Neither(t *testing.T) {
	dir := t.TempDir()
	got, legacy := storage.CosmosFileForRead(dir)
	if got != storage.CosmosFile(dir) {
		t.Errorf("expected canonical fallback, got %q", got)
	}
	if legacy {
		t.Error("expected legacy=false when no file exists")
	}
}

func TestCatalogDirForRead(t *testing.T) {
	dir := t.TempDir()
	must(t, os.MkdirAll(storage.CatalogDir(dir), 0o755))
	if got := storage.CatalogDirForRead(dir); got != storage.CatalogDir(dir) {
		t.Errorf("expected canonical, got %q", got)
	}
}

func TestCatalogDirForRead_Legacy(t *testing.T) {
	dir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(dir, "catalog"), 0o755))
	if got := storage.CatalogDirForRead(dir); got != storage.LegacyCatalogDir(dir) {
		t.Errorf("expected legacy catalog, got %q", got)
	}
}

func TestCatalogServicegraphsDirForRead(t *testing.T) {
	dir := t.TempDir()
	must(t, os.MkdirAll(storage.CatalogDir(dir), 0o755))
	want := filepath.Join(storage.CatalogDir(dir), "servicegraphs")
	if got := storage.CatalogServicegraphsDirForRead(dir); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLegacyLayoutDetected(t *testing.T) {
	dir := t.TempDir()
	// no files: not legacy
	if storage.LegacyLayoutDetected(dir) {
		t.Error("expected false when no files exist")
	}
	// canonical present: not legacy
	must(t, os.MkdirAll(storage.NomosDir(dir), 0o755))
	must(t, os.WriteFile(storage.CosmosFile(dir), []byte(""), 0o644))
	if storage.LegacyLayoutDetected(dir) {
		t.Error("expected false with canonical layout")
	}
}

func TestLegacyLayoutDetected_True(t *testing.T) {
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "cosmos.yaml"), []byte(""), 0o644))
	if !storage.LegacyLayoutDetected(dir) {
		t.Error("expected true for legacy layout")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
