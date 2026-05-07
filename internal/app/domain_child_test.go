package app

import (
	"strings"
	"testing"
)

func TestAddDomainInNamespaceComposesCanonicalName(t *testing.T) {
	p := createAppTestCosmos(t)
	d, err := AddDomainInNamespace(p, "com", "blumer", "UX", false)
	if err != nil {
		t.Fatalf("AddDomainInNamespace: %v", err)
	}
	if d.Canonical != "blumer.com" || d.Label != "blumer" || d.Namespace.Namespace != "com" {
		t.Fatalf("unexpected domain: %+v", d)
	}
}

func TestAddChildDomainComposesCanonicalName(t *testing.T) {
	cases := []struct{ parent, segment, want string }{
		{"blumer.com", "identity", "identity.blumer.com"},
		{"cloud", "blumer", "blumer.cloud"},
		{"blumer.cloud", "home", "home.blumer.cloud"},
	}
	for _, tc := range cases {
		p := createAppTestCosmos(t)
		if strings.Contains(tc.parent, ".") {
			if _, err := AddDomain(p, tc.parent, "UX", false); err != nil {
				t.Fatalf("AddDomain parent %q: %v", tc.parent, err)
			}
		}
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
	invalid := []string{"test2.blumer.cloud", "test 2", "/test2", ".test2", "test2.", "-test2", "test2-"}
	for _, segment := range invalid {
		if _, err := AddChildDomain(createAppTestCosmos(t), "blumer.cloud", segment, "UX", false); err == nil {
			t.Fatalf("expected invalid segment %q to fail", segment)
		}
	}
}

func TestAddDomainNormalizesCanonicalInputBehavior(t *testing.T) {
	p := createAppTestCosmos(t)
	d, err := AddDomain(p, "Team.Example", "Legacy", false)
	if err != nil {
		t.Fatalf("AddDomain should preserve full canonical input behavior: %v", err)
	}
	if d.Canonical != "team.example" {
		t.Fatalf("got %q", d.Canonical)
	}
}

func TestAddChildDomainCreatesTreeOrderedGitPathAndLoadsByCanonical(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := AddDomain(p, "blumer.cloud", "Patrick Blumer", false); err != nil {
		t.Fatalf("AddDomain parent: %v", err)
	}
	child, err := AddChildDomain(p, "blumer.cloud", "example1", "Patrick Blumer", false)
	if err != nil {
		t.Fatalf("AddChildDomain example1: %v", err)
	}
	leaf, err := AddChildDomain(p, child.Canonical, "products", "Patrick Blumer", false)
	if err != nil {
		t.Fatalf("AddChildDomain products: %v", err)
	}
	if leaf.Canonical != "products.example1.blumer.cloud" {
		t.Fatalf("canonical=%q", leaf.Canonical)
	}
	if leaf.TreePath != "/cloud/blumer/example1/products" {
		t.Fatalf("treePath=%q", leaf.TreePath)
	}
	if leaf.GitPath != ".nomos/domains/cloud/blumer/example1/products/domain.yaml" {
		t.Fatalf("gitPath=%q", leaf.GitPath)
	}
	if leaf.ParentCanonical != "example1.blumer.cloud" || leaf.ParentTreePath != "/cloud/blumer/example1" {
		t.Fatalf("unexpected parent identity: %+v", leaf)
	}
	loaded, err := GetDomain(p, "products.example1.blumer.cloud")
	if err != nil {
		t.Fatalf("GetDomain by canonical: %v", err)
	}
	if loaded.Canonical != leaf.Canonical || loaded.Path == "" {
		t.Fatalf("loaded=%+v leaf=%+v", loaded, leaf)
	}
}
