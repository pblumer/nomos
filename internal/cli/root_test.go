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
	mustWriteCLI(t, filepath.Join(p, "catalog", "blueprints", "products", "account.yaml"), `id: PB-ACC-MBX-001
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
	_, _, _ = executeCommand(t, "cosmos", "init", p)
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
	_, _, _ = executeCommand(t, "cosmos", "init", p)
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
	_, _, _ = executeCommand(t, "cosmos", "init", p)
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
	_, _, _ = executeCommand(t, "cosmos", "init", p)
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
	_, _, _ = executeCommand(t, "cosmos", "init", p)
	if err := os.MkdirAll(filepath.Join(p, "domains", "incomplete.example.com"), 0o755); err != nil {
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
	_, _, _ = executeCommand(t, "cosmos", "init", p)
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
	mustWriteCLI(t, filepath.Join(p, "catalog", "blueprints", "products", "account.yaml"), `id: PB-ACC-MBX-001
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
	mustWriteCLI(t, filepath.Join(p, "catalog", "instances", "products", "account-instance.yaml"), "id: PI-ACC-MBX-EXAMPLE-001\ntype: product_instance\nname: Beispielinstanz Benutzerkonto mit Mailbox\nblueprint_ref: PB-ACC-MBX-001\nblueprint_version: 0.1.0\ncompliance_status: compliant\nfindings: []\n")

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

func mustWriteCLI(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
