package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreateDemoCosmosScriptRespectsExplicitTargetAndCreatesDNSLikeDemo(t *testing.T) {
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
		filepath.Join(".nomos", "domains", "com", "blumer", "domain.yaml"),
		filepath.Join(".nomos", "domains", "net", "blumer", "domain.yaml"),
		filepath.Join(".nomos", "domains", "com", "blumer", "identity", "services", "user-account", "service.yaml"),
		filepath.Join(".nomos", "domains", "cloud", "blumer", "identity", "domain.yaml"),
		filepath.Join(".nomos", "domains", "cloud", "blumer", "identity", "services", "user-account", "service.yaml"),
		filepath.Join(".nomos", "domains", "cloud", "blumer", "collaboration", "services", "mailbox", "service.yaml"),
		filepath.Join(".nomos", "domains", "cloud", "blumer", "collaboration", "services", "license-assignment", "service.yaml"),
		filepath.Join(".nomos", "domains", "com", "blumer", "governance", "services", "provisioning-rules", "service.yaml"),
		filepath.Join(".nomos", "domains", "cloud", "blumer", "home", "services", "home-dashboard", "service.yaml"),
		filepath.Join(".nomos", "domains", "cloud", "blumer", "zytlog", "services", "zytlog-api", "service.yaml"),
		filepath.Join(".nomos", "domains", "ch", "beispiel", "domain.yaml"),
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
	for _, want := range []string{"id: PROD-ACC-MBX-001", "offered_by: identity.blumer.cloud", "fulfillment:", "identity.blumer.cloud/user-account", "collaboration.blumer.cloud/mailbox", "collaboration.blumer.cloud/license-assignment", "ola:", "target: 4h", "sla:", "target: 8h"} {
		if !strings.Contains(string(product), want) {
			t.Fatalf("demo product blueprint missing %q", want)
		}
	}
	crossDomainProduct, err := os.ReadFile(filepath.Join(target, ".nomos", "catalog", "blueprints", "products", "cloud-mailbox.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"id: PROD-CLOUD-MAILBOX-001", "offered_by: blumer.net", "identity.blumer.cloud/user-account", "collaboration.blumer.cloud/mailbox", "ola:", "sla:"} {
		if !strings.Contains(string(crossDomainProduct), want) {
			t.Fatalf("demo cross-domain product missing %q", want)
		}
	}

	mailboxService, err := os.ReadFile(filepath.Join(target, ".nomos", "domains", "cloud", "blumer", "collaboration", "services", "mailbox", "service.yaml"))
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
