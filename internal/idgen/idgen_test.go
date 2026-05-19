package idgen

import (
	"regexp"
	"strings"
	"testing"
)

func TestNewFormat(t *testing.T) {
	for i := 0; i < 200; i++ {
		id, err := New("PRD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !idRegex.MatchString(id) {
			t.Errorf("ID %q does not match format", id)
		}
		if !strings.HasPrefix(id, "PRD_") {
			t.Errorf("ID %q missing PRD_ prefix", id)
		}
		if len(id) != 10 {
			t.Errorf("ID %q wrong length: %d", id, len(id))
		}
	}
}

func TestNewRejectsBadPrefix(t *testing.T) {
	bad := []string{"", "pr", "PROD", "P1D", "AB-"}
	for _, p := range bad {
		if _, err := New(p); err == nil {
			t.Errorf("expected error for prefix %q", p)
		}
	}
}

func TestNewLowercasesIsAccepted(t *testing.T) {
	id, err := New("prd")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id, "PRD_") {
		t.Errorf("expected uppercase prefix, got %s", id)
	}
}

func TestNewForType(t *testing.T) {
	cases := map[string]string{
		"product_blueprint":   "PRD_",
		"service_blueprint":   "SVC_",
		"product_instance":    "PRI_",
		"service_instance":    "SVI_",
		"process":             "PRC_",
		"process_step":        "STP_",
		"decision":            "DEC_",
		"evidence":            "EVD_",
		"finding":             "FND_",
		"servicegraph":        "SGR_",
		"requirement":         "REQ_",
		"rule":                "RUL_",
		"validation_scenario": "VSC_",
		"variant":             "PRV_",
		"attribute":           "ATR_",
		"blueprint":           "BLP_",
		"domain":              "DOM_",
	}
	for typ, want := range cases {
		id, err := NewForType(typ)
		if err != nil {
			t.Fatalf("NewForType(%q): %v", typ, err)
		}
		if !strings.HasPrefix(id, want) {
			t.Errorf("NewForType(%q) = %q; want prefix %q", typ, id, want)
		}
	}
}

func TestNewForTypeCosmosIsFixed(t *testing.T) {
	id, err := NewForType("cosmos")
	if err != nil {
		t.Fatal(err)
	}
	if id != CosmosRootID {
		t.Errorf("cosmos ID = %q; want %q", id, CosmosRootID)
	}
}

func TestNewForTypeUnknown(t *testing.T) {
	if _, err := NewForType("not_a_real_type"); err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestIsValid(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"PRD_A7K3M2", true},
		{"REQ_X9P4N1", true},
		{"COS_root", true},
		{"prd_a7k3m2", false},  // lowercase prefix
		{"PRD-A7K3M2", false},  // hyphen
		{"PRD_A7K3M", false},   // 5-char suffix
		{"PRD_A7K3M2Z", false}, // 7-char suffix
		{"PROD_A7K3M2", false}, // 4-char prefix
		{"PR_A7K3M2", false},   // 2-char prefix
		{"PRD_A7I3M2", false},  // contains I (excluded)
		{"PRD_A7L3M2", false},  // contains L (excluded)
		{"PRD_A7O3M2", false},  // contains O (excluded)
		{"PRD_A7U3M2", false},  // contains U (excluded)
		{"", false},
		{"req-login-required", false},
		{"PI-ACC-MBX-EXAMPLE-001", false},
		{"COS_ROOT", false}, // only literal COS_root accepted
	}
	for _, c := range cases {
		if got := IsValid(c.id); got != c.want {
			t.Errorf("IsValid(%q) = %v; want %v", c.id, got, c.want)
		}
	}
}

func TestIsValidForType(t *testing.T) {
	if !IsValidForType("PRD_A7K3M2", "product_blueprint") {
		t.Error("PRD_A7K3M2 should be valid for product_blueprint")
	}
	if IsValidForType("PRD_A7K3M2", "service_blueprint") {
		t.Error("PRD_A7K3M2 should NOT be valid for service_blueprint")
	}
	if !IsValidForType("COS_root", "cosmos") {
		t.Error("COS_root should be valid for cosmos")
	}
	if IsValidForType("COS_root", "domain") {
		t.Error("COS_root should NOT be valid for domain")
	}
	if !IsValidForType("PRD_A7K3M2", "this_type_is_not_registered") {
		t.Error("unknown types should pass when format is valid")
	}
}

func TestIsLegacy(t *testing.T) {
	cases := []struct {
		id   string
		want bool
	}{
		{"", false},
		{"PRD_A7K3M2", false},
		{"COS_root", false},
		{"produkt-1", true},
		{"PI-ACC-MBX-EXAMPLE-001", true},
		{"DEC-001", true},
	}
	for _, c := range cases {
		if got := IsLegacy(c.id); got != c.want {
			t.Errorf("IsLegacy(%q) = %v; want %v", c.id, got, c.want)
		}
	}
}

func TestUniqueness(t *testing.T) {
	// 32^6 ≈ 1.07e9 Werte. Geburtstagsproblem: Kollisionswahrscheinlichkeit
	// für N=1000 Stichproben liegt bei ~5e-4 — testbar ohne Flakes.
	// (Größere Stichproben würden absichtsgemäß irgendwann kollidieren.)
	const N = 1000
	seen := make(map[string]bool, N)
	for i := 0; i < N; i++ {
		id, err := New("PRD")
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Fatalf("collision after %d generations: %s", i, id)
		}
		seen[id] = true
	}
}

func TestRandomCrockfordAlphabet(t *testing.T) {
	s, err := randomCrockford(5000)
	if err != nil {
		t.Fatal(err)
	}
	excluded := regexp.MustCompile(`[ILOU]`)
	if excluded.MatchString(s) {
		t.Error("Crockford string contains excluded letters I/L/O/U")
	}
}

func TestPrefixForType(t *testing.T) {
	if p, ok := PrefixForType("product_blueprint"); !ok || p != "PRD" {
		t.Errorf("PrefixForType(product_blueprint) = %q,%v; want PRD,true", p, ok)
	}
	if _, ok := PrefixForType("nope"); ok {
		t.Error("PrefixForType(nope) should be false")
	}
}
