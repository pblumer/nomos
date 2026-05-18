// Package dmn implements a DMN 1.5 decision table evaluator
// supporting the FEEL unary-test subset needed for governance rules.
package dmn

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Value is a FEEL-typed runtime value.
type Value interface {
	feelType() string
	GoValue() any
	String() string
}

// NumberVal holds a FEEL number (IEEE 754 double).
type NumberVal struct{ V float64 }

// StringVal holds a FEEL string.
type StringVal struct{ V string }

// BoolVal holds a FEEL boolean.
type BoolVal struct{ V bool }

// NullVal represents the FEEL null value.
type NullVal struct{}

// DateVal holds a FEEL date (date only, no time).
type DateVal struct{ V time.Time }

// DateTimeVal holds a FEEL date and time.
type DateTimeVal struct{ V time.Time }

// ListVal holds a FEEL list (used for COLLECT results).
type ListVal struct{ Items []Value }

func (v NumberVal) feelType() string   { return "number" }
func (v StringVal) feelType() string   { return "string" }
func (v BoolVal) feelType() string     { return "boolean" }
func (v NullVal) feelType() string     { return "null" }
func (v DateVal) feelType() string     { return "date" }
func (v DateTimeVal) feelType() string { return "date time" }
func (v ListVal) feelType() string     { return "list" }

func (v NumberVal) GoValue() any   { return v.V }
func (v StringVal) GoValue() any   { return v.V }
func (v BoolVal) GoValue() any     { return v.V }
func (v NullVal) GoValue() any     { return nil }
func (v DateVal) GoValue() any     { return v.V.Format("2006-01-02") }
func (v DateTimeVal) GoValue() any { return v.V.Format(time.RFC3339) }
func (v ListVal) GoValue() any {
	out := make([]any, len(v.Items))
	for i, item := range v.Items {
		out[i] = item.GoValue()
	}
	return out
}

func (v NumberVal) String() string {
	if v.V == math.Trunc(v.V) && !math.IsInf(v.V, 0) {
		return strconv.FormatInt(int64(v.V), 10)
	}
	return strconv.FormatFloat(v.V, 'f', -1, 64)
}
func (v StringVal) String() string   { return v.V }
func (v BoolVal) String() string     { return strconv.FormatBool(v.V) }
func (v NullVal) String() string     { return "null" }
func (v DateVal) String() string     { return v.V.Format("2006-01-02") }
func (v DateTimeVal) String() string { return v.V.Format(time.RFC3339) }
func (v ListVal) String() string {
	parts := make([]string, len(v.Items))
	for i, item := range v.Items {
		parts[i] = item.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// ToValue converts a Go value (from JSON unmarshal or direct use) to a FEEL Value.
func ToValue(v any, typeHint string) Value {
	if v == nil {
		return NullVal{}
	}
	switch t := v.(type) {
	case bool:
		return BoolVal{t}
	case float64:
		return NumberVal{t}
	case float32:
		return NumberVal{float64(t)}
	case int:
		return NumberVal{float64(t)}
	case int64:
		return NumberVal{float64(t)}
	case string:
		switch typeHint {
		case "date":
			if d, err := time.Parse("2006-01-02", t); err == nil {
				return DateVal{d}
			}
		case "date and time", "dateTime":
			for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05"} {
				if d, err := time.Parse(layout, t); err == nil {
					return DateTimeVal{d}
				}
			}
		case "boolean":
			if t == "true" {
				return BoolVal{true}
			}
			if t == "false" {
				return BoolVal{false}
			}
		case "number", "integer":
			if n, err := strconv.ParseFloat(t, 64); err == nil {
				return NumberVal{n}
			}
		}
		// Auto-detect date
		if d, err := time.Parse("2006-01-02", t); err == nil {
			return DateVal{d}
		}
		return StringVal{t}
	default:
		return StringVal{fmt.Sprintf("%v", v)}
	}
}

// feelEqual returns true if a == b under FEEL semantics.
func feelEqual(a, b Value) bool {
	switch av := a.(type) {
	case NumberVal:
		if bv, ok := b.(NumberVal); ok {
			return av.V == bv.V
		}
	case StringVal:
		if bv, ok := b.(StringVal); ok {
			return av.V == bv.V
		}
	case BoolVal:
		if bv, ok := b.(BoolVal); ok {
			return av.V == bv.V
		}
	case NullVal:
		_, ok := b.(NullVal)
		return ok
	case DateVal:
		if bv, ok := b.(DateVal); ok {
			return av.V.Equal(bv.V)
		}
	case DateTimeVal:
		if bv, ok := b.(DateTimeVal); ok {
			return av.V.Equal(bv.V)
		}
	}
	return false
}

// feelCompare evaluates v op ref. op is one of "<", ">", "<=", ">=", "=", "!=".
func feelCompare(v Value, op string, ref Value) bool {
	if op == "=" {
		return feelEqual(v, ref)
	}
	if op == "!=" {
		return !feelEqual(v, ref)
	}
	switch vv := v.(type) {
	case NumberVal:
		if rv, ok := ref.(NumberVal); ok {
			switch op {
			case "<":
				return vv.V < rv.V
			case "<=":
				return vv.V <= rv.V
			case ">":
				return vv.V > rv.V
			case ">=":
				return vv.V >= rv.V
			}
		}
	case StringVal:
		if rv, ok := ref.(StringVal); ok {
			switch op {
			case "<":
				return vv.V < rv.V
			case "<=":
				return vv.V <= rv.V
			case ">":
				return vv.V > rv.V
			case ">=":
				return vv.V >= rv.V
			}
		}
	case DateVal:
		if rv, ok := ref.(DateVal); ok {
			switch op {
			case "<":
				return vv.V.Before(rv.V)
			case "<=":
				return !vv.V.After(rv.V)
			case ">":
				return vv.V.After(rv.V)
			case ">=":
				return !vv.V.Before(rv.V)
			}
		}
	case DateTimeVal:
		if rv, ok := ref.(DateTimeVal); ok {
			switch op {
			case "<":
				return vv.V.Before(rv.V)
			case "<=":
				return !vv.V.After(rv.V)
			case ">":
				return vv.V.After(rv.V)
			case ">=":
				return !vv.V.Before(rv.V)
			}
		}
	}
	return false
}
