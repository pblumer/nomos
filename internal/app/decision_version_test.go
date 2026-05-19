package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/model"
	"github.com/nomos/nomos/internal/storage"
)

// versionTestDMN is a minimal but valid DMN file used to exercise
// UpdateDecisionDMN end-to-end (parse + I/O derivation + snapshot).
const versionTestDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <inputData id="id1" name="dns_name"><variable id="v1" name="dns_name" typeRef="string"/></inputData>
  <decision id="DEC-V1" name="DNS check">
    <variable id="vo" name="result" typeRef="boolean"/>
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1"><inputExpression typeRef="string"><text>dns_name</text></inputExpression></input>
      <output id="o1" name="result" typeRef="boolean"/>
      <rule id="r1"><inputEntry><text>"blumer.cloud"</text></inputEntry><outputEntry><text>true</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>false</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`

// versionTestDMNUpdated has an extra rule — semantically different from the
// original, so saving it must produce a fresh version snapshot.
const versionTestDMNUpdated = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <inputData id="id1" name="dns_name"><variable id="v1" name="dns_name" typeRef="string"/></inputData>
  <decision id="DEC-V1" name="DNS check">
    <variable id="vo" name="result" typeRef="boolean"/>
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1"><inputExpression typeRef="string"><text>dns_name</text></inputExpression></input>
      <output id="o1" name="result" typeRef="boolean"/>
      <rule id="r1"><inputEntry><text>"blumer.cloud"</text></inputEntry><outputEntry><text>true</text></outputEntry></rule>
      <rule id="rNew"><inputEntry><text>"example.com"</text></inputEntry><outputEntry><text>true</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>false</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`

func createVersionTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Dir(storage.CosmosFile(p)), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Team\n"), 0o644))
	domainDir := filepath.Join(storage.DomainsDir(p), "blumer.cloud")
	must(os.MkdirAll(domainDir, 0o755))
	must(os.WriteFile(filepath.Join(domainDir, "domain.yaml"), []byte("name: blumer.cloud\nowner: Team\nstatus: draft\n"), 0o644))
	return p
}

func TestCreateDecision_WritesInitialSnapshot(t *testing.T) {
	p := createVersionTestCosmos(t)
	_, err := CreateDecision(p, "blumer.cloud", CreateDecisionRequest{ID: "DEC-V1", Name: "DNS check"})
	if err != nil {
		t.Fatalf("CreateDecision: %v", err)
	}
	snap := filepath.Join(storage.DomainsDir(p), "blumer.cloud", "decisions", "DEC-V1", "versions", "v0.1.0", "decision.yaml")
	if _, err := os.Stat(snap); err != nil {
		t.Fatalf("expected initial snapshot at %s: %v", snap, err)
	}
}

func TestUpdateDecisionDMN_BumpsAndSnapshots(t *testing.T) {
	p := createVersionTestCosmos(t)
	if _, err := CreateDecision(p, "blumer.cloud", CreateDecisionRequest{ID: "DEC-V1", Name: "DNS check"}); err != nil {
		t.Fatal(err)
	}
	// First DMN attach: bump 0.1.0 → 0.1.1, snapshot directory contains DMN.
	dto1, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMN)
	if err != nil {
		t.Fatalf("UpdateDecisionDMN: %v", err)
	}
	if dto1.Version != "0.1.1" {
		t.Fatalf("after first DMN attach: version = %q, want 0.1.1", dto1.Version)
	}
	if _, err := os.Stat(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "decisions", "DEC-V1", "versions", "v0.1.1", "decision.dmn")); err != nil {
		t.Fatalf("expected v0.1.1 snapshot DMN: %v", err)
	}
	// Saving the same XML again is a no-op — same bytes, same metadata.
	dto2, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMN)
	if err != nil {
		t.Fatalf("idempotent UpdateDecisionDMN: %v", err)
	}
	if dto2.Version != "0.1.1" {
		t.Fatalf("no-op save bumped version: %q", dto2.Version)
	}
	if _, err := os.Stat(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "decisions", "DEC-V1", "versions", "v0.1.2")); err == nil {
		t.Fatal("no-op save created a v0.1.2 snapshot")
	}
	// Real change: bump again.
	dto3, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMNUpdated)
	if err != nil {
		t.Fatalf("UpdateDecisionDMN updated: %v", err)
	}
	if dto3.Version != "0.1.2" {
		t.Fatalf("after material change: version = %q, want 0.1.2", dto3.Version)
	}
	if _, err := os.Stat(filepath.Join(storage.DomainsDir(p), "blumer.cloud", "decisions", "DEC-V1", "versions", "v0.1.2", "decision.dmn")); err != nil {
		t.Fatalf("expected v0.1.2 snapshot DMN: %v", err)
	}
}

func TestUpdateDecision_BumpsOnMetadataChange(t *testing.T) {
	p := createVersionTestCosmos(t)
	if _, err := CreateDecision(p, "blumer.cloud", CreateDecisionRequest{ID: "DEC-V1", Name: "DNS check", Owner: "alice"}); err != nil {
		t.Fatal(err)
	}
	// Status flip is a material change — auto-bump must fire.
	dto, err := UpdateDecision(p, "blumer.cloud", "DEC-V1", UpdateDecisionRequest{Status: "active"})
	if err != nil {
		t.Fatalf("UpdateDecision: %v", err)
	}
	if dto.Version != "0.1.1" {
		t.Fatalf("version = %q, want 0.1.1", dto.Version)
	}
	// No-op update returns current state unchanged.
	dto, err = UpdateDecision(p, "blumer.cloud", "DEC-V1", UpdateDecisionRequest{})
	if err != nil {
		t.Fatalf("UpdateDecision no-op: %v", err)
	}
	if dto.Version != "0.1.1" {
		t.Fatalf("no-op update bumped: %q", dto.Version)
	}
}

func TestTrace_ResolvesAgainstHistoricalVersion(t *testing.T) {
	p := createVersionTestCosmos(t)
	if _, err := CreateDecision(p, "blumer.cloud", CreateDecisionRequest{ID: "DEC-V1", Name: "DNS check"}); err != nil {
		t.Fatal(err)
	}
	dto1, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMN)
	if err != nil {
		t.Fatal(err)
	}
	// Evaluate against the first DMN version.
	_, trace1, err := EvaluateDecisionWithTrace(p, "blumer.cloud", "DEC-V1",
		EvaluateDecisionRequest{Inputs: map[string]any{"dns_name": "blumer.cloud"}},
		model.Evaluator{ID: "test"})
	if err != nil {
		t.Fatalf("evaluate v1: %v", err)
	}
	if trace1.DecisionVersion != dto1.Version {
		t.Fatalf("trace.decision_version = %q, want %q", trace1.DecisionVersion, dto1.Version)
	}
	// Mutate the DMN — HEAD now has 3 rules, but trace1 was recorded at 2 rules.
	dto2, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMNUpdated)
	if err != nil {
		t.Fatal(err)
	}
	if dto2.Version == dto1.Version {
		t.Fatal("DMN change did not bump version")
	}
	// Historical defs must point at the original two-rule snapshot.
	defsOld, err := GetDecisionVersionDefinitions(p, "blumer.cloud", "DEC-V1", trace1.DecisionVersion)
	if err != nil {
		t.Fatalf("GetDecisionVersionDefinitions: %v", err)
	}
	if len(defsOld.Decisions) != 1 || defsOld.Decisions[0].Logic == nil || defsOld.Decisions[0].Logic.DecisionTable == nil {
		t.Fatalf("historical defs missing table: %+v", defsOld)
	}
	if got := len(defsOld.Decisions[0].Logic.DecisionTable.Rules); got != 2 {
		t.Fatalf("historical version has %d rules, want 2 (HEAD has 3)", got)
	}
	// And current HEAD defs reflect the new state.
	defsNew, err := GetDecisionDefinitions(p, "blumer.cloud", "DEC-V1")
	if err != nil {
		t.Fatalf("GetDecisionDefinitions: %v", err)
	}
	if got := len(defsNew.Decisions[0].Logic.DecisionTable.Rules); got != 3 {
		t.Fatalf("HEAD has %d rules, want 3", got)
	}
}

func TestListDecisionVersions(t *testing.T) {
	p := createVersionTestCosmos(t)
	if _, err := CreateDecision(p, "blumer.cloud", CreateDecisionRequest{ID: "DEC-V1", Name: "DNS check"}); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMN); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateDecisionDMN(p, "blumer.cloud", "DEC-V1", versionTestDMNUpdated); err != nil {
		t.Fatal(err)
	}
	list, err := ListDecisionVersions(p, "blumer.cloud", "DEC-V1")
	if err != nil {
		t.Fatalf("ListDecisionVersions: %v", err)
	}
	if list.Current != "0.1.2" {
		t.Fatalf("current = %q, want 0.1.2", list.Current)
	}
	if len(list.Items) != 3 {
		t.Fatalf("got %d versions, want 3", len(list.Items))
	}
	want := []string{"0.1.0", "0.1.1", "0.1.2"}
	for i, v := range want {
		if list.Items[i].Version != v {
			t.Fatalf("versions[%d] = %q, want %q", i, list.Items[i].Version, v)
		}
	}
}

func TestBumpPatch(t *testing.T) {
	cases := []struct{ in, out string }{
		{"0.1.0", "0.1.1"},
		{"1.2.3", "1.2.4"},
		{"v0.1.0", "0.1.1"},
		{"", "0.1.0"},
		{"weird", "weird.1"},
	}
	for _, c := range cases {
		if got := bumpPatch(c.in); got != c.out {
			t.Errorf("bumpPatch(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestGetDecisionVersion_LegacyDecisionWithoutSnapshot(t *testing.T) {
	// Simulate a decision created before snapshotting existed: write the head
	// files by hand without a versions/ directory and verify the HEAD version
	// still resolves through GetDecisionVersionDMN.
	p := createVersionTestCosmos(t)
	dir := filepath.Join(storage.DomainsDir(p), "blumer.cloud", "decisions", "DEC-LEG")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "decision.yaml"), []byte("id: DEC-LEG\ntype: decision\nname: Legacy\nversion: 0.5.0\nstatus: active\nowner: Team\ndmn_file: decision.dmn\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "decision.dmn"), []byte(versionTestDMN), 0o644); err != nil {
		t.Fatal(err)
	}
	xml, err := GetDecisionVersionDMN(p, "blumer.cloud", "DEC-LEG", "0.5.0")
	if err != nil {
		t.Fatalf("HEAD fallback failed: %v", err)
	}
	if !strings.Contains(xml, "decisionTable") {
		t.Fatalf("unexpected XML: %s", xml[:min(len(xml), 80)])
	}
	if _, err := GetDecisionVersionDMN(p, "blumer.cloud", "DEC-LEG", "0.4.0"); err == nil {
		t.Fatal("expected 404 for unknown historical version, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
