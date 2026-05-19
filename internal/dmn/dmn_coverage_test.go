package dmn_test

import (
	"testing"
	"time"

	"github.com/nomos/nomos/internal/dmn"
)

// ---------------------------------------------------------------------------
// Value interface — feelType / GoValue / String on concrete types
// ---------------------------------------------------------------------------

func TestValueTypes(t *testing.T) {
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	dt := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

	cases := []struct {
		val       dmn.Value
		feelType  string
		wantStr   string
		wantGoVal any
	}{
		{dmn.NumberVal{V: 42}, "number", "42", float64(42)},
		{dmn.NumberVal{V: 3.14}, "number", "3.14", float64(3.14)},
		{dmn.StringVal{V: "hello"}, "string", "hello", "hello"},
		{dmn.BoolVal{V: true}, "boolean", "true", true},
		{dmn.BoolVal{V: false}, "boolean", "false", false},
		{dmn.NullVal{}, "null", "null", nil},
		{dmn.DateVal{V: now}, "date", "2024-03-15", "2024-03-15"},
		{dmn.DateTimeVal{V: dt}, "date time", dt.Format(time.RFC3339), dt.Format(time.RFC3339)},
	}
	for _, c := range cases {
		t.Run(c.feelType, func(t *testing.T) {
			if got := c.val.String(); got != c.wantStr {
				t.Errorf("String() = %q, want %q", got, c.wantStr)
			}
			if got := c.val.GoValue(); got != c.wantGoVal {
				t.Errorf("GoValue() = %v, want %v", got, c.wantGoVal)
			}
		})
	}
}

func TestListValStringAndGoValue(t *testing.T) {
	lv := dmn.ListVal{Items: []dmn.Value{dmn.NumberVal{V: 1}, dmn.StringVal{V: "a"}}}
	if got := lv.String(); got != "[1, a]" {
		t.Errorf("ListVal.String() = %q", got)
	}
	gv, ok := lv.GoValue().([]any)
	if !ok || len(gv) != 2 {
		t.Errorf("ListVal.GoValue() = %v", lv.GoValue())
	}
}

// ---------------------------------------------------------------------------
// ToValue additional coercions
// ---------------------------------------------------------------------------

func TestToValue_Float32(t *testing.T) {
	v := dmn.ToValue(float32(1.5), "number")
	if nv, ok := v.(dmn.NumberVal); !ok || nv.V != float64(float32(1.5)) {
		t.Errorf("ToValue float32: %v", v)
	}
}

func TestToValue_Int(t *testing.T) {
	v := dmn.ToValue(int(7), "")
	if nv, ok := v.(dmn.NumberVal); !ok || nv.V != 7 {
		t.Errorf("ToValue int: %v", v)
	}
}

func TestToValue_Int64(t *testing.T) {
	v := dmn.ToValue(int64(99), "")
	if nv, ok := v.(dmn.NumberVal); !ok || nv.V != 99 {
		t.Errorf("ToValue int64: %v", v)
	}
}

func TestToValue_DateHint(t *testing.T) {
	v := dmn.ToValue("2024-06-01", "date")
	if _, ok := v.(dmn.DateVal); !ok {
		t.Errorf("ToValue date hint: %T", v)
	}
}

func TestToValue_DateTimeHint(t *testing.T) {
	v := dmn.ToValue("2024-06-01T10:00:00Z", "date and time")
	if _, ok := v.(dmn.DateTimeVal); !ok {
		t.Errorf("ToValue date and time hint: %T", v)
	}
}

func TestToValue_DateTimeHintFallback(t *testing.T) {
	v := dmn.ToValue("2024-06-01T10:00:00", "dateTime")
	if _, ok := v.(dmn.DateTimeVal); !ok {
		t.Errorf("ToValue dateTime hint: %T", v)
	}
}

func TestToValue_NumberFromString(t *testing.T) {
	v := dmn.ToValue("3.14", "number")
	if nv, ok := v.(dmn.NumberVal); !ok || nv.V != 3.14 {
		t.Errorf("ToValue number from string: %v", v)
	}
}

func TestToValue_IntegerFromString(t *testing.T) {
	v := dmn.ToValue("42", "integer")
	if nv, ok := v.(dmn.NumberVal); !ok || nv.V != 42 {
		t.Errorf("ToValue integer from string: %v", v)
	}
}

func TestToValue_BoolFalseFromString(t *testing.T) {
	v := dmn.ToValue("false", "boolean")
	if bv, ok := v.(dmn.BoolVal); !ok || bv.V {
		t.Errorf("ToValue false from string: %v", v)
	}
}

func TestToValue_Nil(t *testing.T) {
	v := dmn.ToValue(nil, "")
	if _, ok := v.(dmn.NullVal); !ok {
		t.Errorf("ToValue nil: %T", v)
	}
}

func TestToValue_UnknownType(t *testing.T) {
	v := dmn.ToValue(struct{ X int }{X: 1}, "")
	if _, ok := v.(dmn.StringVal); !ok {
		t.Errorf("ToValue unknown: %T", v)
	}
}

// ---------------------------------------------------------------------------
// UnaryTest String() methods
// ---------------------------------------------------------------------------

func TestUnaryTestString(t *testing.T) {
	any, _ := dmn.ParseUnaryTest("-")
	if any.String() != "-" {
		t.Errorf("anyTest.String() = %q", any.String())
	}
	lit, _ := dmn.ParseUnaryTest(`"gold"`)
	if lit.String() == "" {
		t.Error("literalTest.String() should not be empty")
	}
	cmp, _ := dmn.ParseUnaryTest("< 100")
	if cmp.String() == "" {
		t.Error("compareTest.String() should not be empty")
	}
	rng, _ := dmn.ParseUnaryTest("[1..10]")
	if rng.String() == "" {
		t.Error("rangeTest.String() should not be empty")
	}
	lst, _ := dmn.ParseUnaryTest(`"gold", "silver"`)
	if lst.String() == "" {
		t.Error("listTest.String() should not be empty")
	}
	not, _ := dmn.ParseUnaryTest(`not("gold")`)
	if not.String() == "" {
		t.Error("notTest.String() should not be empty")
	}
}

func TestNullTest(t *testing.T) {
	null, _ := dmn.ParseUnaryTest("null")
	if !null.Matches(dmn.NullVal{}) {
		t.Error("nullTest should match NullVal")
	}
	if null.Matches(dmn.StringVal{V: "x"}) {
		t.Error("nullTest should not match non-null")
	}
	if null.String() != "null" {
		t.Errorf("nullTest.String() = %q", null.String())
	}
}

// ---------------------------------------------------------------------------
// Evaluate — hit policies
// ---------------------------------------------------------------------------

func TestEvaluate_FIRST_NoMatch(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1"><inputExpression typeRef="number"><text>score</text></inputExpression></input>
      <output id="o1" name="result" typeRef="string"/>
      <rule id="r1">
        <inputEntry><text>&gt;= 90</text></inputEntry>
        <outputEntry><text>"high"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"score": float64(50)})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.MatchedRules) != 0 {
		t.Errorf("expected no matched rules, got %v", res.MatchedRules)
	}
}

func TestEvaluate_ANY_AllMatch(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="ANY">
      <input id="i1"><inputExpression typeRef="number"><text>score</text></inputExpression></input>
      <output id="o1" name="result" typeRef="string"/>
      <rule id="r1">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>"pass"</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>"pass"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"score": float64(80)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["result"] != "pass" {
		t.Errorf("result = %v", res.Outputs["result"])
	}
}

func TestEvaluate_ANY_NoMatch(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="ANY">
      <input id="i1"><inputExpression typeRef="number"><text>score</text></inputExpression></input>
      <output id="o1" name="result" typeRef="string"/>
      <rule id="r1">
        <inputEntry><text>&gt;= 90</text></inputEntry>
        <outputEntry><text>"high"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"score": float64(50)})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.MatchedRules) != 0 {
		t.Errorf("expected no match")
	}
}

func TestEvaluate_ANY_ConflictError(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="ANY">
      <input id="i1"><inputExpression typeRef="number"><text>score</text></inputExpression></input>
      <output id="o1" name="result" typeRef="string"/>
      <rule id="r1">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>"alpha"</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>"beta"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	_, err = dmn.Evaluate(table, map[string]any{"score": float64(80)})
	if err == nil {
		t.Error("expected ANY conflict error")
	}
}

func TestEvaluate_COLLECT_SUM(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="COLLECT" aggregation="SUM">
      <input id="i1"><inputExpression typeRef="string"><text>cat</text></inputExpression></input>
      <output id="o1" name="score" typeRef="number"/>
      <rule id="r1">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>10</text></outputEntry>
      </rule>
      <rule id="r2">
        <inputEntry><text>-</text></inputEntry>
        <outputEntry><text>5</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"cat": "A"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["score"] != float64(15) {
		t.Errorf("SUM = %v, want 15", res.Outputs["score"])
	}
}

func TestEvaluate_COLLECT_MIN(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="COLLECT" aggregation="MIN">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="val" typeRef="number"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>3</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>7</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["val"] != float64(3) {
		t.Errorf("MIN = %v, want 3", res.Outputs["val"])
	}
}

func TestEvaluate_COLLECT_MAX(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="COLLECT" aggregation="MAX">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="val" typeRef="number"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>3</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>7</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["val"] != float64(7) {
		t.Errorf("MAX = %v, want 7", res.Outputs["val"])
	}
}

func TestEvaluate_COLLECT_COUNT(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="COLLECT" aggregation="COUNT">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="cnt" typeRef="number"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>1</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>2</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["cnt"] != float64(2) {
		t.Errorf("COUNT = %v, want 2", res.Outputs["cnt"])
	}
}

func TestEvaluate_COLLECT_NoAggregation(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="COLLECT">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="vals" typeRef="string"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>"a"</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>"b"</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	lst, ok := res.Outputs["vals"].([]any)
	if !ok || len(lst) != 2 {
		t.Errorf("COLLECT no agg = %v", res.Outputs["vals"])
	}
}

func TestEvaluate_RULE_ORDER(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="RULE ORDER">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="vals" typeRef="string"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>"a"</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>"b"</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.MatchedRules) != 2 {
		t.Errorf("RULE ORDER matched %d rules, want 2", len(res.MatchedRules))
	}
}

func TestEvaluate_DefaultHitPolicy(t *testing.T) {
	// Unknown hit policy falls through to FIRST behaviour
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1"><inputExpression typeRef="number"><text>score</text></inputExpression></input>
      <output id="o1" name="result" typeRef="string"/>
      <rule id="r1">
        <inputEntry><text>&gt;= 90</text></inputEntry>
        <outputEntry><text>"high"</text></outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	table.HitPolicy = "UNKNOWN_POLICY"
	res, err := dmn.Evaluate(table, map[string]any{"score": float64(50)})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.MatchedRules) != 0 {
		t.Errorf("expected no match for unknown policy+no match, got %v", res.MatchedRules)
	}
}

func TestEvaluate_UNIQUE_Violation(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="UNIQUE">
      <input id="i1"><inputExpression typeRef="number"><text>score</text></inputExpression></input>
      <output id="o1" name="result" typeRef="string"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>"a"</text></outputEntry></rule>
      <rule id="r2"><inputEntry><text>-</text></inputEntry><outputEntry><text>"b"</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	_, err = dmn.Evaluate(table, map[string]any{"score": float64(50)})
	if err == nil {
		t.Error("expected UNIQUE violation error")
	}
}

func TestEvaluate_CoerceOutput_Bool(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="flag" typeRef="boolean"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>true</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["flag"] != true {
		t.Errorf("coerce bool output = %v", res.Outputs["flag"])
	}
}

func TestEvaluate_CoerceOutput_Number(t *testing.T) {
	const src = `<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" id="d" name="d">
  <decision id="D1" name="test">
    <decisionTable id="dt1" hitPolicy="FIRST">
      <input id="i1"><inputExpression typeRef="string"><text>x</text></inputExpression></input>
      <output id="o1" name="n" typeRef="number"/>
      <rule id="r1"><inputEntry><text>-</text></inputEntry><outputEntry><text>42</text></outputEntry></rule>
    </decisionTable>
  </decision>
</definitions>`
	table, err := dmn.ParseDMN([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	res, err := dmn.Evaluate(table, map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Outputs["n"] != float64(42) {
		t.Errorf("coerce number output = %v", res.Outputs["n"])
	}
}
