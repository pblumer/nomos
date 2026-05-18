package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nomos/nomos/internal/storage"
)

func createDecisionTestCosmos(t *testing.T) string {
	t.Helper()
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Dir(storage.CosmosFile(p)), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Test Team\n"), 0o644))
	domainDir := filepath.Join(storage.DomainsDir(p), "governance.blumer.com")
	must(os.MkdirAll(filepath.Join(domainDir, "decisions", "DEC-001"), 0o755))
	must(os.WriteFile(filepath.Join(domainDir, "domain.yaml"), []byte("name: governance.blumer.com\nowner: Governance\nstatus: draft\n"), 0o644))
	must(os.WriteFile(filepath.Join(domainDir, "decisions", "DEC-001", "decision.yaml"), []byte(
		"id: DEC-001\ntype: decision\nname: Provisioning Eligibility\nversion: 0.1.0\nstatus: active\nowner: Governance Team\ndmn_file: decision.dmn\ninputs:\n  - name: employmentStatus\n    type: string\noutputs:\n  - name: canProvision\n    type: boolean\n"), 0o644))
	must(os.WriteFile(filepath.Join(domainDir, "decisions", "DEC-001", "decision.dmn"), []byte(`<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20240513/MODEL/" id="Definitions_1" name="Provisioning Eligibility">
  <inputData id="InputData_Status" name="employmentStatus">
    <variable id="Var_Status" name="employmentStatus" typeRef="string"/>
  </inputData>
  <decision id="Decision_1" name="Provisioning Eligibility">
    <variable id="Var_1" name="canProvision" typeRef="boolean"/>
    <informationRequirement id="IR_1"><requiredInput href="#InputData_Status"/></informationRequirement>
    <decisionTable id="DT_1" hitPolicy="FIRST">
      <input id="In_1" label="employmentStatus"><inputExpression id="IE_1" typeRef="string"><text>employmentStatus</text></inputExpression></input>
      <output id="Out_1" name="canProvision" typeRef="boolean"/>
      <rule id="R_1">
        <inputEntry id="IE_R1"><text>"terminated"</text></inputEntry>
        <outputEntry id="OE_R1"><text>false</text></outputEntry>
      </rule>
      <rule id="R_2">
        <inputEntry id="IE_R2"><text>"active"</text></inputEntry>
        <outputEntry id="OE_R2"><text>true</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`), 0o644))
	return p
}

func TestGetDecisionDefinitions_ReturnsParsedDRG(t *testing.T) {
	p := createDecisionTestCosmos(t)
	defs, err := GetDecisionDefinitions(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("GetDecisionDefinitions: %v", err)
	}
	if defs.Name != "Provisioning Eligibility" {
		t.Errorf("name: %q", defs.Name)
	}
	if len(defs.InputData) != 1 || defs.InputData[0].Name != "employmentStatus" {
		t.Errorf("input data: %+v", defs.InputData)
	}
	if len(defs.Decisions) != 1 {
		t.Fatalf("decisions: %d", len(defs.Decisions))
	}
	if got := defs.Decisions[0].InformationRequirements; len(got) != 1 || got[0].RequiredInput != "#InputData_Status" {
		t.Errorf("information requirements: %+v", got)
	}
	if defs.Decisions[0].Logic == nil || defs.Decisions[0].Logic.Kind != "decisionTable" {
		t.Errorf("logic: %+v", defs.Decisions[0].Logic)
	}
	if dt := defs.Decisions[0].Logic.DecisionTable; dt == nil || dt.HitPolicy != "FIRST" || len(dt.Rules) != 2 {
		t.Errorf("decision table: %+v", dt)
	}
}

func TestEvaluateDecision_AgainstDMNFile(t *testing.T) {
	p := createDecisionTestCosmos(t)
	res, err := EvaluateDecision(p, "governance.blumer.com", "DEC-001", EvaluateDecisionRequest{Inputs: map[string]any{"employmentStatus": "terminated"}})
	if err != nil {
		t.Fatalf("EvaluateDecision: %v", err)
	}
	if got := res.Outputs["canProvision"]; got != false {
		t.Errorf("expected false for terminated, got %v", got)
	}
}

func TestGetDecisionDMN_ReturnsRawXML(t *testing.T) {
	p := createDecisionTestCosmos(t)
	raw, err := GetDecisionDMN(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("GetDecisionDMN: %v", err)
	}
	if !contains(raw, "<decisionTable") || !contains(raw, "Provisioning Eligibility") {
		t.Errorf("unexpected DMN xml: %s", raw[:min(200, len(raw))])
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
