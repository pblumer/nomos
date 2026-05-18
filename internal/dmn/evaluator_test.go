package dmn_test

import (
	"testing"

	"github.com/nomos/nomos/internal/dmn"
)

// --- FEEL unary test parser ---

func TestParseUnaryTest_AnyAndEmpty(t *testing.T) {
	for _, expr := range []string{"", "-", "  -  "} {
		ut, err := dmn.ParseUnaryTest(expr)
		if err != nil {
			t.Fatalf("ParseUnaryTest(%q): %v", expr, err)
		}
		if !ut.Matches(dmn.NumberVal{V: 42}) {
			t.Errorf("anyTest should match anything, got false for %q", expr)
		}
	}
}

func TestParseUnaryTest_NumberComparisons(t *testing.T) {
	cases := []struct {
		expr   string
		input  float64
		expect bool
	}{
		{"< 1000", 999, true},
		{"< 1000", 1000, false},
		{"<= 1000", 1000, true},
		{">= 1000", 1000, true},
		{"> 1000", 1001, true},
		{"> 1000", 1000, false},
		{"= 42", 42, true},
		{"= 42", 43, false},
		{"!= 42", 43, true},
		{"!= 42", 42, false},
	}
	for _, c := range cases {
		ut, err := dmn.ParseUnaryTest(c.expr)
		if err != nil {
			t.Fatalf("ParseUnaryTest(%q): %v", c.expr, err)
		}
		got := ut.Matches(dmn.NumberVal{V: c.input})
		if got != c.expect {
			t.Errorf("(%q).Matches(%v) = %v, want %v", c.expr, c.input, got, c.expect)
		}
	}
}

func TestParseUnaryTest_StringLiteral(t *testing.T) {
	ut, _ := dmn.ParseUnaryTest(`"gold"`)
	if !ut.Matches(dmn.StringVal{V: "gold"}) {
		t.Error(`"gold" should match "gold"`)
	}
	if ut.Matches(dmn.StringVal{V: "silver"}) {
		t.Error(`"gold" should not match "silver"`)
	}
}

func TestParseUnaryTest_Range(t *testing.T) {
	cases := []struct {
		expr   string
		input  float64
		expect bool
	}{
		{"[100..500]", 100, true},
		{"[100..500]", 500, true},
		{"[100..500]", 99, false},
		{"[100..500]", 501, false},
		{"(100..500)", 100, false},
		{"(100..500)", 101, true},
		{"[100..500)", 500, false},
		{"[100..500)", 499, true},
		{"(100..500]", 100, false},
		{"(100..500]", 500, true},
	}
	for _, c := range cases {
		ut, err := dmn.ParseUnaryTest(c.expr)
		if err != nil {
			t.Fatalf("ParseUnaryTest(%q): %v", c.expr, err)
		}
		got := ut.Matches(dmn.NumberVal{V: c.input})
		if got != c.expect {
			t.Errorf("(%q).Matches(%v) = %v, want %v", c.expr, c.input, got, c.expect)
		}
	}
}

func TestParseUnaryTest_Alternatives(t *testing.T) {
	ut, err := dmn.ParseUnaryTest(`"gold", "platinum"`)
	if err != nil {
		t.Fatal(err)
	}
	if !ut.Matches(dmn.StringVal{V: "gold"}) {
		t.Error("should match gold")
	}
	if !ut.Matches(dmn.StringVal{V: "platinum"}) {
		t.Error("should match platinum")
	}
	if ut.Matches(dmn.StringVal{V: "silver"}) {
		t.Error("should not match silver")
	}
}

func TestParseUnaryTest_Not(t *testing.T) {
	ut, err := dmn.ParseUnaryTest(`not("draft", "deprecated")`)
	if err != nil {
		t.Fatal(err)
	}
	if ut.Matches(dmn.StringVal{V: "draft"}) {
		t.Error("not() should reject draft")
	}
	if !ut.Matches(dmn.StringVal{V: "active"}) {
		t.Error("not() should accept active")
	}
}

func TestParseUnaryTest_Null(t *testing.T) {
	ut, err := dmn.ParseUnaryTest("null")
	if err != nil {
		t.Fatal(err)
	}
	if !ut.Matches(dmn.NullVal{}) {
		t.Error("null test should match NullVal")
	}
	if ut.Matches(dmn.StringVal{V: ""}) {
		t.Error("null test should not match empty string")
	}
}

// --- DMN XML parser ---

const sampleDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/"
             id="def1" name="Test">
  <decision id="DEC-001" name="Approve Order">
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="in1" label="Order Amount">
        <inputExpression typeRef="number"><text>orderAmount</text></inputExpression>
      </input>
      <input id="in2" label="Customer Tier">
        <inputExpression typeRef="string"><text>customerTier</text></inputExpression>
      </input>
      <output id="out1" name="approved" typeRef="boolean"/>
      <output id="out2" name="approvalReason" typeRef="string"/>
      <rule id="r_vip">
        <inputEntry><text></text></inputEntry>
        <inputEntry><text>"vip"</text></inputEntry>
        <outputEntry><text>true</text></outputEntry>
        <outputEntry><text>"VIP: auto approved"</text></outputEntry>
      </rule>
      <rule id="r_small">
        <inputEntry><text>&lt; 1000</text></inputEntry>
        <inputEntry><text></text></inputEntry>
        <outputEntry><text>true</text></outputEntry>
        <outputEntry><text>"Small order: auto approved"</text></outputEntry>
      </rule>
      <rule id="r_large">
        <inputEntry><text>&gt;= 1000</text></inputEntry>
        <inputEntry><text></text></inputEntry>
        <outputEntry><text>false</text></outputEntry>
        <outputEntry><text>"Large order: manual review"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

func parseSample(t *testing.T) *dmn.Table {
	t.Helper()
	table, err := dmn.ParseDMN([]byte(sampleDMN))
	if err != nil {
		t.Fatalf("ParseDMN: %v", err)
	}
	return table
}

func TestParseDMN_Structure(t *testing.T) {
	table := parseSample(t)
	if table.DecisionID != "DEC-001" {
		t.Errorf("DecisionID = %q, want DEC-001", table.DecisionID)
	}
	if table.HitPolicy != "FIRST" {
		t.Errorf("HitPolicy = %q, want FIRST", table.HitPolicy)
	}
	if len(table.Inputs) != 2 {
		t.Errorf("len(Inputs) = %d, want 2", len(table.Inputs))
	}
	if len(table.Outputs) != 2 {
		t.Errorf("len(Outputs) = %d, want 2", len(table.Outputs))
	}
	if len(table.Rules) != 3 {
		t.Errorf("len(Rules) = %d, want 3", len(table.Rules))
	}
	if table.Inputs[0].Expression != "orderAmount" {
		t.Errorf("Input[0].Expression = %q, want orderAmount", table.Inputs[0].Expression)
	}
}

// --- Evaluator ---

func evaluate(t *testing.T, inputs map[string]any) *dmn.Result {
	t.Helper()
	table := parseSample(t)
	res, err := dmn.Evaluate(table, inputs)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return res
}

func TestEvaluate_VIPAlwaysApproved(t *testing.T) {
	res := evaluate(t, map[string]any{"orderAmount": 5000.0, "customerTier": "vip"})
	if res.Outputs["approved"] != true {
		t.Errorf("approved = %v, want true", res.Outputs["approved"])
	}
	if len(res.MatchedRules) == 0 || res.MatchedRules[0] != "r_vip" {
		t.Errorf("matched rule = %v, want [r_vip]", res.MatchedRules)
	}
}

func TestEvaluate_SmallOrderApproved(t *testing.T) {
	res := evaluate(t, map[string]any{"orderAmount": 500.0, "customerTier": "standard"})
	if res.Outputs["approved"] != true {
		t.Errorf("approved = %v, want true", res.Outputs["approved"])
	}
	if len(res.MatchedRules) == 0 || res.MatchedRules[0] != "r_small" {
		t.Errorf("matched rule = %v, want [r_small]", res.MatchedRules)
	}
}

func TestEvaluate_LargeOrderManual(t *testing.T) {
	res := evaluate(t, map[string]any{"orderAmount": 1500.0, "customerTier": "standard"})
	if res.Outputs["approved"] != false {
		t.Errorf("approved = %v, want false", res.Outputs["approved"])
	}
	if res.Outputs["approvalReason"] != "Large order: manual review" {
		t.Errorf("approvalReason = %v", res.Outputs["approvalReason"])
	}
}

func TestEvaluate_NoMatch(t *testing.T) {
	// Edge: amount exactly 1000 → r_large matches (>= 1000)
	res := evaluate(t, map[string]any{"orderAmount": 1000.0, "customerTier": "standard"})
	if res.Outputs["approved"] != false {
		t.Errorf("at exactly 1000 should be manual review, got approved=%v", res.Outputs["approved"])
	}
}

// --- COLLECT hit policy ---

const collectDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="Discount">
    <decisionTable id="dt1" hitPolicy="COLLECT" aggregation="SUM">
      <input id="i1" label="Category">
        <inputExpression typeRef="string"><text>category</text></inputExpression>
      </input>
      <output id="o1" name="discount" typeRef="number"/>
      <rule id="r1">
        <inputEntry><text>"premium"</text></inputEntry>
        <outputEntry><text>5</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry><text>"loyalty"</text></inputEntry>
        <outputEntry><text>3</text></outputEntry>
      </rule>
      <rule id="r3">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>1</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`

func TestEvaluate_CollectSum(t *testing.T) {
	table, err := dmn.ParseDMN([]byte(collectDMN))
	if err != nil {
		t.Fatal(err)
	}
	// "premium" matches r1 (5) and r3 (1 = any), sum = 6
	res, err := dmn.Evaluate(table, map[string]any{"category": "premium"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["discount"] != 6.0 {
		t.Errorf("discount = %v, want 6", res.Outputs["discount"])
	}
}

// --- Hit policy shortcuts ---

func TestNormaliseHitPolicy_Shorthand(t *testing.T) {
	const uniqueDMN = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="Test">
    <decisionTable id="dt1" hitPolicy="U">
      <input id="i1"><inputExpression typeRef="number"><text>x</text></inputExpression></input>
      <output id="o1" name="y" typeRef="number"/>
      <rule id="r1">
        <inputEntry><text>&lt; 10</text></inputEntry>
        <outputEntry><text>1</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(uniqueDMN))
	if err != nil {
		t.Fatal(err)
	}
	if table.HitPolicy != "UNIQUE" {
		t.Errorf("HitPolicy = %q, want UNIQUE", table.HitPolicy)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": 5.0})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["y"] != 1.0 {
		t.Errorf("y = %v, want 1", res.Outputs["y"])
	}
}

// --- ToValue type coercion ---

func TestToValue_NumberFromJSON(t *testing.T) {
	v := dmn.ToValue(float64(42), "number")
	if nv, ok := v.(dmn.NumberVal); !ok || nv.V != 42 {
		t.Errorf("ToValue(42, number) = %v", v)
	}
}

func TestToValue_BoolFromString(t *testing.T) {
	v := dmn.ToValue("true", "boolean")
	if bv, ok := v.(dmn.BoolVal); !ok || !bv.V {
		t.Errorf("ToValue(true, boolean) = %v", v)
	}
}

func TestToValue_DateAutoDetect(t *testing.T) {
	v := dmn.ToValue("2024-03-15", "")
	if _, ok := v.(dmn.DateVal); !ok {
		t.Errorf("ToValue(date string) should return DateVal, got %T", v)
	}
}
