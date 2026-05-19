package dmn

import (
	"strings"
	"testing"
)

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"Customer Status":        "customer_status",
		"Value":                  "value",
		"value":                  "value",
		"  Müller-Lüdenscheidt ": "mueller_luedenscheidt",
		"Größe in kg":            "groesse_in_kg",
		"foo__bar":               "foo_bar",
		"123":                    "123",
		"_already_snake_":        "already_snake",
		"":                       "",
		"   ":                    "",
		"!!!":                    "",
	}
	for in, want := range cases {
		if got := SnakeCase(in); got != want {
			t.Errorf("SnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}

// Mirrors the screenshot scenario: a DMN whose <variable> attributes already
// match the descriptive pill labels ("Customer Status"). The normalizer must
// derive snake_case variable names and rewrite every decision-table input
// column that was still pointing at the old descriptive form.
func TestNormalizeInputIdentifiers_SyncsPillsToTable(t *testing.T) {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="Defs" name="Kreditpruefung">
  <inputData id="ID_value" name="Value">
    <variable id="V_value" name="Value"/>
  </inputData>
  <inputData id="ID_status" name="Customer Status">
    <variable id="V_status" name="Customer Status"/>
  </inputData>
  <decision id="Dec_1" name="Kreditpruefung">
    <variable id="Var_Dec1" name="kredit_stufe" typeRef="string"/>
    <decisionTable id="DT_1" hitPolicy="FIRST">
      <input id="In_value" label="Value"><inputExpression id="IE_value" typeRef="number"><text>Value</text></inputExpression></input>
      <input id="In_status" label="Customer Status"><inputExpression id="IE_status" typeRef="string"><text>Customer Status</text></inputExpression></input>
      <output id="Out_1" name="kredit_stufe" typeRef="string"/>
      <rule id="R_1">
        <inputEntry id="IE_R1_1"><text>&lt; 100</text></inputEntry>
        <inputEntry id="IE_R1_2"><text>-</text></inputEntry>
        <outputEntry id="OE_R1"><text>"kleiner Betrag"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`)
	out, err := NormalizeInputIdentifiers(xml)
	if err != nil {
		t.Fatalf("NormalizeInputIdentifiers: %v", err)
	}
	s := string(out)
	// Pill labels stay readable.
	if !strings.Contains(s, `<inputData id="ID_status" name="Customer Status">`) {
		t.Errorf("pill label rewritten unexpectedly: %s", s)
	}
	// Variable names are snake_cased.
	for _, want := range []string{
		`<variable id="V_value" name="value"`,
		`<variable id="V_status" name="customer_status"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in: %s", want, s)
		}
	}
	// Decision-table input columns reference the new variable names.
	for _, want := range []string{
		`<text>value</text>`,
		`<text>customer_status</text>`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in: %s", want, s)
		}
	}
	// And no longer reference the descriptive forms.
	for _, dontWant := range []string{
		`<text>Value</text>`,
		`<text>Customer Status</text>`,
	} {
		if strings.Contains(s, dontWant) {
			t.Errorf("unexpected %q lingering in: %s", dontWant, s)
		}
	}
}

// When the DMN is already canonical the rewrite must be a no-op (byte
// identical) so we don't churn rule_hash / snapshot version on every save.
func TestNormalizeInputIdentifiers_NoOpWhenCanonical(t *testing.T) {
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="Defs" name="X">
  <inputData id="ID_v" name="value">
    <variable id="V_v" name="value"/>
  </inputData>
  <decision id="D" name="D">
    <variable id="Var_D" name="r" typeRef="string"/>
    <decisionTable id="DT" hitPolicy="FIRST">
      <input id="In"><inputExpression id="IE" typeRef="number"><text>value</text></inputExpression></input>
      <output id="Out" name="r" typeRef="string"/>
      <rule id="R"><inputEntry id="IE_R"><text>-</text></inputEntry><outputEntry id="OE_R"><text>"x"</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`)
	out, err := NormalizeInputIdentifiers(xml)
	if err != nil {
		t.Fatalf("NormalizeInputIdentifiers: %v", err)
	}
	if string(out) != string(xml) {
		t.Errorf("expected no-op on canonical input.\nin:  %s\nout: %s", xml, out)
	}
}
