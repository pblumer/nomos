package namespace

import (
	"reflect"
	"testing"
)

func TestNamespaceTransforms(t *testing.T) {
	cases := []struct {
		canonical       string
		namespace       string
		labels          []string
		label           string
		parentCanonical string
		treePath        string
		displayPath     string
	}{
		{"blumer.com", "com", []string{"blumer"}, "blumer", "", "/com/blumer", "com / blumer"},
		{"identity.blumer.com", "com", []string{"blumer", "identity"}, "identity", "blumer.com", "/com/blumer/identity", "com / blumer / identity"},
		{"blumer.cloud", "cloud", []string{"blumer"}, "blumer", "", "/cloud/blumer", "cloud / blumer"},
		{"home.blumer.cloud", "cloud", []string{"blumer", "home"}, "home", "blumer.cloud", "/cloud/blumer/home", "cloud / blumer / home"},
		{"beispiel.ch", "ch", []string{"beispiel"}, "beispiel", "", "/ch/beispiel", "ch / beispiel"},
	}
	for _, tc := range cases {
		if got := Namespace(tc.canonical); got != tc.namespace {
			t.Fatalf("Namespace(%q)=%q want %q", tc.canonical, got, tc.namespace)
		}
		if got := Labels(tc.canonical); !reflect.DeepEqual(got, tc.labels) {
			t.Fatalf("Labels(%q)=%v want %v", tc.canonical, got, tc.labels)
		}
		if got := Label(tc.canonical); got != tc.label {
			t.Fatalf("Label(%q)=%q want %q", tc.canonical, got, tc.label)
		}
		if got := Leaf(tc.canonical); got != tc.label {
			t.Fatalf("Leaf(%q)=%q want %q", tc.canonical, got, tc.label)
		}
		if got := ParentCanonical(tc.canonical); got != tc.parentCanonical {
			t.Fatalf("ParentCanonical(%q)=%q want %q", tc.canonical, got, tc.parentCanonical)
		}
		if got := TreePath(tc.canonical); got != tc.treePath {
			t.Fatalf("TreePath(%q)=%q want %q", tc.canonical, got, tc.treePath)
		}
		if got := DisplayPath(tc.canonical); got != tc.displayPath {
			t.Fatalf("DisplayPath(%q)=%q want %q", tc.canonical, got, tc.displayPath)
		}
	}
}

func TestView(t *testing.T) {
	got := View(" identity.blumer.com ")
	if got.Canonical != "identity.blumer.com" || got.Namespace != "com" || got.Label != "identity" || got.Leaf != "identity" || got.ParentCanonical != "blumer.com" || got.TreePath != "/com/blumer/identity" || got.DisplayPath != "com / blumer / identity" {
		t.Fatalf("unexpected view: %+v", got)
	}
}

func TestComposeCanonical(t *testing.T) {
	got, err := ComposeCanonical("com", "blumer", "identity")
	if err != nil {
		t.Fatal(err)
	}
	if got != "identity.blumer.com" {
		t.Fatalf("got %q", got)
	}
	got, err = ComposeChildCanonical("cloud", "blumer")
	if err != nil {
		t.Fatal(err)
	}
	if got != "blumer.cloud" {
		t.Fatalf("got %q", got)
	}
}

func TestTreePathCanonicalAndGitPathMappings(t *testing.T) {
	cases := []struct {
		treePath  string
		canonical string
	}{
		{"/cloud/blumer", "blumer.cloud"},
		{"/cloud/blumer/example1", "example1.blumer.cloud"},
		{"/cloud/blumer/example1/products", "products.example1.blumer.cloud"},
	}
	for _, tc := range cases {
		gotCanonical, err := TreePathToCanonical(tc.treePath)
		if err != nil {
			t.Fatalf("TreePathToCanonical(%q): %v", tc.treePath, err)
		}
		if gotCanonical != tc.canonical {
			t.Fatalf("TreePathToCanonical(%q)=%q want %q", tc.treePath, gotCanonical, tc.canonical)
		}
		gotTreePath, err := CanonicalToTreePath(tc.canonical)
		if err != nil {
			t.Fatalf("CanonicalToTreePath(%q): %v", tc.canonical, err)
		}
		if gotTreePath != tc.treePath {
			t.Fatalf("CanonicalToTreePath(%q)=%q want %q", tc.canonical, gotTreePath, tc.treePath)
		}
	}

	gotGitPath, err := TreePathToGitPath("/cloud/blumer/example1/products")
	if err != nil {
		t.Fatalf("TreePathToGitPath: %v", err)
	}
	if want := ".nomos/domains/cloud/blumer/example1/products/domain.yaml"; gotGitPath != want {
		t.Fatalf("TreePathToGitPath=%q want %q", gotGitPath, want)
	}
}
