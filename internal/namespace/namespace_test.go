package namespace

import (
	"reflect"
	"testing"
)

func TestNamespaceTransforms(t *testing.T) {
	for _, canonical := range []string{"identity.blumer.cloud", "assurance.blumer.cloud", "platform.blumer.cloud"} {
		parts := []string{canonical[:len(canonical)-len(".blumer.cloud")], "blumer", "cloud"}
		if got := Parts(canonical); !reflect.DeepEqual(got, parts) {
			t.Fatalf("Parts(%q)=%v", canonical, got)
		}
		wantTree := []string{"cloud", "blumer", parts[0]}
		if got := TreeParts(canonical); !reflect.DeepEqual(got, wantTree) {
			t.Fatalf("TreeParts(%q)=%v", canonical, got)
		}
		if got, want := TreePath(canonical), "cloud/blumer/"+parts[0]; got != want {
			t.Fatalf("TreePath(%q)=%q want %q", canonical, got, want)
		}
		if got, want := DisplayPath(canonical), "cloud / blumer / "+parts[0]; got != want {
			t.Fatalf("DisplayPath(%q)=%q want %q", canonical, got, want)
		}
		if got := Leaf(canonical); got != parts[0] {
			t.Fatalf("Leaf(%q)=%q", canonical, got)
		}
	}
}

func TestView(t *testing.T) {
	got := View(" identity.blumer.cloud ")
	if got.Canonical != "identity.blumer.cloud" || got.TreePath != "cloud/blumer/identity" || got.DisplayPath != "cloud / blumer / identity" || got.Leaf != "identity" {
		t.Fatalf("unexpected view: %+v", got)
	}
}
