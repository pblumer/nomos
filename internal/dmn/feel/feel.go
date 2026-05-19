package feel

import (
	"strconv"
	"time"
)

// Expression is a parsed, evaluable FEEL expression.
type Expression struct {
	src  string
	root Expr
}

// Parse compiles a FEEL expression into a reusable Expression.
func Parse(src string) (*Expression, error) {
	root, err := parse(src)
	if err != nil {
		return nil, err
	}
	return &Expression{src: src, root: root}, nil
}

// Eval evaluates the expression against the given environment.
func (e *Expression) Eval(env Env) (Value, error) {
	return eval(e.root, env)
}

// Source returns the original FEEL source.
func (e *Expression) Source() string { return e.src }

// EvalString is a shortcut: parse + eval. For repeated evaluations
// (e.g. table cells) prefer Parse once + Eval many.
func EvalString(src string, env Env) (Value, error) {
	expr, err := Parse(src)
	if err != nil {
		return nil, err
	}
	return expr.Eval(env)
}

// FromGo converts a Go value (as produced by json.Unmarshal or
// constructed by the caller) into a feel.Value. Strings that look
// like ISO-8601 dates are auto-detected.
func FromGo(v any) Value {
	if v == nil {
		return Null{}
	}
	switch t := v.(type) {
	case Value:
		return t
	case bool:
		return Bool(t)
	case float64:
		return Number(t)
	case float32:
		return Number(float64(t))
	case int:
		return Number(float64(t))
	case int64:
		return Number(float64(t))
	case string:
		if d, err := time.Parse("2006-01-02", t); err == nil {
			return Date{T: d}
		}
		if d, err := time.Parse(time.RFC3339, t); err == nil {
			return DateTime{T: d}
		}
		// Try number-as-string.
		if n, err := strconv.ParseFloat(t, 64); err == nil {
			return Number(n)
		}
		return String(t)
	case []any:
		out := make(List, len(t))
		for i, item := range t {
			out[i] = FromGo(item)
		}
		return out
	case map[string]any:
		out := Context{}
		for k, vv := range t {
			out[k] = FromGo(vv)
		}
		return out
	}
	return Null{}
}

// EnvFromGo is a convenience: build an Env from a Go map. Useful when
// callers pass JSON-decoded inputs directly.
func EnvFromGo(in map[string]any) Env {
	env := Env{}
	for k, v := range in {
		env[k] = FromGo(v)
	}
	return env
}
