package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

func executeCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newRoot()
	out := &strings.Builder{}
	errOut := &strings.Builder{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestVersionCommandTextOutput(t *testing.T) {
	out, _, err := executeCommand(t, "version")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	for _, s := range []string{"nomos version", "commit:", "date:", "dirty:", "builtBy:", "go:", "os/arch:"} {
		if !strings.Contains(out, s) {
			t.Fatalf("output missing %q: %s", s, out)
		}
	}
}

func TestVersionCommandShortOutput(t *testing.T) {
	out, _, err := executeCommand(t, "version", "--short")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		t.Fatal("expected short output")
	}
	if strings.Contains(trimmed, "commit:") || strings.Contains(trimmed, "go:") {
		t.Fatalf("expected only version output, got %q", trimmed)
	}
}

func TestVersionCommandJSONOutput(t *testing.T) {
	out, _, err := executeCommand(t, "version", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v (%s)", err, out)
	}
	if got["name"] != "nomos" {
		t.Fatalf("unexpected name: %v", got["name"])
	}
	if got["version"] == "" {
		t.Fatalf("missing version: %v", got)
	}
	for _, key := range []string{"commit", "date", "dirty", "builtBy", "go", "os", "arch"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("missing key %q in %v", key, got)
		}
	}
}

func TestVersionCommandExplicitTextFormat(t *testing.T) {
	out, _, err := executeCommand(t, "version", "--format", "text")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(out, "nomos version") || !strings.Contains(out, "os/arch:") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestVersionCommandRejectsInvalidFormat(t *testing.T) {
	_, _, err := executeCommand(t, "version", "--format", "xml")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "format") && !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersionCommandShortOverridesFormat(t *testing.T) {
	out, _, err := executeCommand(t, "version", "--short", "--format", "json")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		t.Fatal("expected output")
	}
	if strings.Contains(trimmed, "{") || strings.Contains(trimmed, "commit") {
		t.Fatalf("expected short output to override format, got %q", trimmed)
	}
}

func TestCosmosInitCreatesExpectedStructure(t *testing.T) {
	p := t.TempDir()
	_, _, err := executeCommand(t, "cosmos", "init", p)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	for _, rel := range []string{"cosmos.yaml", "README.md", "domains", ".nomos/cache", ".nomos/index", ".nomos/evidence"} {
		if _, err := os.Stat(filepath.Join(p, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	var c model.Cosmos
	if err := fsx.ReadYAML(filepath.Join(p, "cosmos.yaml"), &c); err != nil {
		t.Fatal(err)
	}
	if c.ID != "cosmos-local" || c.Type != "cosmos" || c.Version != "0.1.0" || c.Status != "draft" {
		t.Fatalf("unexpected cosmos: %+v", c)
	}
}

func TestCosmosInitWithGitInitializesGitRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	p := t.TempDir()
	_, _, err := executeCommand(t, "cosmos", "init", p, "--git")
	if err != nil {
		t.Fatalf("init --git failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, ".git")); err != nil {
		t.Fatalf(".git missing: %v", err)
	}
	out, _, err := executeCommand(t, "cosmos", "doctor", "--path", p)
	if err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	if !strings.Contains(out, "OK Git Repository gefunden") {
		t.Fatalf("unexpected doctor output: %s", out)
	}
}

func TestCosmosInfoReadsCosmosYaml(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	out, _, err := executeCommand(t, "cosmos", "info", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"id=cosmos-local", "name=Local Cosmos", "version=0.1.0", "status=draft"} {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in %q", s, out)
		}
	}
}

func TestCosmosDoctorFailsWhenCosmosYamlMissing(t *testing.T) {
	p := t.TempDir()
	out, errOut, err := executeCommand(t, "cosmos", "doctor", "--path", p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(out+errOut+err.Error(), "cosmos.yaml") {
		t.Fatalf("missing cosmos.yaml hint: %s %s %v", out, errOut, err)
	}
}

func TestDomainAddCreatesDomainArtifact(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, err := executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"domains/identity.blumer.cloud/domain.yaml", "domains/identity.blumer.cloud/README.md"} {
		if _, err := os.Stat(filepath.Join(p, rel)); err != nil {
			t.Fatalf("missing %s", rel)
		}
	}
	var d model.Domain
	if err := fsx.ReadYAML(filepath.Join(p, "domains/identity.blumer.cloud/domain.yaml"), &d); err != nil {
		t.Fatal(err)
	}
	if d.Type != "domain" || d.Name != "identity.blumer.cloud" || d.Owner != "Identity Team" || d.DNSName != "identity.blumer.cloud" {
		t.Fatalf("unexpected domain: %+v", d)
	}
}

func TestDomainAddRejectsInvalidDNSName(t *testing.T) {
	p := t.TempDir()
	_, _, err := executeCommand(t, "domain", "add", "invalid", "--path", p)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDomainListSupportsPathFlag(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	out, _, err := executeCommand(t, "domain", "list", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "identity.blumer.cloud") {
		t.Fatalf("unexpected out: %s", out)
	}
}

func TestServiceAddCreatesServiceArtifact(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	_, _, err := executeCommand(t, "service", "add", "user-account", "--domain", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	if err != nil {
		t.Fatal(err)
	}
	base := "domains/identity.blumer.cloud/services/user-account"
	for _, rel := range []string{base + "/service.yaml", base + "/README.md", base + "/capabilities", base + "/requirements", base + "/rules", base + "/processes", base + "/skills", base + "/findings", base + "/evidence"} {
		if _, err := os.Stat(filepath.Join(p, rel)); err != nil {
			t.Fatalf("missing %s", rel)
		}
	}
}

func TestServiceAddRequiresDomainFlag(t *testing.T) {
	p := t.TempDir()
	_, _, err := executeCommand(t, "service", "add", "user-account", "--path", p)
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestValidateSupportsPathFlag(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	out, _, err := executeCommand(t, "validate", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Nomos Validierung") {
		t.Fatal(out)
	}
}

func TestValidateSupportsJSONFormat(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	out, _, err := executeCommand(t, "validate", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid json: %v (%s)", err, out)
	}
	if got["status"] != "ok" {
		t.Fatalf("unexpected status: %v", got["status"])
	}
	if _, ok := got["findings"]; !ok {
		t.Fatal("missing findings")
	}
}

func TestValidateReturnsErrorWhenCosmosYamlMissing(t *testing.T) {
	p := t.TempDir()
	out, errOut, err := executeCommand(t, "validate", "--path", p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(out+errOut+err.Error(), "COSMOS_MISSING") && !strings.Contains(out+errOut+err.Error(), "cosmos.yaml fehlt") {
		t.Fatal("missing marker")
	}
}
func TestValidateRejectsInvalidFormat(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, err := executeCommand(t, "validate", "--path", p, "--format", "xml")
	if err == nil {
		t.Fatal("expected err")
	}
}
func TestGraphSupportsPathFlag(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	out, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "graph TD") {
		t.Fatal(out)
	}
}

func TestServeMuxHealthEndpoint(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	newServeMux(".").ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) || !strings.Contains(rr.Body.String(), `"service":"nomos"`) {
		t.Fatal(rr.Body.String())
	}
}
func TestServeMuxValidateEndpoint(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/validate", nil)
	newServeMux(".").ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Fatal(rr.Body.String())
	}
}
func TestServeCommandRegistersPathFlag(t *testing.T) {
	if f := serveCmd().Flags().Lookup("path"); f == nil {
		t.Fatal("missing path flag")
	}
}

func TestRootCommandContainsExpectedSubcommands(t *testing.T) {
	root := newRoot()
	want := []string{"version", "cosmos", "domain", "service", "validate", "graph", "verify", "serve"}
	for _, w := range want {
		if _, _, err := root.Find([]string{w}); err != nil {
			t.Fatalf("missing subcommand %s", w)
		}
	}
}
