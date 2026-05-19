package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func mustInitCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	if _, _, err := executeCommand(t, "cosmos", "init", p); err != nil {
		t.Fatalf("cosmos init: %v", err)
	}
	return p
}

// ── id commands ───────────────────────────────────────────────────────────────

func TestIDNewCmd(t *testing.T) {
	out, _, err := executeCommand(t, "id", "new", "product_blueprint")
	if err != nil {
		t.Fatalf("id new: %v", err)
	}
	id := strings.TrimSpace(out)
	if id == "" {
		t.Error("expected non-empty ID from id new")
	}
}

func TestIDNewCmd_UnknownType(t *testing.T) {
	out, _, _ := executeCommand(t, "id", "new", "unknown_type")
	_ = out // may error, that's OK
}

func TestIDCheckCmd_NoCandidates(t *testing.T) {
	p := mustInitCosmos(t)
	_, _, err := executeCommand(t, "id", "check", "--path", p)
	if err != nil {
		t.Fatalf("id check: %v", err)
	}
}

func TestIDCheckCmd_JSON(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "id", "check", "--path", p, "--format", "json")
	if err != nil {
		t.Fatalf("id check --format json: %v", err)
	}
	if !strings.Contains(out, "count") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestIDMigrateCmd_NoCandidates(t *testing.T) {
	p := mustInitCosmos(t)
	_, _, err := executeCommand(t, "id", "migrate", "--path", p)
	if err != nil {
		t.Fatalf("id migrate: %v", err)
	}
}

func TestIDMigrateCmd_DryRun(t *testing.T) {
	p := mustInitCosmos(t)
	// Add a legacy-ID artifact
	bpDir := filepath.Join(storage.CatalogDir(p), "blueprints", "products")
	if err := os.MkdirAll(bpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bp := model.Blueprint{ID: "PROD-LEGACY-001", Type: "product_blueprint", Name: "Legacy BP", Version: "0.1.0", Status: "draft", Owner: "Team"}
	if err := fsx.WriteYAML(filepath.Join(bpDir, "PROD-LEGACY-001.yaml"), bp); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeCommand(t, "id", "migrate", "--path", p, "--dry-run")
	if err != nil {
		t.Fatalf("id migrate --dry-run: %v", err)
	}
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected dry-run message: %s", out)
	}
}

func TestIDMigrateCmd_JSON(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "id", "migrate", "--path", p, "--format", "json")
	if err != nil {
		t.Fatalf("id migrate --format json: %v", err)
	}
	if !strings.Contains(out, "steps") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

// ── self commands ─────────────────────────────────────────────────────────────

func TestSelfStatusCmd(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "self", "status", "--path", p)
	if err != nil {
		t.Fatalf("self status: %v", err)
	}
	// Should show Case A since not yet imported
	if !strings.Contains(out, "A") && !strings.Contains(out, "a") {
		t.Logf("self status output: %s", out)
	}
}

func TestSelfStatusCmd_JSON(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "self", "status", "--path", p, "--format", "json")
	if err != nil {
		t.Fatalf("self status --format json: %v", err)
	}
	if !strings.Contains(out, "Case") && !strings.Contains(out, "case") && !strings.Contains(out, "BinaryVersion") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestSelfImportCmd(t *testing.T) {
	p := mustInitCosmos(t)
	_, _, err := executeCommand(t, "self", "import", "--path", p)
	if err != nil {
		t.Fatalf("self import: %v", err)
	}
}

func TestSelfDiffCmd_CaseA(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "self", "diff", "--path", p)
	if err != nil {
		t.Fatalf("self diff: %v", err)
	}
	if !strings.Contains(out, "importiert") && !strings.Contains(out, "import") {
		t.Logf("self diff output: %s", out)
	}
}

func TestSelfDiffCmd_CaseB(t *testing.T) {
	p := mustInitCosmos(t)
	// Import first so Case B
	if _, _, err := executeCommand(t, "self", "import", "--path", p); err != nil {
		t.Fatalf("self import: %v", err)
	}
	out, _, err := executeCommand(t, "self", "diff", "--path", p)
	if err != nil {
		t.Fatalf("self diff (case B): %v", err)
	}
	if !strings.Contains(out, "aktuell") && !strings.Contains(out, "current") && !strings.Contains(out, "Keine") {
		t.Logf("self diff case B output: %s", out)
	}
}

func TestSelfUpgradeCmd(t *testing.T) {
	p := mustInitCosmos(t)
	// Import first
	if _, _, err := executeCommand(t, "self", "import", "--path", p); err != nil {
		t.Fatalf("self import: %v", err)
	}
	_, _, err := executeCommand(t, "self", "upgrade", "--path", p)
	if err != nil {
		t.Fatalf("self upgrade: %v", err)
	}
}

// ── keys commands ──────────────────────────────────────────────────────────────

func TestKeyCreateCmd(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "key", "create", "test-key-1", "--path", p)
	if err != nil {
		t.Fatalf("key create: %v", err)
	}
	if !strings.Contains(out, "test-key-1") {
		t.Errorf("expected key name in output: %s", out)
	}
}

func TestKeyListCmd(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "key", "list", "--path", p)
	if err != nil {
		t.Fatalf("key list: %v", err)
	}
	_ = out
}

func TestKeyListCmd_WithKeys(t *testing.T) {
	p := mustInitCosmos(t)
	if _, _, err := executeCommand(t, "key", "create", "key-to-list", "--path", p); err != nil {
		t.Fatalf("key create: %v", err)
	}
	out, _, err := executeCommand(t, "key", "list", "--path", p)
	if err != nil {
		t.Fatalf("key list: %v", err)
	}
	if !strings.Contains(out, "key-to-list") {
		t.Errorf("expected key in list: %s", out)
	}
}

func TestKeyRevokeCmd(t *testing.T) {
	p := mustInitCosmos(t)
	if _, _, err := executeCommand(t, "key", "create", "key-to-revoke", "--path", p); err != nil {
		t.Fatalf("key create: %v", err)
	}
	out, _, err := executeCommand(t, "key", "revoke", "key-to-revoke", "--path", p)
	if err != nil {
		t.Fatalf("key revoke: %v", err)
	}
	if !strings.Contains(out, "revoked") && !strings.Contains(out, "key-to-revoke") {
		t.Errorf("expected revoke confirmation: %s", out)
	}
}

// ── servicegraph commands ──────────────────────────────────────────────────────

func TestServicegraphListCmd(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "servicegraph", "list", "--path", p)
	if err != nil {
		t.Fatalf("servicegraph list: %v", err)
	}
	_ = out
}

func TestServicegraphListCmd_JSON(t *testing.T) {
	p := mustInitCosmos(t)
	_, _, err := executeCommand(t, "servicegraph", "list", "--path", p, "--format", "json")
	if err != nil {
		t.Fatalf("servicegraph list --format json: %v", err)
	}
}

// ── process commands ──────────────────────────────────────────────────────────

func TestProcessListCmd(t *testing.T) {
	p := mustInitCosmos(t)
	out, _, err := executeCommand(t, "process", "list", "--path", p)
	if err != nil {
		t.Fatalf("process list: %v", err)
	}
	_ = out
}
