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
	"github.com/nomos/nomos/internal/storage"
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
	for _, rel := range []string{filepath.Join(".nomos", "cosmos.yaml"), "README.md", filepath.Join(".nomos", "domains"), filepath.Join(".nomos", "catalog"), filepath.Join(".nomos", "cache"), filepath.Join(".nomos", "index"), filepath.Join(".nomos", "evidence"), ".gitignore"} {
		if _, err := os.Stat(filepath.Join(p, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	for _, rel := range []string{"cosmos.yaml", "domains", "catalog"} {
		if _, err := os.Stat(filepath.Join(p, rel)); !os.IsNotExist(err) {
			t.Fatalf("root-level %s should not exist", rel)
		}
	}
	gitignore, err := os.ReadFile(filepath.Join(p, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gitignore), ".nomos/cache/") || !strings.Contains(string(gitignore), ".nomos/index/") || strings.Contains(string(gitignore), "\n.nomos/\n") {
		t.Fatalf("unexpected .gitignore: %s", gitignore)
	}
	var c model.Cosmos
	if err := fsx.ReadYAML(storage.CosmosFile(p), &c); err != nil {
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

func TestCosmosInfoWarnsOnLegacyLayout(t *testing.T) {
	p := t.TempDir()
	if err := os.WriteFile(filepath.Join(p, "cosmos.yaml"), []byte("id: legacy\ntype: cosmos\nname: Legacy\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := executeCommand(t, "cosmos", "info", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "id=legacy") {
		t.Fatalf("expected legacy cosmos to be read, got %q", out)
	}
	if !strings.Contains(errOut, "WARNING legacy Nomos layout detected") {
		t.Fatalf("expected legacy warning, got %q", errOut)
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
	for _, rel := range []string{filepath.Join(".nomos", "domains", "cloud", "blumer", "identity", "domain.yaml"), filepath.Join(".nomos", "domains", "cloud", "blumer", "identity", "README.md")} {
		if _, err := os.Stat(filepath.Join(p, rel)); err != nil {
			t.Fatalf("missing %s", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(p, "domains", "identity.blumer.cloud")); !os.IsNotExist(err) {
		t.Fatalf("root-level domain directory should not exist")
	}
	var d model.Domain
	if err := fsx.ReadYAML(filepath.Join(storage.DomainsDir(p), "cloud", "blumer", "identity", "domain.yaml"), &d); err != nil {
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
	base := filepath.Join(".nomos", "domains", "cloud", "blumer", "identity", "services", "user-account")
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

func TestGraphIncludesCosmosDomainsAndServices(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	_, _, _ = executeCommand(t, "domain", "add", "platform.blumer.cloud", "--path", p, "--owner", "Platform Team")
	_, _, _ = executeCommand(t, "service", "add", "user-account", "--domain", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	_, _, _ = executeCommand(t, "service", "add", "privileged-account", "--domain", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	_, _, _ = executeCommand(t, "service", "add", "rule-validation-api", "--domain", "platform.blumer.cloud", "--path", p, "--owner", "Platform Team")
	out, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"graph TD", "Cosmos:", "Domain: identity.blumer.cloud", "Domain: platform.blumer.cloud", "Service: user-account", "Service: privileged-account", "Service: rule-validation-api", "cosmos", "-->"} {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in output:\n%s", s, out)
		}
	}
}

func TestGraphIncludesProductBlueprintRequiredServices(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "user-account", "--domain", "identity.blumer.cloud", "--path", p)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), `id: PB-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Team
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-IDENTITY-ACCOUNT-001
required_services:
  - service_ref: identity.blumer.cloud/user-account
    service_blueprint_ref: SB-IDENTITY-ACCOUNT-001
    required: true
`)
	out, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"PB-ACC-MBX-001", "identity.blumer.cloud/user-account", "SB-IDENTITY-ACCOUNT-001", "requires", "uses"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in graph:\n%s", want, out)
		}
	}
}

func TestGraphOutputIsDeterministic(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "platform.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "z-service", "--domain", "identity.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "a-service", "--domain", "identity.blumer.cloud", "--path", p)
	out1, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	out2, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if out1 != out2 {
		t.Fatalf("graph output differs between runs\n1:\n%s\n2:\n%s", out1, out2)
	}
	if strings.Index(out1, "Domain: identity.blumer.cloud") > strings.Index(out1, "Domain: platform.blumer.cloud") {
		t.Fatalf("domains are not sorted:\n%s", out1)
	}
	if strings.Index(out1, "Service: a-service") > strings.Index(out1, "Service: z-service") {
		t.Fatalf("services are not sorted:\n%s", out1)
	}
}

func TestGraphFailsWhenCosmosYAMLMissing(t *testing.T) {
	p := t.TempDir()
	_, _, err := executeCommand(t, "graph", "--path", p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cosmos.yaml") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGraphHandlesCosmosWithoutDomains(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p, "--without-self")
	out, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "graph TD") || !strings.Contains(out, "Cosmos:") {
		t.Fatalf("unexpected output: %s", out)
	}
	if strings.Contains(out, "Domain:") || strings.Contains(out, "Service:") {
		t.Fatalf("unexpected domain/service in output: %s", out)
	}
}

func TestGraphHandlesDomainWithoutServices(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p, "--without-self")
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	out, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Domain: identity.blumer.cloud") || !strings.Contains(out, "-->") {
		t.Fatalf("unexpected output: %s", out)
	}
	if strings.Contains(out, "Service:") {
		t.Fatalf("unexpected service output: %s", out)
	}
}

func TestCosmosInfoReportsFilesystemDomainCount(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p, "--without-self")
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "domain", "add", "platform.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "domain", "add", "governance.blumer.cloud", "--path", p)
	out, _, err := executeCommand(t, "cosmos", "info", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "domains=3") {
		t.Fatalf("expected domains=3, got: %s", out)
	}
}

func TestCosmosInfoReportsZeroWhenNoDomains(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p, "--without-self")
	out, _, err := executeCommand(t, "cosmos", "info", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "domains=0") {
		t.Fatalf("expected domains=0, got: %s", out)
	}
}

func TestCosmosInfoIgnoresIncompleteDomainDirectories(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p, "--without-self")
	if err := os.MkdirAll(filepath.Join(storage.DomainsDir(p), "incomplete.example.com"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	out, _, err := executeCommand(t, "cosmos", "info", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "domains=1") {
		t.Fatalf("expected domains=1, got: %s", out)
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
	if !strings.Contains(rr.Body.String(), `"status"`) {
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

func createParityCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p, "--without-self")
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	_, _, _ = executeCommand(t, "domain", "add", "platform.blumer.cloud", "--path", p, "--owner", "Platform Team")
	_, _, _ = executeCommand(t, "service", "add", "user-account", "--domain", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	_, _, _ = executeCommand(t, "service", "add", "rule-validation-api", "--domain", "platform.blumer.cloud", "--path", p, "--owner", "Platform Team")
	return p
}

func TestCLIRESTJSONParity(t *testing.T) {
	p := createParityCosmos(t)
	h := newServeMux(p)
	cliCosmos, _, err := executeCommand(t, "cosmos", "info", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	var c1, c2 map[string]any
	if err := json.Unmarshal([]byte(cliCosmos), &c1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(getCLIHTTP(t, h, "/api/v1/cosmos").Body.Bytes(), &c2); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "name", "version", "status", "owner", "domainCount", "serviceCount"} {
		if c1[key] != c2[key] {
			t.Fatalf("cosmos %s mismatch: %v != %v", key, c1[key], c2[key])
		}
	}

	cliDomains, _, err := executeCommand(t, "domain", "list", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	var d1, d2 map[string][]map[string]any
	if err := json.Unmarshal([]byte(cliDomains), &d1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(getCLIHTTP(t, h, "/api/v1/domains").Body.Bytes(), &d2); err != nil {
		t.Fatal(err)
	}
	if d1["domains"][0]["canonical"] != d2["domains"][0]["canonical"] {
		t.Fatalf("domains mismatch: %v %v", d1, d2)
	}
	if d1["domains"][0]["namespace"].(map[string]any)["displayPath"] != "cloud / blumer / identity" {
		t.Fatalf("missing namespace metadata: %v", d1)
	}

	cliDomain, _, err := executeCommand(t, "domain", "get", "identity.blumer.cloud", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	var gd1, gd2 map[string]any
	_ = json.Unmarshal([]byte(cliDomain), &gd1)
	_ = json.Unmarshal(getCLIHTTP(t, h, "/api/v1/domains/identity.blumer.cloud").Body.Bytes(), &gd2)
	if gd1["canonical"] != gd2["canonical"] || gd1["serviceCount"] != gd2["serviceCount"] {
		t.Fatalf("domain mismatch: %v %v", gd1, gd2)
	}

	cliService, _, err := executeCommand(t, "service", "get", "user-account", "--domain", "identity.blumer.cloud", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	var s1, s2 map[string]any
	_ = json.Unmarshal([]byte(cliService), &s1)
	_ = json.Unmarshal(getCLIHTTP(t, h, "/api/v1/domains/identity.blumer.cloud/services/user-account").Body.Bytes(), &s2)
	if s1["name"] != s2["name"] || s1["domain"] != s2["domain"] {
		t.Fatalf("service mismatch: %v %v", s1, s2)
	}

	cliGraph, _, err := executeCommand(t, "graph", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if restGraph := getCLIHTTP(t, h, "/api/v1/graph").Body.String(); cliGraph != restGraph {
		t.Fatalf("graph mismatch\n%s\n%s", cliGraph, restGraph)
	}

	cliValidate, _, err := executeCommand(t, "validate", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(cliValidate) != strings.TrimSpace(getCLIHTTP(t, h, "/api/v1/validate").Body.String()) {
		t.Fatalf("validate mismatch: %s", cliValidate)
	}
}

func getCLIHTTP(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}

func TestCLIJSONErrorsAndInvalidFormats(t *testing.T) {
	p := createParityCosmos(t)
	out, errOut, err := executeCommand(t, "domain", "get", "does-not-exist.example", "--path", p, "--format", "json")
	if err == nil || !strings.Contains(out+errOut+err.Error(), "DOMAIN_NOT_FOUND") {
		t.Fatalf("expected DOMAIN_NOT_FOUND, got out=%s errOut=%s err=%v", out, errOut, err)
	}
	out, errOut, err = executeCommand(t, "service", "get", "does-not-exist", "--domain", "identity.blumer.cloud", "--path", p, "--format", "json")
	if err == nil || !strings.Contains(out+errOut+err.Error(), "SERVICE_NOT_FOUND") {
		t.Fatalf("expected SERVICE_NOT_FOUND, got out=%s errOut=%s err=%v", out, errOut, err)
	}
	for _, args := range [][]string{{"cosmos", "info", "--path", p, "--format", "xml"}, {"domain", "list", "--path", p, "--format", "xml"}, {"validate", "--path", p, "--format", "xml"}} {
		_, _, err := executeCommand(t, args...)
		if err == nil || !strings.Contains(err.Error(), "INVALID_FORMAT") {
			t.Fatalf("expected invalid format for %v, got %v", args, err)
		}
	}
}

func TestServiceListCommand(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "collaboration.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "mailbox", "--domain", "collaboration.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "license-assignment", "--domain", "collaboration.blumer.cloud", "--path", p)
	out, _, err := executeCommand(t, "service", "list", "--domain", "collaboration.blumer.cloud", "--path", p)
	if err != nil {
		t.Fatalf("service list failed: %v", err)
	}
	if !strings.Contains(out, "mailbox") || !strings.Contains(out, "license-assignment") {
		t.Fatalf("missing services in output: %s", out)
	}
}

func TestBlueprintAndInstanceCommands(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products", "account.yaml"), `id: PB-ACC-MBX-001
type: product_blueprint
name: Benutzerkonto mit Mailbox
version: 0.1.0
status: draft
owner: Team
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-1
required_services:
  - service_ref: identity.blumer.cloud/user-account
    service_blueprint_ref: SB-1
    purpose: Erstellt und verwaltet das Benutzerkonto.
    required: true
`)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "instances", "products", "account-instance.yaml"), "id: PI-ACC-MBX-EXAMPLE-001\ntype: product_instance\nname: Beispielinstanz Benutzerkonto mit Mailbox\nblueprint_ref: PB-ACC-MBX-001\nblueprint_version: 0.1.0\ncompliance_status: compliant\nfindings: []\n")

	out, _, err := executeCommand(t, "blueprint", "list", "--path", p)
	if err != nil {
		t.Fatalf("blueprint list failed: %v", err)
	}
	if !strings.Contains(out, "PB-ACC-MBX-001") {
		t.Fatalf("missing blueprint in output: %s", out)
	}

	out, _, err = executeCommand(t, "blueprint", "show", "PB-ACC-MBX-001", "--path", p, "--format", "json")
	if err != nil {
		t.Fatalf("blueprint show json failed: %v", err)
	}
	if !strings.Contains(out, "required_services") || !strings.Contains(out, "identity.blumer.cloud/user-account") {
		t.Fatalf("missing required_services in json output: %s", out)
	}

	out, _, err = executeCommand(t, "instance", "show", "PI-ACC-MBX-EXAMPLE-001", "--path", p)
	if err != nil {
		t.Fatalf("instance show failed: %v", err)
	}
	if !strings.Contains(out, "Compliance: compliant") {
		t.Fatalf("missing compliance in output: %s", out)
	}
}

func TestBlueprintCreateCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)

	out, _, err := executeCommand(t, "blueprint", "create", "--path", p,
		"--id", "PB-CLI-001",
		"--type", "product_blueprint",
		"--name", "CLI Product",
		"--version", "0.2.0",
		"--status", "draft",
		"--owner", "CLI Team",
		"--summary", "Created via CLI",
		"--service-ref", "identity.blumer.cloud/user-account",
		"--service-blueprint-ref", "SB-ACC-001",
	)
	if err != nil {
		t.Fatalf("blueprint create failed: %v\nstdout: %s", err, out)
	}
	if !strings.Contains(out, "PB-CLI-001") {
		t.Fatalf("expected blueprint ID in output: %s", out)
	}

	out, _, err = executeCommand(t, "blueprint", "show", "PB-CLI-001", "--path", p)
	if err != nil {
		t.Fatalf("blueprint show after create failed: %v", err)
	}
	if !strings.Contains(out, "CLI Product") {
		t.Fatalf("expected blueprint name in show output: %s", out)
	}

	_, _, err = executeCommand(t, "blueprint", "create", "--path", p,
		"--id", "PB-CLI-001",
		"--type", "product_blueprint",
		"--name", "Duplicate",
	)
	if err == nil {
		t.Fatal("expected duplicate ID error")
	}
}

func mustWriteCLI(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBlueprintDeleteCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "blueprint", "create", "--path", p,
		"--id", "PB-DEL-CLI-001",
		"--type", "product_blueprint",
		"--name", "To Delete",
	)

	out, _, err := executeCommand(t, "blueprint", "delete", "PB-DEL-CLI-001", "--path", p)
	if err != nil {
		t.Fatalf("blueprint delete failed: %v", err)
	}
	if !strings.Contains(out, "PB-DEL-CLI-001") {
		t.Fatalf("expected deleted ID in output: %s", out)
	}

	_, _, err = executeCommand(t, "blueprint", "show", "PB-DEL-CLI-001", "--path", p)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestDomainDeleteCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p, "--owner", "Identity Team")
	_, _, _ = executeCommand(t, "service", "add", "user-account", "--domain", "identity.blumer.cloud", "--path", p)

	out, _, err := executeCommand(t, "domain", "delete", "identity.blumer.cloud", "--path", p)
	if err != nil {
		t.Fatalf("domain delete failed: %v", err)
	}
	if !strings.Contains(out, "identity.blumer.cloud") {
		t.Fatalf("expected deleted domain in output: %s", out)
	}

	_, _, err = executeCommand(t, "domain", "get", "identity.blumer.cloud", "--path", p)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestServiceDeleteCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "domain", "add", "identity.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "user-account", "--domain", "identity.blumer.cloud", "--path", p)
	_, _, _ = executeCommand(t, "service", "add", "privileged-account", "--domain", "identity.blumer.cloud", "--path", p)

	out, _, err := executeCommand(t, "service", "delete", "user-account", "--domain", "identity.blumer.cloud", "--path", p)
	if err != nil {
		t.Fatalf("service delete failed: %v", err)
	}
	if !strings.Contains(out, "user-account") {
		t.Fatalf("expected deleted service in output: %s", out)
	}

	_, _, err = executeCommand(t, "service", "get", "user-account", "--domain", "identity.blumer.cloud", "--path", p)
	if err == nil {
		t.Fatal("expected error after delete")
	}

	// Remaining service still exists
	out, _, err = executeCommand(t, "service", "get", "privileged-account", "--domain", "identity.blumer.cloud", "--path", p)
	if err != nil {
		t.Fatalf("remaining service should still exist: %v", err)
	}
	if !strings.Contains(out, "privileged-account") {
		t.Fatalf("expected remaining service in output: %s", out)
	}
}

func TestBlueprintPublishCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "blueprint", "create", "--path", p,
		"--id", "PB-PUB-001",
		"--type", "product_blueprint",
		"--name", "To Publish",
	)

	out, _, err := executeCommand(t, "blueprint", "show", "PB-PUB-001", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Status: draft") {
		t.Fatalf("expected draft status before publish: %s", out)
	}

	out, _, err = executeCommand(t, "blueprint", "publish", "PB-PUB-001", "--path", p)
	if err != nil {
		t.Fatalf("blueprint publish failed: %v", err)
	}
	if !strings.Contains(out, "PB-PUB-001") {
		t.Fatalf("expected blueprint ID in publish output: %s", out)
	}

	out, _, err = executeCommand(t, "blueprint", "show", "PB-PUB-001", "--path", p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Status: published") {
		t.Fatalf("expected published status after publish: %s", out)
	}
}

func TestBlueprintValidateCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products", "valid.yaml"), `id: PB-VALID-001
type: product_blueprint
name: Valid Blueprint
version: 0.1.0
status: draft
owner: Team
required_inputs:
  - person_reference
required_service_blueprints:
  - SB-1
required_services:
  - service_ref: identity.blumer.cloud/user-account
    service_blueprint_ref: SB-1
    required: true
`)

	// validate on a blueprint that has cross-ref errors (service not in cosmos): should fail
	_, _, err := executeCommand(t, "blueprint", "validate", "PB-VALID-001", "--path", p)
	if err == nil {
		t.Fatal("expected validation error for missing service ref")
	}

	// validate with --format json returns valid JSON
	out, _, _ := executeCommand(t, "blueprint", "validate", "PB-VALID-001", "--path", p, "--format", "json")
	var res map[string]any
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatalf("expected JSON output from blueprint validate: %v (%s)", err, out)
	}

	// blueprint that does not exist returns error
	_, _, err = executeCommand(t, "blueprint", "validate", "DOES-NOT-EXIST", "--path", p)
	if err == nil {
		t.Fatal("expected error for non-existent blueprint")
	}
}

func TestInstanceCreateCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products", "bp.yaml"), "id: PB-IC-001\ntype: product_blueprint\nname: BP\nversion: 0.1.0\nstatus: draft\nowner: Team\nrequired_inputs: []\nrequired_service_blueprints: []\nquality_criteria: []\nevidence_requirements: []\n")

	out, _, err := executeCommand(t, "instance", "create", "--path", p,
		"--id", "PI-IC-001",
		"--blueprint-ref", "PB-IC-001",
		"--type", "product_instance",
		"--owner", "Test Team",
	)
	if err != nil {
		t.Fatalf("instance create failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "PI-IC-001") {
		t.Fatalf("expected instance ID in output: %s", out)
	}

	out, _, err = executeCommand(t, "instance", "show", "PI-IC-001", "--path", p)
	if err != nil {
		t.Fatalf("instance show after create failed: %v", err)
	}
	if !strings.Contains(out, "PB-IC-001") {
		t.Fatalf("expected blueprint ref in show output: %s", out)
	}

	// Duplicate should fail
	_, _, err = executeCommand(t, "instance", "create", "--path", p,
		"--id", "PI-IC-001",
		"--blueprint-ref", "PB-IC-001",
	)
	if err == nil {
		t.Fatal("expected duplicate ID error")
	}

	// Missing --id should fail
	_, _, err = executeCommand(t, "instance", "create", "--path", p, "--blueprint-ref", "PB-IC-001")
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
}

func TestInstanceVerifyCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	_, _, _ = executeCommand(t, "instance", "create", "--path", p,
		"--id", "PI-VERIFY-001",
		"--blueprint-ref", "PB-ANYTHING",
		"--type", "product_instance",
	)

	out, _, err := executeCommand(t, "instance", "verify", "PI-VERIFY-001", "--path", p)
	if err != nil {
		t.Fatalf("instance verify failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "PI-VERIFY-001") {
		t.Fatalf("expected instance ID in verify output: %s", out)
	}
	if !strings.Contains(out, "compliant") {
		t.Fatalf("expected compliant status in verify output: %s", out)
	}

	// After verify, instance show should reflect evidence
	out, _, err = executeCommand(t, "instance", "show", "PI-VERIFY-001", "--path", p, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	var inst map[string]any
	if err := json.Unmarshal([]byte(out), &inst); err != nil {
		t.Fatalf("invalid JSON from instance show: %v", err)
	}
	evidence, _ := inst["evidence"].([]any)
	if len(evidence) == 0 {
		t.Fatalf("expected evidence entries after verify, got none: %s", out)
	}
}

func TestInstanceListFilterCLI(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "instances", "products", "active.yaml"), "id: PI-ACTIVE-001\ntype: product_instance\nname: Active\nblueprint_ref: PB-X\nblueprint_version: 0.1.0\nstatus: active\ncompliance_status: compliant\nfindings: []\n")
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "instances", "products", "draft.yaml"), "id: PI-DRAFT-001\ntype: product_instance\nname: Draft\nblueprint_ref: PB-X\nblueprint_version: 0.1.0\nstatus: draft\ncompliance_status: unknown\nfindings: []\n")

	out, _, err := executeCommand(t, "instance", "list", "--path", p, "--filter", "status=active")
	if err != nil {
		t.Fatalf("instance list --filter failed: %v", err)
	}
	if !strings.Contains(out, "PI-ACTIVE-001") {
		t.Fatalf("expected active instance in filtered output: %s", out)
	}
	if strings.Contains(out, "PI-DRAFT-001") {
		t.Fatalf("draft instance should be filtered out: %s", out)
	}

	// Invalid filter format should fail
	_, _, err = executeCommand(t, "instance", "list", "--path", p, "--filter", "badfilter")
	if err == nil {
		t.Fatal("expected error for invalid filter format")
	}
}

func TestValidateCmdExitCodes(t *testing.T) {
	p := t.TempDir()
	_, _, _ = executeCommand(t, "cosmos", "init", p)

	// Empty cosmos with no findings should succeed (exit 0)
	_, _, err := executeCommand(t, "validate", "--path", p)
	if err != nil {
		t.Fatalf("validate on empty cosmos should succeed, got: %v", err)
	}

	// Blueprint with only warnings should also succeed (exit 0)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "blueprints", "services", "warn.yaml"), `id: SB-WARN-001
type: service_blueprint
name: Warn Blueprint
version: 0.1.0
status: draft
owner: Team
capabilities:
  - email
`)
	_, _, err = executeCommand(t, "validate", "--path", p)
	if err != nil {
		t.Fatalf("validate with warnings only should succeed, got: %v", err)
	}

	// Blueprint with error-severity findings should fail (exit 1)
	mustWriteCLI(t, filepath.Join(storage.CatalogDir(p), "blueprints", "products", "bad.yaml"), `id: PB-BAD-001
type: product_blueprint
name: Bad Blueprint
version: 0.1.0
status: draft
owner: Team
`)
	_, _, err = executeCommand(t, "validate", "--path", p)
	if err == nil {
		t.Fatal("expected validate to fail on error-severity findings")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "  ", "hello", "world"); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	if got := firstNonEmpty("first"); got != "first" {
		t.Fatalf("expected first, got %q", got)
	}
}

func TestLoadCosmosTree(t *testing.T) {
	p := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Dir(storage.CosmosFile(p)), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-test\nname: Test Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "example.com", "services", "api"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "example.com", "domain.yaml"), []byte("name: example.com\n"), 0o644))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "example.com", "services", "api", "service.yaml"), []byte("name: api\n"), 0o644))

	tree, err := loadCosmosTree(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tree.Cosmos.Name != "Test Cosmos" {
		t.Fatalf("expected Test Cosmos, got %q", tree.Cosmos.Name)
	}
	if len(tree.Domains) != 1 || tree.Domains[0].Name != "example.com" {
		t.Fatalf("expected one domain example.com, got %+v", tree.Domains)
	}
	if len(tree.Domains[0].Services) != 1 || tree.Domains[0].Services[0].Name != "api" {
		t.Fatalf("expected one service api, got %+v", tree.Domains[0].Services)
	}
}

func TestScanDomains_MissingRoot(t *testing.T) {
	p := t.TempDir()
	domains, err := scanDomains(p)
	if err != nil {
		t.Fatalf("unexpected error for missing domains dir: %v", err)
	}
	if len(domains) != 0 {
		t.Fatalf("expected empty domains, got %+v", domains)
	}
}

func TestScanServices_MissingDir(t *testing.T) {
	p := t.TempDir()
	services, err := scanServices(p)
	if err != nil {
		t.Fatalf("unexpected error for missing services dir: %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected empty services, got %+v", services)
	}
}

func TestPrintNamespaceNode(t *testing.T) {
	p := t.TempDir()
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	must(os.MkdirAll(filepath.Dir(storage.CosmosFile(p)), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-test\nname: Test\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644))
	must(os.MkdirAll(filepath.Join(storage.DomainsDir(p), "example.com", "services"), 0o755))
	must(os.WriteFile(filepath.Join(storage.DomainsDir(p), "example.com", "domain.yaml"), []byte("name: example.com\n"), 0o644))

	out, _, err := executeCommand(t, "namespace", "tree", "--path", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "com") {
		t.Fatalf("expected namespace output to contain 'com', got: %s", out)
	}
}
