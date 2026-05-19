package feel

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Value type methods: feelType / GoValue
// ---------------------------------------------------------------------------

func TestValueTypeMethods(t *testing.T) {
	if Number(1).feelType() != "number" {
		t.Error("Number.feelType()")
	}
	if String("x").feelType() != "string" {
		t.Error("String.feelType()")
	}
	if Bool(true).feelType() != "boolean" {
		t.Error("Bool.feelType()")
	}
	if (Null{}).feelType() != "null" {
		t.Error("Null.feelType()")
	}
	if Number(3).GoValue() != float64(3) {
		t.Error("Number.GoValue()")
	}
	if String("y").GoValue() != "y" {
		t.Error("String.GoValue()")
	}
	if Bool(false).GoValue() != false {
		t.Error("Bool.GoValue()")
	}
	if (Null{}).GoValue() != nil {
		t.Error("Null.GoValue()")
	}
	l := List{Number(1), Number(2)}
	if l.feelType() != "list" {
		t.Error("List.feelType()")
	}
	lv := l.GoValue().([]any)
	if len(lv) != 2 {
		t.Error("List.GoValue()")
	}
	c := Context{"a": Number(1)}
	if c.feelType() != "context" {
		t.Error("Context.feelType()")
	}
	cv := c.GoValue().(map[string]any)
	if cv["a"] != float64(1) {
		t.Error("Context.GoValue()")
	}
}

func TestDateDateTimeTypeMethods(t *testing.T) {
	// Test Date feelType / GoValue
	d, err := EvalString(`date("2024-03-15")`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.feelType() != "date" {
		t.Errorf("Date.feelType() = %q", d.feelType())
	}
	if d.GoValue() == nil {
		t.Error("Date.GoValue() should not be nil")
	}
	// DateTime via FromGo
	dt := FromGo("2024-03-15T10:00:00Z")
	if dt.feelType() != "date and time" {
		t.Errorf("DateTime.feelType() = %q", dt.feelType())
	}
	if dt.GoValue() == nil {
		t.Error("DateTime.GoValue() should not be nil")
	}
}

func TestBuiltinRefTypeMethods(t *testing.T) {
	// Calling a builtin function name as a value yields a builtinRef
	// which has feelType() = "function" and GoValue() = name string.
	// We can observe this by calling a nonexistent name: it should error.
	// Instead, test via reflection on the function lookup.
	fn := lookupBuiltin("string length")
	if fn == nil {
		t.Fatal("lookupBuiltin string length")
	}
	// Also test the underscore alias.
	fn2 := lookupBuiltin("string_length")
	if fn2 == nil {
		t.Fatal("lookupBuiltin string_length (underscore alias)")
	}
	// Unknown builtin returns nil.
	if lookupBuiltin("not_a_real_builtin") != nil {
		t.Error("expected nil for unknown builtin")
	}
}

// ---------------------------------------------------------------------------
// Source() method on Expression
// ---------------------------------------------------------------------------

func TestExpressionSource(t *testing.T) {
	expr, err := Parse("1 + 2")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Source() != "1 + 2" {
		t.Errorf("Source() = %q", expr.Source())
	}
}

// ---------------------------------------------------------------------------
// FromGo conversions
// ---------------------------------------------------------------------------

func TestFromGoList(t *testing.T) {
	v := FromGo([]any{float64(1), "hello", true})
	l, ok := v.(List)
	if !ok || len(l) != 3 {
		t.Errorf("FromGo([]any{...}) = %T", v)
	}
}

func TestFromGoContext(t *testing.T) {
	v := FromGo(map[string]any{"x": float64(1)})
	c, ok := v.(Context)
	if !ok || len(c) != 1 {
		t.Errorf("FromGo(map) = %T", v)
	}
}

func TestFromGoNil(t *testing.T) {
	v := FromGo(nil)
	if _, ok := v.(Null); !ok {
		t.Errorf("FromGo(nil) = %T", v)
	}
}

func TestFromGoFloat32(t *testing.T) {
	v := FromGo(float32(1.5))
	if _, ok := v.(Number); !ok {
		t.Errorf("FromGo(float32) = %T", v)
	}
}

func TestFromGoInt(t *testing.T) {
	v := FromGo(int(7))
	if _, ok := v.(Number); !ok {
		t.Errorf("FromGo(int) = %T", v)
	}
}

func TestFromGoInt64(t *testing.T) {
	v := FromGo(int64(99))
	if _, ok := v.(Number); !ok {
		t.Errorf("FromGo(int64) = %T", v)
	}
}

func TestFromGoNumberString(t *testing.T) {
	v := FromGo("3.14")
	if _, ok := v.(Number); !ok {
		t.Errorf("FromGo(number string) = %T", v)
	}
}

func TestFromGoUnknown(t *testing.T) {
	v := FromGo(struct{}{})
	if _, ok := v.(Null); !ok {
		t.Errorf("FromGo(unknown) = %T", v)
	}
}

// ---------------------------------------------------------------------------
// EvalString — uncovered expression paths
// ---------------------------------------------------------------------------

func TestEvalIfElse(t *testing.T) {
	v, err := EvalString("if 1 < 2 then \"yes\" else \"no\"", nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(v.(String)) != "yes" {
		t.Errorf("if-then-else = %v", v)
	}
	v2, _ := EvalString("if false then 1 else 2", nil)
	if v2.(Number) != 2 {
		t.Errorf("if false branch = %v", v2)
	}
}

func TestEvalListExpr(t *testing.T) {
	v, err := EvalString("[1, 2, 3]", nil)
	if err != nil {
		t.Fatal(err)
	}
	l, ok := v.(List)
	if !ok || len(l) != 3 {
		t.Errorf("list expr = %v", v)
	}
}

func TestEvalContextExpr(t *testing.T) {
	v, err := EvalString(`{name: "Alice", age: 30}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := v.(Context)
	if !ok || len(c) != 2 {
		t.Errorf("context expr = %v", v)
	}
}

func TestEvalPathExpr(t *testing.T) {
	env := EnvFromGo(map[string]any{"person": map[string]any{"name": "Bob"}})
	v, err := EvalString("person.name", env)
	if err != nil {
		t.Fatal(err)
	}
	if string(v.(String)) != "Bob" {
		t.Errorf("path expr = %v", v)
	}
}

func TestEvalDatePathNavigation(t *testing.T) {
	v, err := EvalString(`date("2024-06-15").year`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.(Number) != 2024 {
		t.Errorf("date.year = %v", v)
	}
	v2, _ := EvalString(`date("2024-06-15").month`, nil)
	if v2.(Number) != 6 {
		t.Errorf("date.month = %v", v2)
	}
	v3, _ := EvalString(`date("2024-06-15").day`, nil)
	if v3.(Number) != 15 {
		t.Errorf("date.day = %v", v3)
	}
}

func TestEvalDateTimePathNavigation(t *testing.T) {
	// DateTime via env variable (FromGo converts RFC3339 strings to DateTime)
	env := EnvFromGo(map[string]any{"ts": "2024-06-15T10:30:45Z"})
	v, err := EvalString("ts.hour", env)
	if err != nil {
		t.Fatal(err)
	}
	if v.(Number) != 10 {
		t.Errorf("datetime.hour = %v", v)
	}
	v2, _ := EvalString("ts.minute", env)
	if v2.(Number) != 30 {
		t.Errorf("datetime.minute = %v", v2)
	}
	v3, _ := EvalString("ts.second", env)
	if v3.(Number) != 45 {
		t.Errorf("datetime.second = %v", v3)
	}
}

func TestEvalStringConcatenation(t *testing.T) {
	v, err := EvalString(`"hello" + " " + "world"`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(v.(String)) != "hello world" {
		t.Errorf("string concat = %v", v)
	}
}

func TestEvalListConcatenation(t *testing.T) {
	v, err := EvalString("[1,2] + [3]", nil)
	if err != nil {
		t.Fatal(err)
	}
	l, ok := v.(List)
	if !ok || len(l) != 3 {
		t.Errorf("list concat = %v", v)
	}
}

func TestEvalStringComparison(t *testing.T) {
	v, _ := EvalString(`"apple" < "banana"`, nil)
	if !bool(v.(Bool)) {
		t.Error("string < comparison")
	}
}

func TestEvalDateComparison(t *testing.T) {
	v, err := EvalString(`date("2024-01-01") < date("2024-12-31")`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bool(v.(Bool)) {
		t.Error("date < comparison")
	}
}

func TestEvalDateTimeComparison(t *testing.T) {
	env := EnvFromGo(map[string]any{"a": "2024-01-01T00:00:00Z", "b": "2024-12-31T00:00:00Z"})
	v, err := EvalString("a < b", env)
	if err != nil {
		t.Fatal(err)
	}
	if !bool(v.(Bool)) {
		t.Error("datetime < comparison")
	}
}

func TestEvalEqualValues(t *testing.T) {
	// null == null
	v, _ := EvalString("null = null", nil)
	if !bool(v.(Bool)) {
		t.Error("null = null")
	}
	// null != string
	v2, _ := EvalString(`null = "x"`, nil)
	if bool(v2.(Bool)) {
		t.Error("null = string should be false")
	}
	// list equality
	v3, _ := EvalString("[1,2] = [1,2]", nil)
	if !bool(v3.(Bool)) {
		t.Error("[1,2] = [1,2]")
	}
	// context equality
	v4, _ := EvalString(`{a: 1} = {a: 1}`, nil)
	if !bool(v4.(Bool)) {
		t.Error("{a:1} = {a:1}")
	}
}

func TestEvalNotEqual(t *testing.T) {
	v, _ := EvalString("1 != 2", nil)
	if !bool(v.(Bool)) {
		t.Error("1 != 2")
	}
}

func TestEvalAndOr(t *testing.T) {
	// Short-circuit and
	v, _ := EvalString("false and true", nil)
	if bool(v.(Bool)) {
		t.Error("false and true should be false")
	}
	v2, _ := EvalString("true and true", nil)
	if !bool(v2.(Bool)) {
		t.Error("true and true should be true")
	}
	// Short-circuit or
	v3, _ := EvalString("true or false", nil)
	if !bool(v3.(Bool)) {
		t.Error("true or false")
	}
	v4, _ := EvalString("false or true", nil)
	if !bool(v4.(Bool)) {
		t.Error("false or true")
	}
}

func TestEvalNot(t *testing.T) {
	v, _ := EvalString("not true", nil)
	if bool(v.(Bool)) {
		t.Error("not true should be false")
	}
	v2, _ := EvalString("not false", nil)
	if !bool(v2.(Bool)) {
		t.Error("not false should be true")
	}
}

func TestEvalListIndex(t *testing.T) {
	v, err := EvalString("[10, 20, 30][2]", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.(Number) != 20 {
		t.Errorf("list index [2] = %v", v)
	}
}

func TestEvalBuiltinStringFunctions(t *testing.T) {
	cases := []struct {
		expr string
		want any
	}{
		{`string_length("hello")`, 5.0},
		{`upper_case("hello")`, "HELLO"},
		{`lower_case("WORLD")`, "world"},
		{`contains("hello world", "world")`, true},
		{`starts_with("hello", "he")`, true},
		{`ends_with("hello", "lo")`, true},
		{`substring("hello", 2, 3)`, "ell"},
	}
	for _, c := range cases {
		v, err := EvalString(c.expr, nil)
		if err != nil {
			t.Errorf("EvalString(%q): %v", c.expr, err)
			continue
		}
		switch want := c.want.(type) {
		case float64:
			if float64(v.(Number)) != want {
				t.Errorf("%s = %v, want %v", c.expr, v, want)
			}
		case string:
			if string(v.(String)) != want {
				t.Errorf("%s = %v, want %v", c.expr, v, want)
			}
		case bool:
			if bool(v.(Bool)) != want {
				t.Errorf("%s = %v, want %v", c.expr, v, want)
			}
		}
	}
}

func TestEvalBuiltinMathFunctions(t *testing.T) {
	cases := []struct {
		expr string
		want float64
	}{
		{"abs(-5)", 5},
		{"ceiling(1.2)", 2},
		{"floor(1.9)", 1},
		{"modulo(10, 3)", 1},
	}
	for _, c := range cases {
		v, err := EvalString(c.expr, nil)
		if err != nil {
			t.Errorf("EvalString(%q): %v", c.expr, err)
			continue
		}
		if float64(v.(Number)) != c.want {
			t.Errorf("%s = %v, want %v", c.expr, v, c.want)
		}
	}
}

func TestEvalBuiltinListFunctions(t *testing.T) {
	cases := []struct {
		expr string
		want any
	}{
		{"count([1,2,3])", 3.0},
		{"sum([1,2,3])", 6.0},
		{"min([3,1,2])", 1.0},
		{"max([3,1,2])", 3.0},
		{"mean([1,2,3])", 2.0},
		{`list_contains([1,2,3], 2)`, true},
		{"all([true, true])", true},
		{"all([true, false])", false},
		{"any([false, true])", true},
		{"any([false, false])", false},
	}
	for _, c := range cases {
		v, err := EvalString(c.expr, nil)
		if err != nil {
			t.Errorf("EvalString(%q): %v", c.expr, err)
			continue
		}
		switch want := c.want.(type) {
		case float64:
			if float64(v.(Number)) != want {
				t.Errorf("%s = %v, want %v", c.expr, v, want)
			}
		case bool:
			if bool(v.(Bool)) != want {
				t.Errorf("%s = %v, want %v", c.expr, v, want)
			}
		}
	}
}

func TestEvalToday(t *testing.T) {
	v, err := EvalString("today()", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.feelType() != "date" {
		t.Errorf("today() type = %q", v.feelType())
	}
}

func TestEvalNow(t *testing.T) {
	v, err := EvalString("now()", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.feelType() != "date and time" {
		t.Errorf("now() type = %q", v.feelType())
	}
}

func TestEvalDateBuiltin(t *testing.T) {
	v, err := EvalString(`date("2024-03-15")`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.feelType() != "date" {
		t.Errorf("date() type = %q", v.feelType())
	}
	// 3-arg form: date(year, month, day)
	v2, err := EvalString("date(2024, 3, 15)", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v2.feelType() != "date" {
		t.Errorf("date(y,m,d) type = %q", v2.feelType())
	}
}

// ---------------------------------------------------------------------------
// Error paths
// ---------------------------------------------------------------------------

func TestArgWrongType(t *testing.T) {
	// argString called with non-string → error
	_, err := EvalString("string_length(42)", nil)
	if err == nil {
		t.Error("expected error for string_length(number)")
	}
	// argNumber called with non-number → error
	_, err = EvalString(`abs("x")`, nil)
	if err == nil {
		t.Error("expected error for abs(string)")
	}
	// argList called with non-list → error (sum on a non-number triggers collectNumbers error)
	_, err = EvalString(`sum("a", "b")`, nil)
	if err == nil {
		t.Error("expected error for sum(strings)")
	}
}

func TestArgWrongCount(t *testing.T) {
	_, err := EvalString("abs(1, 2)", nil)
	if err == nil {
		t.Error("expected error for abs with 2 args")
	}
}

func TestEvalDivisionByZero(t *testing.T) {
	v, err := EvalString("10 / 0", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v.(Null); !ok {
		t.Errorf("10/0 = %v, want null", v)
	}
}

func TestEvalParseError(t *testing.T) {
	_, err := EvalString("@@@invalid@@@", nil)
	if err == nil {
		t.Error("expected parse error")
	}
}
