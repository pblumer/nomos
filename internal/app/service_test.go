package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createAppTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Join(p, "domains/identity.blumer.cloud/services/user-account"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "domains/identity.blumer.cloud/services/privileged-account"), 0o755))
	must(os.MkdirAll(filepath.Join(p, "domains/platform.blumer.cloud/services/rule-validation-api"), 0o755))
	must(os.WriteFile(filepath.Join(p, "cosmos.yaml"), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Platform Team\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/domain.yaml"), []byte("name: identity.blumer.cloud\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/platform.blumer.cloud/domain.yaml"), []byte("name: platform.blumer.cloud\nowner: Platform Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/services/user-account/service.yaml"), []byte("name: user-account\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/identity.blumer.cloud/services/privileged-account/service.yaml"), []byte("name: privileged-account\nowner: Identity Team\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(p, "domains/platform.blumer.cloud/services/rule-validation-api/service.yaml"), []byte("name: rule-validation-api\nowner: Platform Team\nstatus: draft\n"), 0o644))
	return p
}

func TestAppUseCasesExposeNamespaceMetadata(t *testing.T) {
	p := createAppTestCosmos(t)
	cosmos, err := GetCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	if cosmos.DomainCount != 2 || cosmos.ServiceCount != 3 {
		t.Fatalf("unexpected counts: %+v", cosmos)
	}
	domains, err := ListDomains(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := domains.Domains[0]; got.Canonical != "identity.blumer.cloud" || got.Namespace.DisplayPath != "cloud / blumer / identity" || got.DisplayName != "identity" || got.ServiceCount != 2 {
		t.Fatalf("unexpected domain: %+v", got)
	}
	domain, err := GetDomain(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatal(err)
	}
	if len(domain.Services) != 2 || domain.Services[0].Name != "privileged-account" || domain.Services[1].Name != "user-account" {
		t.Fatalf("services not sorted: %+v", domain.Services)
	}
	service, err := GetService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatal(err)
	}
	if service.Domain != "identity.blumer.cloud" || service.Name != "user-account" {
		t.Fatalf("unexpected service: %+v", service)
	}
	graph, err := BuildGraph(p)
	if err != nil {
		t.Fatal(err)
	}
	if graph.Format != "mermaid" || !strings.Contains(graph.Content, "cloud / blumer / identity") {
		t.Fatalf("unexpected graph: %+v", graph)
	}
	validation, err := ValidateCosmos(p)
	if err != nil {
		t.Fatal(err)
	}
	if validation.Status != "ok" || len(validation.Findings) != 0 {
		t.Fatalf("unexpected validation: %+v", validation)
	}
}

func TestBuildNamespaceTree(t *testing.T) {
	tree, err := BuildNamespaceTree(createAppTestCosmos(t))
	if err != nil {
		t.Fatal(err)
	}
	body := strings.Join(flattenLabels(tree.Root), "|")
	for _, want := range []string{"Local Cosmos", "cloud", "blumer", "identity", "platform", "user-account", "rule-validation-api"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s in %s", want, body)
		}
	}
}

func flattenLabels(n NamespaceTreeNodeDTO) []string {
	out := []string{n.Label}
	for _, c := range n.Children {
		out = append(out, flattenLabels(c)...)
	}
	return out
}

func TestAppNotFoundErrors(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := GetDomain(p, "does-not-exist.example"); err == nil || !strings.Contains(err.Error(), CodeDomainNotFound) {
		t.Fatalf("expected domain error, got %v", err)
	}
	if _, err := GetService(p, "identity.blumer.cloud", "does-not-exist"); err == nil || !strings.Contains(err.Error(), CodeServiceNotFound) {
		t.Fatalf("expected service error, got %v", err)
	}
}

func TestDeleteDomain(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteDomain(p, "identity.blumer.cloud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, "domains", "identity.blumer.cloud")); !os.IsNotExist(err) {
		t.Fatal("expected domain directory to be removed")
	}
	if _, err := GetDomain(p, "identity.blumer.cloud"); err == nil {
		t.Fatal("expected domain not found after delete")
	}

	// Remaining domain should still exist
	if _, err := GetDomain(p, "platform.blumer.cloud"); err != nil {
		t.Fatalf("unexpected error for remaining domain: %v", err)
	}
}

func TestDeleteDomain_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteDomain(p, "does-not-exist.example")
	if err == nil {
		t.Fatal("expected error")
	}
	if ae, ok := AsAppError(err); !ok || ae.Code != CodeDomainNotFound {
		t.Fatalf("expected DOMAIN_NOT_FOUND, got %v", err)
	}
}

func TestDeleteService(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteService(p, "identity.blumer.cloud", "user-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p, "domains", "identity.blumer.cloud", "services", "user-account")); !os.IsNotExist(err) {
		t.Fatal("expected service directory to be removed")
	}
	if _, err := GetService(p, "identity.blumer.cloud", "user-account"); err == nil {
		t.Fatal("expected service not found after delete")
	}

	// Remaining services should still exist
	if _, err := GetService(p, "identity.blumer.cloud", "privileged-account"); err != nil {
		t.Fatalf("unexpected error for remaining service: %v", err)
	}
}

func TestDeleteService_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := DeleteService(p, "identity.blumer.cloud", "does-not-exist")
	if err == nil {
		t.Fatal("expected error")
	}
	if ae, ok := AsAppError(err); !ok || ae.Code != CodeServiceNotFound {
		t.Fatalf("expected SERVICE_NOT_FOUND, got %v", err)
	}
}

func TestRenameDomain(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameDomain(p, "identity.blumer.cloud", "user.blumer.cloud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, e := os.Stat(filepath.Join(p, "domains", "identity.blumer.cloud")); !os.IsNotExist(e) {
		t.Fatal("expected old domain dir to be gone")
	}
	if _, e := os.Stat(filepath.Join(p, "domains", "user.blumer.cloud")); os.IsNotExist(e) {
		t.Fatal("expected new domain dir to exist")
	}
	if _, e := os.Stat(filepath.Join(p, "domains", "user.blumer.cloud", "services", "user-account", "service.yaml")); os.IsNotExist(e) {
		t.Fatal("expected service to survive domain rename")
	}
	d, err := GetDomain(p, "user.blumer.cloud")
	if err != nil {
		t.Fatalf("could not get renamed domain: %v", err)
	}
	if d.Name != "user.blumer.cloud" {
		t.Fatalf("expected name user.blumer.cloud, got %s", d.Name)
	}
	// Old name gone
	if _, err := GetDomain(p, "identity.blumer.cloud"); err == nil {
		t.Fatal("expected old name to be gone")
	}
}

func TestRenameDomain_NameConflict(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameDomain(p, "identity.blumer.cloud", "platform.blumer.cloud")
	if err == nil {
		t.Fatal("expected error for name conflict")
	}
}

func TestRenameService(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "identity.blumer.cloud", "user-account", "identity-account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, e := os.Stat(filepath.Join(p, "domains", "identity.blumer.cloud", "services", "user-account")); !os.IsNotExist(e) {
		t.Fatal("expected old service dir to be gone")
	}
	if _, e := os.Stat(filepath.Join(p, "domains", "identity.blumer.cloud", "services", "identity-account")); os.IsNotExist(e) {
		t.Fatal("expected new service dir to exist")
	}
	s, err := GetService(p, "identity.blumer.cloud", "identity-account")
	if err != nil {
		t.Fatalf("could not get renamed service: %v", err)
	}
	if s.Name != "identity-account" || s.Domain != "identity.blumer.cloud" {
		t.Fatalf("unexpected service: %+v", s)
	}
	if _, err := GetService(p, "identity.blumer.cloud", "user-account"); err == nil {
		t.Fatal("expected old name to be gone")
	}
}

func TestRenameService_NotFound(t *testing.T) {
	p := createAppTestCosmos(t)
	err := RenameService(p, "identity.blumer.cloud", "does-not-exist", "new-name")
	if err == nil {
		t.Fatal("expected error")
	}
}
