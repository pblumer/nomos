package feel

import (
	"testing"
)

func eq(t *testing.T, got Value, want any) {
	t.Helper()
	if g, ok := got.(Number); ok {
		if w, ok := want.(float64); ok && float64(g) == w {
			return
		}
		if w, ok := want.(int); ok && float64(g) == float64(w) {
			return
		}
	}
	if g, ok := got.(String); ok {
		if w, ok := want.(string); ok && string(g) == w {
			return
		}
	}
	if g, ok := got.(Bool); ok {
		if w, ok := want.(bool); ok && bool(g) == w {
			return
		}
	}
	if _, ok := got.(Null); ok && want == nil {
		return
	}
	t.Fatalf("got %v (%T), want %v (%T)", got, got, want, want)
}

func mustEval(t *testing.T, src string, env Env) Value {
	t.Helper()
	v, err := EvalString(src, env)
	if err != nil {
		t.Fatalf("EvalString(%q): %v", src, err)
	}
	return v
}

func TestLiterals(t *testing.T) {
	eq(t, mustEval(t, "42", nil), 42.0)
	eq(t, mustEval(t, `"hello"`, nil), "hello")
	eq(t, mustEval(t, "true", nil), true)
	eq(t, mustEval(t, "false", nil), false)
	eq(t, mustEval(t, "null", nil), nil)
}

func TestArithmetic(t *testing.T) {
	eq(t, mustEval(t, "1 + 2", nil), 3.0)
	eq(t, mustEval(t, "10 - 3", nil), 7.0)
	eq(t, mustEval(t, "4 * 5", nil), 20.0)
	eq(t, mustEval(t, "20 / 4", nil), 5.0)
	eq(t, mustEval(t, "2 ** 10", nil), 1024.0)
	eq(t, mustEval(t, "2 + 3 * 4", nil), 14.0)
	eq(t, mustEval(t, "(2 + 3) * 4", nil), 20.0)
	eq(t, mustEval(t, "-5 + 8", nil), 3.0)
	// Right-assoc power: 2**3**2 = 2**(3**2) = 2**9 = 512
	eq(t, mustEval(t, "2 ** 3 ** 2", nil), 512.0)
}

func TestDivByZero(t *testing.T) {
	v := mustEval(t, "5 / 0", nil)
	if _, ok := v.(Null); !ok {
		t.Fatalf("want Null, got %v", v)
	}
}

func TestComparison(t *testing.T) {
	eq(t, mustEval(t, "3 < 4", nil), true)
	eq(t, mustEval(t, "3 > 4", nil), false)
	eq(t, mustEval(t, "5 = 5", nil), true)
	eq(t, mustEval(t, "5 != 5", nil), false)
	eq(t, mustEval(t, `"a" < "b"`, nil), true)
}

func TestLogic(t *testing.T) {
	eq(t, mustEval(t, "true and false", nil), false)
	eq(t, mustEval(t, "true or false", nil), true)
	eq(t, mustEval(t, "not true", nil), false)
	// Short-circuit: rhs may evaluate to non-bool but result is bool.
	eq(t, mustEval(t, "false and (1 / 0)", nil), false)
	eq(t, mustEval(t, "true or (1 / 0)", nil), true)
}

func TestIfThenElse(t *testing.T) {
	eq(t, mustEval(t, "if 1 < 2 then 100 else 200", nil), 100.0)
	eq(t, mustEval(t, "if 1 > 2 then 100 else 200", nil), 200.0)
}

func TestNamesAndEnv(t *testing.T) {
	env := Env{"x": Number(7), "y": Number(3)}
	eq(t, mustEval(t, "x + y", env), 10.0)
	eq(t, mustEval(t, "x * 2 + y", env), 17.0)
}

func TestPathNavigation(t *testing.T) {
	env := EnvFromGo(map[string]any{
		"applicant": map[string]any{
			"age":    34,
			"income": 95000,
			"name":   "alice",
		},
	})
	eq(t, mustEval(t, "applicant.age", env), 34.0)
	eq(t, mustEval(t, "applicant.income > 50000", env), true)
	eq(t, mustEval(t, "applicant.name", env), "alice")
}

func TestContextLiteral(t *testing.T) {
	v := mustEval(t, "{a: 1, b: 2, c: 3}.b", nil)
	eq(t, v, 2.0)
}

func TestListLiteralAndIndex(t *testing.T) {
	eq(t, mustEval(t, "[10, 20, 30][2]", nil), 20.0)
	// 1-based; out of range yields null.
	v := mustEval(t, "[10, 20, 30][99]", nil)
	if _, ok := v.(Null); !ok {
		t.Fatalf("want Null, got %v", v)
	}
}

func TestBuiltinsString(t *testing.T) {
	eq(t, mustEval(t, `string_length("hello")`, nil), 5.0)
	eq(t, mustEval(t, `upper_case("abc")`, nil), "ABC")
	eq(t, mustEval(t, `lower_case("ABC")`, nil), "abc")
	eq(t, mustEval(t, `contains("hello world", "world")`, nil), true)
	eq(t, mustEval(t, `starts_with("abcdef", "abc")`, nil), true)
	eq(t, mustEval(t, `ends_with("abcdef", "def")`, nil), true)
	eq(t, mustEval(t, `substring("abcdef", 2, 3)`, nil), "bcd")
}

func TestBuiltinsNumber(t *testing.T) {
	eq(t, mustEval(t, "abs(-7)", nil), 7.0)
	eq(t, mustEval(t, "ceiling(1.2)", nil), 2.0)
	eq(t, mustEval(t, "floor(1.8)", nil), 1.0)
	eq(t, mustEval(t, "modulo(10, 3)", nil), 1.0)
}

func TestBuiltinsList(t *testing.T) {
	eq(t, mustEval(t, "sum([1, 2, 3, 4])", nil), 10.0)
	eq(t, mustEval(t, "sum(1, 2, 3, 4)", nil), 10.0)
	eq(t, mustEval(t, "count([1, 2, 3, 4])", nil), 4.0)
	eq(t, mustEval(t, "min([5, 1, 3])", nil), 1.0)
	eq(t, mustEval(t, "max([5, 1, 3])", nil), 5.0)
	eq(t, mustEval(t, "mean([2, 4, 6])", nil), 4.0)
	eq(t, mustEval(t, "list_contains([1, 2, 3], 2)", nil), true)
	eq(t, mustEval(t, "list_contains([1, 2, 3], 99)", nil), false)
	eq(t, mustEval(t, "all([true, true, true])", nil), true)
	eq(t, mustEval(t, "all([true, false, true])", nil), false)
	eq(t, mustEval(t, "any([false, false, true])", nil), true)
}

func TestBuiltinsDate(t *testing.T) {
	v := mustEval(t, `date("2026-05-19")`, nil)
	d, ok := v.(Date)
	if !ok {
		t.Fatalf("want Date, got %T", v)
	}
	if d.T.Year() != 2026 || d.T.Month() != 5 || d.T.Day() != 19 {
		t.Fatalf("date mismatch: %v", d.T)
	}
	v2 := mustEval(t, `date("2026-05-19").year`, nil)
	eq(t, v2, 2026.0)
	v3 := mustEval(t, `date(2026, 5, 19) = date("2026-05-19")`, nil)
	eq(t, v3, true)
}

func TestRealisticDecisionRule(t *testing.T) {
	// A literal-expression decision body, evaluated with input bindings.
	env := EnvFromGo(map[string]any{
		"applicant": map[string]any{
			"age":     17,
			"country": "DE",
			"income":  120000,
		},
	})
	src := `if applicant.age < 18 then "rejected: minor"
            else if applicant.income < 30000 then "rejected: income"
            else "approved"`
	got := mustEval(t, src, env)
	eq(t, got, "rejected: minor")

	env2 := EnvFromGo(map[string]any{
		"applicant": map[string]any{
			"age":     45,
			"country": "DE",
			"income":  120000,
		},
	})
	eq(t, mustEval(t, src, env2), "approved")
}

func TestReuseCompiledExpr(t *testing.T) {
	expr, err := Parse("amount * rate + fee")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		a, r, f, want float64
	}{
		{100, 0.1, 5, 15},
		{200, 0.2, 0, 40},
		{50, 1, 7, 57},
	} {
		env := Env{"amount": Number(c.a), "rate": Number(c.r), "fee": Number(c.f)}
		v, err := expr.Eval(env)
		if err != nil {
			t.Fatal(err)
		}
		if float64(v.(Number)) != c.want {
			t.Fatalf("got %v want %v", v, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, src := range []string{
		"1 +",
		"(1 + 2",
		"if true then 1",
		"{a: 1, b}",
	} {
		if _, err := Parse(src); err == nil {
			t.Errorf("expected error for %q", src)
		}
	}
}
