package app

import (
	"os"
	"path/filepath"
	"strings"
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

// Cosmos Explorer's "Inputs/Outputs (aus Diagramm)" panel derives I/O from
// the DMN file. Many DMN editors only let modellers pick types on the
// decision table's input columns (the <inputData> variable stays typeless)
// and routinely leave the decision <variable> stale after renaming or
// retyping the table's output column. Both situations must be recovered
// from the table itself.
func TestUpdateDecisionDMN_DerivesIOFromDecisionTable(t *testing.T) {
	p := createDecisionTestCosmos(t)
	dmnXML := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20240513/MODEL/" id="Defs" name="Kreditpruefungsstufe evaluieren">
  <inputData id="ID_value" name="value"><variable id="V_value" name="value"/></inputData>
  <inputData id="ID_status" name="customer_status"><variable id="V_status" name="customer_status"/></inputData>
  <decision id="Dec_1" name="Kreditpruefungsstufe evaluieren">
    <variable id="Var_Dec1" name="result" typeRef="boolean"/>
    <informationRequirement id="IR_1"><requiredInput href="#ID_value"/></informationRequirement>
    <informationRequirement id="IR_2"><requiredInput href="#ID_status"/></informationRequirement>
    <decisionTable id="DT_1" hitPolicy="FIRST">
      <input id="In_value" label="value"><inputExpression id="IE_value" typeRef="number"><text>value</text></inputExpression></input>
      <input id="In_status" label="customer_status"><inputExpression id="IE_status" typeRef="string"><text>customer_status</text></inputExpression></input>
      <output id="Out_1" name="kredit_stufe" typeRef="string"/>
      <rule id="R_1">
        <inputEntry id="IE_R1_1"><text>&lt; 100</text></inputEntry>
        <inputEntry id="IE_R1_2"><text>-</text></inputEntry>
        <outputEntry id="OE_R1"><text>"kleiner Betrag"</text></outputEntry>
      </rule>
      <rule id="R_2">
        <inputEntry id="IE_R2_1"><text>-</text></inputEntry>
        <inputEntry id="IE_R2_2"><text>-</text></inputEntry>
        <outputEntry id="OE_R2"><text>"grosser Betrag"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	dec, err := UpdateDecisionDMN(p, "governance.blumer.com", "DEC-001", dmnXML)
	if err != nil {
		t.Fatalf("UpdateDecisionDMN: %v", err)
	}
	wantInputs := map[string]string{"value": "number", "customer_status": "string"}
	if len(dec.Inputs) != len(wantInputs) {
		t.Fatalf("inputs: got %+v, want %v", dec.Inputs, wantInputs)
	}
	for _, in := range dec.Inputs {
		if wantInputs[in.Name] != in.Type {
			t.Errorf("input %q: got type %q, want %q", in.Name, in.Type, wantInputs[in.Name])
		}
	}
	if len(dec.Outputs) != 1 || dec.Outputs[0].Name != "kredit_stufe" || dec.Outputs[0].Type != "string" {
		t.Errorf("outputs: got %+v, want [{kredit_stufe string}]", dec.Outputs)
	}
}

// When the modeller saves a DMN with descriptive pill labels ("Customer
// Status"), the server normalizes the FEEL-callable identifiers
// (variable.name and the decision-table column references) to snake_case
// so the pill stays human-readable while the table runs on the machine name.
func TestUpdateDecisionDMN_NormalizesDescriptiveInputNames(t *testing.T) {
	p := createDecisionTestCosmos(t)
	dmnXML := `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="Defs" name="K">
  <inputData id="ID_value" name="Value"><variable id="V_value" name="Value"/></inputData>
  <inputData id="ID_status" name="Customer Status"><variable id="V_status" name="Customer Status"/></inputData>
  <decision id="Dec_1" name="K">
    <variable id="Var_Dec1" name="kredit_stufe" typeRef="string"/>
    <decisionTable id="DT_1" hitPolicy="FIRST">
      <input id="In_value" label="Value"><inputExpression id="IE_value" typeRef="number"><text>Value</text></inputExpression></input>
      <input id="In_status" label="Customer Status"><inputExpression id="IE_status" typeRef="string"><text>Customer Status</text></inputExpression></input>
      <output id="Out_1" name="kredit_stufe" typeRef="string"/>
      <rule id="R_1">
        <inputEntry id="IE_R1_1"><text>-</text></inputEntry>
        <inputEntry id="IE_R1_2"><text>-</text></inputEntry>
        <outputEntry id="OE_R1"><text>"x"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	dec, err := UpdateDecisionDMN(p, "governance.blumer.com", "DEC-001", dmnXML)
	if err != nil {
		t.Fatalf("UpdateDecisionDMN: %v", err)
	}
	wantInputs := map[string]string{"value": "number", "customer_status": "string"}
	if len(dec.Inputs) != len(wantInputs) {
		t.Fatalf("inputs: got %+v, want %v", dec.Inputs, wantInputs)
	}
	for _, in := range dec.Inputs {
		if wantInputs[in.Name] != in.Type {
			t.Errorf("input %q: got type %q, want %q", in.Name, in.Type, wantInputs[in.Name])
		}
	}
	// Persisted DMN bytes carry the normalized variable names and the
	// pill labels remain untouched.
	raw, err := GetDecisionDMN(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("GetDecisionDMN: %v", err)
	}
	wantSubstrings := []string{
		`<inputData id="ID_status" name="Customer Status">`,
		`name="customer_status"`,
		`<text>customer_status</text>`,
		`<text>value</text>`,
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(raw, want) {
			t.Errorf("persisted DMN missing %q\n--- DMN ---\n%s", want, raw)
		}
	}
	if strings.Contains(raw, `<text>Customer Status</text>`) {
		t.Errorf("persisted DMN still references descriptive label in column text\n--- DMN ---\n%s", raw)
	}
}

// GetDecision re-derives inputs from the DMN file at runtime, so a decision
// whose decision.yaml predates the inputsFromDMN fix still reflects the
// column-level typeRef without forcing the user to re-save the DMN.
func TestGetDecision_ReDerivesInputTypeFromDMNColumn(t *testing.T) {
	p := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.MkdirAll(filepath.Dir(storage.CosmosFile(p)), 0o755))
	must(os.WriteFile(storage.CosmosFile(p), []byte("id: cosmos-local\nname: Local Cosmos\nversion: 0.1.0\nstatus: draft\nowner: Test Team\n"), 0o644))
	domainDir := filepath.Join(storage.DomainsDir(p), "blumer.com")
	must(os.MkdirAll(filepath.Join(domainDir, "decisions", "DEC-002"), 0o755))
	must(os.WriteFile(filepath.Join(domainDir, "domain.yaml"), []byte("name: blumer.com\nowner: Test\nstatus: draft\n"), 0o644))
	// Stale cache: decision.yaml records no input type at all.
	must(os.WriteFile(filepath.Join(domainDir, "decisions", "DEC-002", "decision.yaml"), []byte(
		"id: DEC-002\ntype: decision\nname: Kreditpruefung\nversion: 0.1.0\nstatus: draft\nowner: Test\ndmn_file: decision.dmn\ninputs:\n  - name: value\noutputs:\n  - name: kredit_stufe\n    type: string\n"), 0o644))
	// DMN: <inputData> variable has no typeRef (bpmn.io default); the
	// decision-table column carries the authoritative typeRef="number".
	must(os.WriteFile(filepath.Join(domainDir, "decisions", "DEC-002", "decision.dmn"), []byte(`<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20240513/MODEL/" id="Definitions_2" name="Kreditpruefung">
  <inputData id="InputData_Value" name="value"><variable id="Var_Value" name="value"/></inputData>
  <decision id="Decision_1" name="Kreditpruefung">
    <variable id="Var_1" name="kredit_stufe" typeRef="string"/>
    <informationRequirement id="IR_1"><requiredInput href="#InputData_Value"/></informationRequirement>
    <decisionTable id="DT_1" hitPolicy="FIRST">
      <input id="In_1" label="value"><inputExpression id="IE_1" typeRef="number"><text>value</text></inputExpression></input>
      <output id="Out_1" name="kredit_stufe" typeRef="string"/>
      <rule id="R_1"><inputEntry id="IE_R1"><text>-</text></inputEntry><outputEntry id="OE_R1"><text>"x"</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`), 0o644))

	dto, err := GetDecision(p, "blumer.com", "DEC-002")
	if err != nil {
		t.Fatalf("GetDecision: %v", err)
	}
	if len(dto.Inputs) != 1 || dto.Inputs[0].Type != "number" {
		t.Errorf("expected inputs[0].Type=number from DMN column, got %+v", dto.Inputs)
	}
}

func TestGetDecisionDMN_ReturnsRawXML(t *testing.T) {
	p := createDecisionTestCosmos(t)
	raw, err := GetDecisionDMN(p, "governance.blumer.com", "DEC-001")
	if err != nil {
		t.Fatalf("GetDecisionDMN: %v", err)
	}
	if !strings.Contains(raw, "<decisionTable") || !strings.Contains(raw, "Provisioning Eligibility") {
		head := raw
		if len(head) > 200 {
			head = head[:200]
		}
		t.Errorf("unexpected DMN xml: %s", head)
	}
}
