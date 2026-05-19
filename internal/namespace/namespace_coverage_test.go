package namespace_test

import (
	"testing"

	"github.com/nomos/nomos/internal/namespace"
)

func TestIdentity(t *testing.T) {
	id, err := namespace.Identity("identity.blumer.cloud")
	if err != nil {
		t.Fatalf("Identity: %v", err)
	}
	if id.CanonicalName != "identity.blumer.cloud" {
		t.Errorf("unexpected CanonicalName: %q", id.CanonicalName)
	}
	if id.Label != "identity" {
		t.Errorf("unexpected Label: %q", id.Label)
	}
	if id.Namespace != "cloud" {
		t.Errorf("unexpected Namespace: %q", id.Namespace)
	}
	if id.GitPath == "" {
		t.Error("expected non-empty GitPath")
	}
}

func TestIdentity_Invalid(t *testing.T) {
	_, err := namespace.Identity("notsufficient")
	if err == nil {
		t.Error("expected error for single-label canonical")
	}
}

func TestCanonicalToTreePath(t *testing.T) {
	cases := []struct {
		canonical, wantTree string
	}{
		{"identity.blumer.cloud", "/cloud/blumer/identity"},
		{"core.nomos", "/nomos/core"},
	}
	for _, c := range cases {
		got, err := namespace.CanonicalToTreePath(c.canonical)
		if err != nil {
			t.Errorf("CanonicalToTreePath(%q) error: %v", c.canonical, err)
			continue
		}
		if string(got) != c.wantTree {
			t.Errorf("CanonicalToTreePath(%q) = %q, want %q", c.canonical, got, c.wantTree)
		}
	}
}

func TestTreePathToGitPath(t *testing.T) {
	gitPath, err := namespace.TreePathToGitPath("/cloud/blumer/identity")
	if err != nil {
		t.Fatalf("TreePathToGitPath: %v", err)
	}
	if gitPath == "" {
		t.Error("expected non-empty git path")
	}
}

func TestTreePathToGitPath_InvalidLabel(t *testing.T) {
	_, err := namespace.TreePathToGitPath("/cloud/UPPERCASE")
	if err == nil {
		t.Error("expected error for uppercase label")
	}
}

func TestValidLabelEdgeCases(t *testing.T) {
	// too-long label (> 63 chars) — validLabel rejects it
	longLabel := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 64 chars
	_, err := namespace.ComposeCanonical("cloud", longLabel)
	if err == nil {
		t.Error("expected error for label > 63 chars")
	}
	// label with a dot should fail
	_, err = namespace.ComposeCanonical("cloud", "has.dot")
	if err == nil {
		t.Error("expected error for label with dot")
	}
	// label with leading dash should fail
	_, err = namespace.ComposeCanonical("cloud", "-bad")
	if err == nil {
		t.Error("expected error for leading-dash label")
	}
}

func TestParts_MultiLevel(t *testing.T) {
	p := namespace.Parts("sub.identity.blumer.cloud")
	if len(p) != 4 {
		t.Errorf("expected 4 parts, got %d: %v", len(p), p)
	}
}

func TestNamespace_Extraction(t *testing.T) {
	if ns := namespace.Namespace("identity.blumer.cloud"); ns != "cloud" {
		t.Errorf("Namespace = %q, want cloud", ns)
	}
}

func TestLabel_Extraction(t *testing.T) {
	if l := namespace.Label("identity.blumer.cloud"); l != "identity" {
		t.Errorf("Label = %q, want identity", l)
	}
}

func TestLabels_Extraction(t *testing.T) {
	// Labels returns all tree parts except the namespace (TLD).
	// "sub.identity.blumer.cloud" → tree parts [cloud, blumer, identity, sub]
	// → labels (tree[1:]) = [blumer, identity, sub] = 3 labels
	ls := namespace.Labels("sub.identity.blumer.cloud")
	if len(ls) != 3 {
		t.Errorf("expected 3 labels, got %v", ls)
	}
	// Two-level: "identity.cloud" → tree [cloud, identity] → labels [identity] = 1
	ls2 := namespace.Labels("identity.cloud")
	if len(ls2) != 1 {
		t.Errorf("expected 1 label, got %v", ls2)
	}
}

func TestParentTreePath(t *testing.T) {
	got := namespace.ParentTreePath("identity.blumer.cloud")
	want := namespace.TreePath("blumer.cloud")
	if got != want {
		t.Errorf("ParentTreePath = %q, want %q", got, want)
	}
}

func TestComposeChildCanonical_Valid(t *testing.T) {
	got, err := namespace.ComposeChildCanonical("blumer.cloud", "identity")
	if err != nil {
		t.Fatal(err)
	}
	if got != "identity.blumer.cloud" {
		t.Errorf("got %q, want identity.blumer.cloud", got)
	}
}

func TestComposeChildCanonical_InvalidParent(t *testing.T) {
	_, err := namespace.ComposeChildCanonical("", "identity")
	if err == nil {
		t.Error("expected error for empty parent")
	}
}

func TestComposeChildCanonical_InvalidChild(t *testing.T) {
	_, err := namespace.ComposeChildCanonical("blumer.cloud", "has space")
	if err == nil {
		t.Error("expected error for invalid child label")
	}
}

func TestValidateCanonical_TooFewParts(t *testing.T) {
	_, err := namespace.Identity("onelabel")
	if err == nil {
		t.Error("expected error for single-part canonical")
	}
}
