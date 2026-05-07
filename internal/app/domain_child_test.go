package app

import "testing"

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
