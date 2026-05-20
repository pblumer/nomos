package app

import "testing"

func TestBuildIDIndexAndResolve(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := CreateDecision(p, CreateDecisionRequest{Name: "Seed", ID: "DEC-SEED-1"}); err != nil {
		t.Fatal(err)
	}
	idx, err := BuildIDIndex(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(idx.Entries) == 0 {
		t.Fatal("expected index entries")
	}
	// createAppTestCosmos has flat services; service IDs
	// may be empty, so assert resolution works for an entry that has an ID.
	var sample IndexEntryDTO
	for _, e := range idx.Entries {
		if e.ID != "" {
			sample = e
			break
		}
	}
	if sample.ID == "" {
		t.Skip("no ID-bearing artifacts in fixture")
	}
	got, err := ResolveID(p, sample.ID)
	if err != nil || got.Address != sample.Address {
		t.Fatalf("resolve %s = %+v err %v", sample.ID, got, err)
	}
	if _, err := ResolveID(p, "does-not-exist"); err == nil {
		t.Fatal("expected not-found for unknown id")
	}
}

func TestIDIndexResolvesDecisionByID(t *testing.T) {
	p := createAppTestCosmos(t)
	if _, err := CreateDecision(p, CreateDecisionRequest{Name: "Rule", ID: "DEC-IDX-1"}); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveID(p, "DEC-IDX-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != "decision" || got.Address != "decisions/DEC-IDX-1" {
		t.Fatalf("unexpected resolution: %+v", got)
	}
}
