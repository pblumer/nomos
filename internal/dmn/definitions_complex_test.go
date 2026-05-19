package dmn_test

import (
	"testing"

	"github.com/nomos/nomos/internal/dmn"
)

// complexExprDMN exercises context, invocation, list, relation, conditional,
// for, every, some, filter expressions, decision services (hrefs), and
// mapItemDefinition with anonymous type structures.
const complexExprDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20240513/MODEL/"
  id="Definitions_Complex" name="Complex Expressions"
  namespace="https://nomos.local/dmn/complex">

  <!-- itemDefinition with item components and allowed values -->
  <itemDefinition id="ID_Person" name="tPerson" isCollection="false">
    <itemComponent id="IC_Name" name="name" isCollection="false">
      <typeRef>string</typeRef>
    </itemComponent>
    <itemComponent id="IC_Age" name="age" isCollection="false">
      <typeRef>number</typeRef>
      <allowedValues><text>[0..120]</text></allowedValues>
    </itemComponent>
  </itemDefinition>

  <!-- Decision service with hrefs -->
  <decisionService id="DS_Service1" name="Service1">
    <variable id="DS_Var1" name="Service1" typeRef="string"/>
    <outputDecision href="#Dec_Context"/>
    <encapsulatedDecision href="#Dec_Invocation"/>
    <inputDecision href="#Dec_List"/>
    <inputData href="#InputData_X"/>
  </decisionService>

  <inputData id="InputData_X" name="xInput">
    <variable id="Var_X" name="xInput" typeRef="number"/>
  </inputData>

  <!-- Context expression -->
  <decision id="Dec_Context" name="Context Decision">
    <variable id="Var_Ctx" name="Context Decision" typeRef="context"/>
    <context id="Ctx_1">
      <contextEntry id="CE_1">
        <variable id="V_CE_1" name="result" typeRef="string"/>
        <literalExpression id="LE_CE_1">
          <text>"hello"</text>
        </literalExpression>
      </contextEntry>
      <contextEntry id="CE_2">
        <literalExpression id="LE_CE_2">
          <text>42</text>
        </literalExpression>
      </contextEntry>
    </context>
  </decision>

  <!-- Invocation expression -->
  <decision id="Dec_Invocation" name="Invocation Decision">
    <variable id="Var_Inv" name="Invocation Decision" typeRef="string"/>
    <invocation id="Inv_1">
      <literalExpression id="LE_Fn"><text>myFunction</text></literalExpression>
      <binding>
        <parameter id="P_1" name="param1" typeRef="string"/>
        <literalExpression id="LE_P1"><text>"arg1"</text></literalExpression>
      </binding>
    </invocation>
  </decision>

  <!-- List expression -->
  <decision id="Dec_List" name="List Decision">
    <variable id="Var_List" name="List Decision" typeRef="list"/>
    <list id="L_1">
      <literalExpression id="LE_L1"><text>1</text></literalExpression>
      <literalExpression id="LE_L2"><text>2</text></literalExpression>
      <context id="Ctx_L1">
        <contextEntry id="CE_L1">
          <literalExpression id="LE_CL1"><text>"a"</text></literalExpression>
        </contextEntry>
      </context>
    </list>
  </decision>

  <!-- Relation expression -->
  <decision id="Dec_Relation" name="Relation Decision">
    <variable id="Var_Rel" name="Relation Decision" typeRef="list"/>
    <relation id="Rel_1">
      <column id="Col_1" name="name" typeRef="string"/>
      <column id="Col_2" name="score" typeRef="number"/>
      <row>
        <literalExpression id="LE_R1"><text>"Alice"</text></literalExpression>
        <literalExpression id="LE_R2"><text>95</text></literalExpression>
      </row>
    </relation>
  </decision>

  <!-- Conditional expression -->
  <decision id="Dec_Conditional" name="Conditional Decision">
    <variable id="Var_Cond" name="Conditional Decision" typeRef="string"/>
    <conditional id="Cond_1">
      <if>
        <literalExpression id="LE_If"><text>x > 0</text></literalExpression>
      </if>
      <then>
        <literalExpression id="LE_Then"><text>"positive"</text></literalExpression>
      </then>
      <else>
        <literalExpression id="LE_Else"><text>"non-positive"</text></literalExpression>
      </else>
    </conditional>
  </decision>

  <!-- For iterator expression -->
  <decision id="Dec_For" name="For Decision">
    <variable id="Var_For" name="For Decision" typeRef="list"/>
    <for id="For_1" iteratorVariable="item">
      <in>
        <literalExpression id="LE_ForIn"><text>[1,2,3]</text></literalExpression>
      </in>
      <return>
        <literalExpression id="LE_ForReturn"><text>item * 2</text></literalExpression>
      </return>
    </for>
  </decision>

  <!-- Every quantifier -->
  <decision id="Dec_Every" name="Every Decision">
    <variable id="Var_Every" name="Every Decision" typeRef="boolean"/>
    <every id="Every_1" iteratorVariable="item">
      <in>
        <literalExpression id="LE_EvIn"><text>[1,2,3]</text></literalExpression>
      </in>
      <satisfies>
        <literalExpression id="LE_EvSat"><text>item > 0</text></literalExpression>
      </satisfies>
    </every>
  </decision>

  <!-- Some quantifier -->
  <decision id="Dec_Some" name="Some Decision">
    <variable id="Var_Some" name="Some Decision" typeRef="boolean"/>
    <some id="Some_1" iteratorVariable="item">
      <in>
        <literalExpression id="LE_SomeIn"><text>[1,2,3]</text></literalExpression>
      </in>
      <satisfies>
        <literalExpression id="LE_SomeSat"><text>item > 2</text></literalExpression>
      </satisfies>
    </some>
  </decision>

  <!-- Filter expression -->
  <decision id="Dec_Filter" name="Filter Decision">
    <variable id="Var_Filter" name="Filter Decision" typeRef="list"/>
    <filter id="Filter_1">
      <in>
        <literalExpression id="LE_FiltIn"><text>[1,2,3,4,5]</text></literalExpression>
      </in>
      <match>
        <literalExpression id="LE_FiltMatch"><text>item > 2</text></literalExpression>
      </match>
    </filter>
  </decision>
</definitions>`

func TestParseDefinitions_ComplexExpressions(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(complexExprDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}

	// Verify decisions parsed
	if len(defs.Decisions) == 0 {
		t.Error("expected decisions to be parsed")
	}

	decByID := map[string]struct{ Kind string }{}
	for _, d := range defs.Decisions {
		if d.Logic != nil {
			decByID[d.ID] = struct{ Kind string }{d.Logic.Kind}
		}
	}

	cases := []struct {
		id   string
		kind string
	}{
		{"Dec_Context", "context"},
		{"Dec_Invocation", "invocation"},
		{"Dec_List", "list"},
		{"Dec_Relation", "relation"},
		{"Dec_Conditional", "conditional"},
		{"Dec_For", "for"},
		{"Dec_Every", "every"},
		{"Dec_Some", "some"},
		{"Dec_Filter", "filter"},
	}
	for _, c := range cases {
		got, ok := decByID[c.id]
		if !ok {
			t.Errorf("decision %q not found in parsed definitions", c.id)
			continue
		}
		if got.Kind != c.kind {
			t.Errorf("decision %q: expected kind %q, got %q", c.id, c.kind, got.Kind)
		}
	}
}

func TestParseDefinitions_DecisionService(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(complexExprDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	if len(defs.DecisionService) == 0 {
		t.Error("expected at least one decision service")
	}
	ds := defs.DecisionService[0]
	if len(ds.OutputDecisions) == 0 {
		t.Error("expected output decisions")
	}
	if len(ds.EncapsulatedDecisions) == 0 {
		t.Error("expected encapsulated decisions")
	}
	if len(ds.InputDecisions) == 0 {
		t.Error("expected input decisions")
	}
	if len(ds.InputData) == 0 {
		t.Error("expected input data")
	}
}

func TestParseDefinitions_ItemDefinitionWithComponents(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(complexExprDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	if len(defs.ItemDefinitions) == 0 {
		t.Error("expected item definitions")
	}
	// Find tPerson item def with components
	for _, id := range defs.ItemDefinitions {
		if id.Name == "tPerson" {
			if len(id.ItemComponents) == 0 {
				t.Error("expected item components for tPerson")
			}
			return
		}
	}
	t.Error("tPerson item definition not found")
}

func TestParseDefinitions_ContextWithEntries(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(complexExprDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	for _, d := range defs.Decisions {
		if d.ID == "Dec_Context" {
			if d.Logic == nil || d.Logic.Context == nil {
				t.Error("expected context logic")
				return
			}
			if len(d.Logic.Context.Entries) == 0 {
				t.Error("expected context entries")
			}
			return
		}
	}
	t.Error("Dec_Context not found")
}

func TestParseDefinitions_ListWithItems(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(complexExprDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	for _, d := range defs.Decisions {
		if d.ID == "Dec_List" {
			if d.Logic == nil || d.Logic.List == nil {
				t.Error("expected list logic")
				return
			}
			if len(d.Logic.List.Items) == 0 {
				t.Error("expected list items")
			}
			return
		}
	}
	t.Error("Dec_List not found")
}

func TestParseDefinitions_RelationWithRowsAndColumns(t *testing.T) {
	defs, err := dmn.ParseDefinitions([]byte(complexExprDMN))
	if err != nil {
		t.Fatalf("ParseDefinitions: %v", err)
	}
	for _, d := range defs.Decisions {
		if d.ID == "Dec_Relation" {
			if d.Logic == nil || d.Logic.Relation == nil {
				t.Error("expected relation logic")
				return
			}
			if len(d.Logic.Relation.Columns) == 0 {
				t.Error("expected relation columns")
			}
			return
		}
	}
	t.Error("Dec_Relation not found")
}
