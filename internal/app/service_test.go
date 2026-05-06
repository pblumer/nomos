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

func TestAddChildDomainComposesCanonicalName(t *testing.T) {
	cases := []struct{ parent, segment, want string }{
		{"blumer.cloud", "test2", "test2.blumer.cloud"},
		{"cloud", "blumer", "blumer.cloud"},
		{"identity.blumer.cloud", "iam", "iam.identity.blumer.cloud"},
	}
	for _, tc := range cases {
		p := createAppTestCosmos(t)
		d, err := AddChildDomain(p, tc.parent, tc.segment, "UX", false)
		if err != nil {
			t.Fatalf("AddChildDomain(%q,%q): %v", tc.parent, tc.segment, err)
		}
		if d.Canonical != tc.want {
			t.Fatalf("got %q want %q", d.Canonical, tc.want)
		}
	}
}

func TestAddChildDomainRejectsFullOrInvalidSegment(t *testing.T) {
	invalid := []string{"test2.blumer.cloud", "test 2", "/test2", ".test2", "test2."}
	for _, segment := range invalid {
		if _, err := AddChildDomain(createAppTestCosmos(t), "blumer.cloud", segment, "UX", false); err == nil {
			t.Fatalf("expected invalid segment %q to fail", segment)
		}
	}
}

func TestAddDomainPreservesFullCanonicalInputBehavior(t *testing.T) {
	p := createAppTestCosmos(t)
	d, err := AddDomain(p, "Team.Example", "Legacy", false)
	if err != nil {
		t.Fatalf("AddDomain should preserve full canonical input behavior: %v", err)
	}
	if d.Canonical != "Team.Example" {
		t.Fatalf("got %q", d.Canonical)
	}
}
