package dmn

import (
	"fmt"
	"strings"

	"github.com/nomos/nomos/internal/dmn/feel"
)

// feelInputPlaceholder is the magic identifier the input cell evaluator
// substitutes for "?" — FEEL's implicit unary-test input reference.
// "?" is not a valid identifier in the feel package's lexer, so we
// rewrite it before parsing.
const feelInputPlaceholder = "__feel_input"

// evaluateInputEntry decides whether a rule's input cell matches the
// provided inputs. It first tries the simple FEEL unary-test grammar
// (literals, ranges, comparisons, lists, not). If that fails, it falls
// back to parsing the cell text as a full FEEL boolean expression with
// all column inputs in scope and "?" bound to the current column's
// value — matching how Camunda-style DMN editors author expressions
// like matches(dns_name, "...").
func evaluateInputEntry(entryText, colExpr string, inputs map[string]Value) (bool, error) {
	inputVal, ok := inputs[colExpr]
	if !ok {
		inputVal = NullVal{}
	}
	test, err := ParseUnaryTest(entryText)
	if err == nil {
		return test.Matches(inputVal), nil
	}
	// Fallback: full FEEL boolean expression.
	src := rewriteInputPlaceholder(strings.TrimSpace(entryText))
	expr, perr := feel.Parse(src)
	if perr != nil {
		// Surface the original unary-test error — it's the one the
		// user is most likely trying to write.
		return false, err
	}
	env := buildFeelEnv(inputs, inputVal)
	result, eerr := expr.Eval(env)
	if eerr != nil {
		return false, fmt.Errorf("FEEL evaluation error in %q: %w", entryText, eerr)
	}
	b, ok := result.(feel.Bool)
	if !ok {
		return false, fmt.Errorf("input entry %q did not evaluate to a boolean (got %s)", entryText, result.GoValue())
	}
	return bool(b), nil
}

// rewriteInputPlaceholder replaces "?" tokens with feelInputPlaceholder,
// skipping occurrences inside string literals.
func rewriteInputPlaceholder(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	inStr := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' && inStr && i+1 < len(s):
			sb.WriteByte(c)
			sb.WriteByte(s[i+1])
			i++
		case c == '"':
			inStr = !inStr
			sb.WriteByte(c)
		case c == '?' && !inStr:
			sb.WriteString(feelInputPlaceholder)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// buildFeelEnv converts the dmn inputs map (plus the current column's
// input value) into a feel.Env.
func buildFeelEnv(inputs map[string]Value, current Value) feel.Env {
	env := feel.Env{}
	for k, v := range inputs {
		env[k] = toFeelValue(v)
	}
	env[feelInputPlaceholder] = toFeelValue(current)
	return env
}

// toFeelValue bridges a dmn.Value to a feel.Value.
func toFeelValue(v Value) feel.Value {
	switch t := v.(type) {
	case NumberVal:
		return feel.Number(t.V)
	case StringVal:
		return feel.String(t.V)
	case BoolVal:
		return feel.Bool(t.V)
	case NullVal:
		return feel.Null{}
	case DateVal:
		return feel.Date{T: t.V}
	case DateTimeVal:
		return feel.DateTime{T: t.V}
	case ListVal:
		out := make(feel.List, len(t.Items))
		for i, item := range t.Items {
			out[i] = toFeelValue(item)
		}
		return out
	case nil:
		return feel.Null{}
	}
	return feel.Null{}
}
