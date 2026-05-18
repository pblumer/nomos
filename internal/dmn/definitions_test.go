package dmn_test

import (
	"strings"
	"testing"

	"github.com/nomos/nomos/internal/dmn"
)

const provisioningEligibilityDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20240513/MODEL/" xmlns:dmndi="https://www.omg.org/spec/DMN/20230324/DMNDI/" xmlns:dc="http://www.omg.org/spec/DMN/20180521/DC/" xmlns:di="http://www.omg.org/spec/DMN/20180521/DI/"
  id="Definitions_ProvisioningEligibility" name="Provisioning Eligibility"
  namespace="https://nomos.local/dmn/provisioning-eligibility" exporter="Nomos" exporterVersion="0.1.0">
  <itemDefinition id="ItemDef_Employment" name="EmploymentStatus" isCollection="false">
    <typeRef>string</typeRef>
    <allowedValues><text>"active","onLeave","terminated"</text></allowedValues>
  </itemDefinition>
  <inputData id="InputData_Status" name="employmentStatus">
    <variable id="Var_Status" name="employmentStatus" typeRef="string"/>
  </inputData>
  <inputData id="InputData_Risk" name="riskScore">
    <variable id="Var_Risk" name="riskScore" typeRef="number"/>
  </inputData>
  <knowledgeSource id="KS_Policy" name="HR Provisioning Policy">
    <type>policy</type>
    <locationURI>https://policies.example.com/hr/provisioning</locationURI>
  </knowledgeSource>
  <businessKnowledgeModel id="BKM_RiskBand" name="riskBand">
    <variable id="Var_RiskBand" name="riskBand" typeRef="string"/>
    <encapsulatedLogic>
      <formalParameter id="P_Score" name="score" typeRef="number"/>
      <literalExpression id="LE_RiskBand"><text>if score &lt; 30 then "low" else if score &lt; 70 then "medium" else "high"</text></literalExpression>
    </encapsulatedLogic>
  </businessKnowledgeModel>
  <decision id="Decision_Eligibility" name="Provisioning Eligibility">
    <question>Should this employee be provisioned?</question>
    <allowedAnswers>true, false</allowedAnswers>
    <variable id="Var_Eligibility" name="canProvision" typeRef="boolean"/>
    <informationRequirement id="IR_Status"><requiredInput href="#InputData_Status"/></informationRequirement>
    <informationRequirement id="IR_Risk"><requiredInput href="#InputData_Risk"/></informationRequirement>
    <knowledgeRequirement id="KR_RiskBand"><requiredKnowledge href="#BKM_RiskBand"/></knowledgeRequirement>
    <authorityRequirement id="AR_Policy"><requiredAuthority href="#KS_Policy"/></authorityRequirement>
    <decisionTable id="DT_Eligibility" hitPolicy="FIRST" outputLabel="canProvision">
      <input id="In_Status" label="employmentStatus">
        <inputExpression id="IE_Status" typeRef="string"><text>employmentStatus</text></inputExpression>
      </input>
      <input id="In_Risk" label="riskBand">
        <inputExpression id="IE_Risk" typeRef="string"><text>riskBand(riskScore)</text></inputExpression>
      </input>
      <output id="Out_Decision" name="canProvision" typeRef="boolean"/>
      <annotation name="rationale"/>
      <rule id="Rule_1">
        <inputEntry id="IE_R1_1"><text>"terminated"</text></inputEntry>
        <inputEntry id="IE_R1_2"><text>-</text></inputEntry>
        <outputEntry id="OE_R1_1"><text>false</text></outputEntry>
        <annotationEntry><text>Terminated employees never get accounts</text></annotationEntry>
      </rule>
      <rule id="Rule_2">
        <inputEntry id="IE_R2_1"><text>"active","onLeave"</text></inputEntry>
        <inputEntry id="IE_R2_2"><text>"high"</text></inputEntry>
        <outputEntry id="OE_R2_1"><text>false</text></outputEntry>
        <annotationEntry><text>High risk requires manual approval</text></annotationEntry>
      </rule>
      <rule id="Rule_3">
        <inputEntry id="IE_R3_1"><text>"active","onLeave"</text></inputEntry>
        <inputEntry id="IE_R3_2"><text>"low","medium"</text></inputEntry>
        <outputEntry id="OE_R3_1"><text>true</text></outputEntry>
        <annotationEntry><text>Standard provisioning path</text></annotationEntry>
      </rule>
    </decisionTable>
  </decision>
  <dmndi:DMNDI>
    <dmndi:DMNDiagram id="DMNDiagram_1">
      <dmndi:DMNShape id="Shape_Decision_Eligibility" dmnElementRef="Decision_Eligibility">
        <dc:Bounds x="360" y="80" width="180" height="80"/>
      </dmndi:DMNShape>
      <dmndi:DMNShape id="Shape_InputData_Status" dmnElementRef="InputData_Status">
        <dc:Bounds x="200" y="240" width="125" height="45"/>
      </dmndi:DMNShape>
      <dmndi:DMNEdge id="Edge_IR_Status" dmnElementRef="IR_Status">
        <di:waypoint x="262" y="240"/>
        <di:waypoint x="450" y="160"/>
      </dmndi:DMNEdge>
    </dmndi:DMNDiagram>
  </dmndi:DMNDI>
</definitions>`

func TestParseDefinitions_FullDRG(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(provisioningEligibilityDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	if defs.Name != "Provisioning Eligibility" {
		t.Errorf("definitions name: got %q", defs.Name)
	}
	if len(defs.ItemDefinitions) != 1 || defs.ItemDefinitions[0].Name != "EmploymentStatus" {
		t.Errorf("item definitions: %+v", defs.ItemDefinitions)
	}
	if got := defs.ItemDefinitions[0].AllowedValues; len(got) != 1 || !strings.Contains(got[0], "terminated") {
		t.Errorf("allowed values: %+v", got)
	}
	if len(defs.InputData) != 2 {
		t.Fatalf("input data count: got %d", len(defs.InputData))
	}
	if defs.InputData[0].Variable.TypeRef != "string" {
		t.Errorf("input variable typeRef: %q", defs.InputData[0].Variable.TypeRef)
	}
	if len(defs.KnowledgeSource) != 1 || defs.KnowledgeSource[0].Type != "policy" {
		t.Errorf("knowledge source: %+v", defs.KnowledgeSource)
	}
	if len(defs.BKMs) != 1 || defs.BKMs[0].Name != "riskBand" {
		t.Errorf("BKM: %+v", defs.BKMs)
	}
	if defs.BKMs[0].EncapsulatedLogic == nil || defs.BKMs[0].EncapsulatedLogic.Body == nil ||
		defs.BKMs[0].EncapsulatedLogic.Body.Kind != "literalExpression" {
		t.Errorf("BKM encapsulated logic: %+v", defs.BKMs[0].EncapsulatedLogic)
	}
	if len(defs.Decisions) != 1 {
		t.Fatalf("decision count: got %d", len(defs.Decisions))
	}
	d := defs.Decisions[0]
	if d.Question == "" || d.AllowedAnswers == "" {
		t.Errorf("decision Q&A: %q / %q", d.Question, d.AllowedAnswers)
	}
	if len(d.InformationRequirements) != 2 ||
		d.InformationRequirements[0].RequiredInput != "#InputData_Status" {
		t.Errorf("information requirements: %+v", d.InformationRequirements)
	}
	if len(d.KnowledgeRequirements) != 1 ||
		d.KnowledgeRequirements[0].RequiredKnowledge != "#BKM_RiskBand" {
		t.Errorf("knowledge requirements: %+v", d.KnowledgeRequirements)
	}
	if len(d.AuthorityRequirements) != 1 ||
		d.AuthorityRequirements[0].RequiredAuthority != "#KS_Policy" {
		t.Errorf("authority requirements: %+v", d.AuthorityRequirements)
	}
	if d.Logic == nil || d.Logic.Kind != "decisionTable" || d.Logic.DecisionTable == nil {
		t.Fatalf("decision logic: %+v", d.Logic)
	}
	dt := d.Logic.DecisionTable
	if dt.HitPolicy != "FIRST" {
		t.Errorf("hit policy: %q", dt.HitPolicy)
	}
	if len(dt.Inputs) != 2 || dt.Inputs[1].Expression != "riskBand(riskScore)" {
		t.Errorf("inputs: %+v", dt.Inputs)
	}
	if len(dt.Outputs) != 1 || dt.Outputs[0].Name != "canProvision" || dt.Outputs[0].TypeRef != "boolean" {
		t.Errorf("outputs: %+v", dt.Outputs)
	}
	if len(dt.Annotations) != 1 || dt.Annotations[0].Name != "rationale" {
		t.Errorf("annotations: %+v", dt.Annotations)
	}
	if len(dt.Rules) != 3 {
		t.Fatalf("rules: got %d", len(dt.Rules))
	}
	if got := dt.Rules[1].InputEntries[1]; got != `"high"` {
		t.Errorf("rule 2 risk band entry: %q", got)
	}
	if got := dt.Rules[2].Annotations; len(got) != 1 || !strings.Contains(got[0], "Standard") {
		t.Errorf("rule annotation: %+v", got)
	}
	if len(defs.Diagrams) != 1 || len(defs.Diagrams[0].Shapes) != 2 || len(defs.Diagrams[0].Edges) != 1 {
		t.Errorf("diagrams: %+v", defs.Diagrams)
	}
	if defs.Diagrams[0].Shapes[0].Width != 180 {
		t.Errorf("shape width: %v", defs.Diagrams[0].Shapes[0].Width)
	}
}

func TestParseDefinitions_LiteralExpressionDecision(t *testing.T) {
	const literal = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20240513/MODEL/" id="d" name="d">
  <decision id="Dec_1" name="Greeting">
    <variable id="V1" name="greeting" typeRef="string"/>
    <literalExpression id="LE_1"><text>"Hello, " + name</text></literalExpression>
  </decision>
</definitions>`
	defs, err := dmn.ParseDefinitions([]byte(literal))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	if defs.Decisions[0].Logic.Kind != "literalExpression" {
		t.Fatalf("logic kind: %q", defs.Decisions[0].Logic.Kind)
	}
	if defs.Decisions[0].Logic.LiteralExpression.Text != `"Hello, " + name` {
		t.Errorf("FEEL text: %q", defs.Decisions[0].Logic.LiteralExpression.Text)
	}
}

func TestParseDefinitions_LegacyParseDMNStillWorks(t *testing.T) {
	t.Helper()
	// Confirm the older table-only ParseDMN keeps reading the same DMN file
	// so the FEEL evaluator pipeline is not broken by the new structs.
	tbl, err := dmn.ParseDMN([]byte(provisioningEligibilityDMN))
	if err != nil {
		t.Fatalf("ParseDMN: %v", err)
	}
	if tbl.HitPolicy != "FIRST" || len(tbl.Rules) != 3 || len(tbl.Inputs) != 2 {
		t.Errorf("legacy parse: %+v", tbl)
	}
}
