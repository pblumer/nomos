package selfmodel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func makeWorkspace(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	if err := os.MkdirAll(storage.NomosDir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(storage.CosmosFile(p), []byte("id: test-cosmos\ntype: cosmos\nname: Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadBundleMeta(t *testing.T) {
	meta, err := LoadBundleMeta()
	if err != nil {
		t.Fatalf("LoadBundleMeta: %v", err)
	}
	if meta.ID == "" {
		t.Error("expected non-empty bundle ID")
	}
	if meta.Version == "" {
		t.Error("expected non-empty bundle version")
	}
	if len(meta.Manifest) == 0 {
		t.Error("expected non-empty manifest")
	}
}

func TestBinaryChecksum(t *testing.T) {
	cs, err := BinaryChecksum()
	if err != nil {
		t.Fatalf("BinaryChecksum: %v", err)
	}
	if cs == "" {
		t.Error("expected non-empty checksum")
	}
	// Deterministic: calling twice returns same value.
	cs2, err := BinaryChecksum()
	if err != nil {
		t.Fatalf("BinaryChecksum (2nd): %v", err)
	}
	if cs != cs2 {
		t.Errorf("checksum not deterministic: %q vs %q", cs, cs2)
	}
}

func TestStatusCaseA(t *testing.T) {
	// No SelfModel block in cosmos.yaml → Case A
	p := makeWorkspace(t)
	st, err := Status(p)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Case != "A" {
		t.Errorf("expected case A, got %q", st.Case)
	}
	if !strings.Contains(st.Message, "not yet imported") {
		t.Errorf("unexpected message: %q", st.Message)
	}
}

func TestImportAndStatusCaseB(t *testing.T) {
	p := makeWorkspace(t)

	if err := Import(p, false); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// After import, status should be Case B (up-to-date).
	st, err := Status(p)
	if err != nil {
		t.Fatalf("Status after import: %v", err)
	}
	if st.Case != "B" {
		t.Errorf("expected case B, got %q (msg=%s)", st.Case, st.Message)
	}
	if st.BinaryVersion == "" {
		t.Error("expected BinaryVersion set")
	}
	if st.BinaryChecksum == "" {
		t.Error("expected BinaryChecksum set")
	}
}

func TestImportIdempotent(t *testing.T) {
	// Second Import (forceOverwrite=false) with case B should be a no-op.
	p := makeWorkspace(t)
	if err := Import(p, false); err != nil {
		t.Fatalf("first Import: %v", err)
	}
	if err := Import(p, false); err != nil {
		t.Fatalf("second Import (no-op): %v", err)
	}
}

func TestImportCaseDBlocksWithoutForce(t *testing.T) {
	p := makeWorkspace(t)
	if err := Import(p, false); err != nil {
		t.Fatalf("initial Import: %v", err)
	}

	// Corrupt a workspace file to trigger Case D.
	meta, _ := LoadBundleMeta()
	if len(meta.Manifest) == 0 {
		t.Skip("no manifest files")
	}
	dst := filepath.Join(storage.NomosDir(p), filepath.FromSlash(meta.Manifest[0]))
	if err := os.WriteFile(dst, []byte("# modified\n"), 0o644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	// Without force: should error.
	if err := Import(p, false); err == nil {
		t.Error("expected error for case D without forceOverwrite")
	}

	// With force: should succeed.
	if err := Import(p, true); err != nil {
		t.Errorf("Import with forceOverwrite: %v", err)
	}
}

func TestChecksumFS(t *testing.T) {
	cs, err := checksumFS(bundleFS, "bundle")
	if err != nil {
		t.Fatalf("checksumFS: %v", err)
	}
	if cs == "" {
		t.Error("expected non-empty checksum")
	}
}

func TestChecksumWorkspace(t *testing.T) {
	p := makeWorkspace(t)
	if err := Import(p, false); err != nil {
		t.Fatalf("Import: %v", err)
	}
	meta, _ := LoadBundleMeta()
	cs, err := checksumWorkspace(p, meta.Manifest)
	if err != nil {
		t.Fatalf("checksumWorkspace: %v", err)
	}
	// Should equal the binary checksum after a clean import.
	bcs, _ := BinaryChecksum()
	if cs != bcs {
		t.Errorf("workspace checksum %q != binary checksum %q", cs, bcs)
	}
}

func TestStatusMissingCosmosFile(t *testing.T) {
	// Status on a dir with no cosmos.yaml → Case A (graceful)
	p := t.TempDir()
	st, err := Status(p)
	if err != nil {
		t.Fatalf("Status on missing file: %v", err)
	}
	if st.Case != "A" {
		t.Errorf("expected case A, got %q", st.Case)
	}
}
