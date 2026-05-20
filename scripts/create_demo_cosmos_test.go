package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreateDemoCosmosScriptRespectsExplicitTargetAndCreatesFlatDemo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test path")
	}
	repoRoot := filepath.Dir(filepath.Dir(file))
	bin := filepath.Join(t.TempDir(), "nomos")
	build := exec.Command("go", "build", "-o", bin, "./cmd/nomos")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build nomos: %v\n%s", err, out)
	}
	target := filepath.Join(t.TempDir(), "nomos-demo")
	cmd := exec.Command("sh", "scripts/create-demo-cosmos.sh", target)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "NOMOS_BIN="+bin, "COSMOS_PATH="+filepath.Join(t.TempDir(), "ignored-cosmos-path"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create demo cosmos: %v\n%s", err, out)
	}
	for _, rel := range []string{
		filepath.Join(".nomos", "services", "user-account", "service.yaml"),
		filepath.Join(".nomos", "services", "mailbox", "service.yaml"),
		filepath.Join(".nomos", "services", "exchange", "service.yaml"),
		filepath.Join(".nomos", "services", "license-assignment", "service.yaml"),
		filepath.Join(".nomos", "services", "license", "service.yaml"),
		filepath.Join(".nomos", "services", "provisioning-rules", "service.yaml"),
		filepath.Join(".nomos", "services", "home-dashboard", "service.yaml"),
		filepath.Join(".nomos", "services", "zytlog-api", "service.yaml"),
		filepath.Join(".nomos", "services", "checkout", "service.yaml"),
		filepath.Join(".nomos", "services", "inventory", "service.yaml"),
		filepath.Join(".nomos", "services", "payment", "service.yaml"),
		filepath.Join(".nomos", "catalog", "blueprints", "products", "benutzerkonto-mit-mailbox.yaml"),
		filepath.Join(".nomos", "catalog", "blueprints", "products", "cloud-mailbox.yaml"),
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing %s in explicit target: %v", rel, err)
		}
	}
	product, err := os.ReadFile(filepath.Join(target, ".nomos", "catalog", "blueprints", "products", "benutzerkonto-mit-mailbox.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"id: PROD-ACC-MBX-001", "fulfillment:", "service_ref: user-account", "service_ref: exchange", "sla_ref: SLA-IDENTITY-ACCOUNT-STANDARD", "ola_ref: OLA-IDENTITY-OPS-STANDARD", "sla_ref: SLA-MAILBOX-STANDARD", "ola_ref: OLA-MAILING-OPS-STANDARD"} {
		if !strings.Contains(string(product), want) {
			t.Fatalf("demo product blueprint missing %q", want)
		}
	}
	cloudProduct, err := os.ReadFile(filepath.Join(target, ".nomos", "catalog", "blueprints", "products", "cloud-mailbox.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"id: PROD-CLOUD-MAILBOX-001", "service_ref: user-account", "service_ref: exchange", "ola:", "sla:"} {
		if !strings.Contains(string(cloudProduct), want) {
			t.Fatalf("demo cloud product missing %q", want)
		}
	}

	mailboxService, err := os.ReadFile(filepath.Join(target, ".nomos", "services", "mailbox", "service.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mailboxService), "sla:") || !strings.Contains(string(mailboxService), "target: 8h") {
		t.Fatalf("demo mailbox service missing SLA fallback metadata")
	}
	validate := exec.Command(bin, "validate", "--path", target)
	validate.Dir = repoRoot
	if out, err := validate.CombinedOutput(); err != nil {
		t.Fatalf("demo validation should have no errors: %v\n%s", err, out)
	}
}
