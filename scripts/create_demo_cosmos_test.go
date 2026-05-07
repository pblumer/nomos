package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		"domains/blumer.com/domain.yaml",
		"domains/identity.blumer.com/services/user-account/service.yaml",
		"domains/governance.blumer.com/services/provisioning-rules/service.yaml",
		"domains/home.blumer.cloud/services/home-dashboard/service.yaml",
		"domains/zytlog.blumer.cloud/services/zytlog-api/service.yaml",
		"domains/beispiel.ch/domain.yaml",
		"catalog/blueprints/products/benutzerkonto-mit-mailbox.yaml",
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing %s in explicit target: %v", rel, err)
		}
	}
}
