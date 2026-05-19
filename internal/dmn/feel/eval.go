package feel

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Value is the runtime value type for the FEEL Common Subset.
// We keep it self-contained inside this package; conversion helpers
// bridge to/from the parent internal/dmn Value types in feel.go.
type Value interface {
	feelType() string
	GoValue() any
}

type Number float64
type String string
type Bool bool
type Null struct{}
type Date struct{ T time.Time }
type DateTime struct{ T time.Time }
type List []Value
type Context map[string]Value

func (Number) feelType() string   { return "number" }
func (String) feelType() string   { return "string" }
func (Bool) feelType() string     { return "boolean" }
func (Null) feelType() string     { return "null" }
func (Date) feelType() string     { return "date" }
func (DateTime) feelType() string { return "date and time" }
func (List) feelType() string     { return "list" }
func (Context) feelType() string  { return "context" }

func (n Number) GoValue() any { return float64(n) }
func (s String) GoValue() any { return string(s) }
func (b Bool) GoValue() any   { return bool(b) }
func (Null) GoValue() any     { return nil }
func (d Date) GoValue() any   { return d.T.Format("2006-01-02") }
func (d DateTime) GoValue() any {
	return d.T.Format(time.RFC3339)
}
func (l List) GoValue() any {
	out := make([]any, len(l))
	for i, v := range l {
		out[i] = v.GoValue()
	}
	return out
}
func (c Context) GoValue() any {
	out := make(map[string]any, len(c))
	for k, v := range c {
		out[k] = v.GoValue()
	}
	return out
}

// Env is a lookup environment for free names (input variables).
// Names are matched first against the entries map, then as the key of
// a nested Context.
type Env map[string]Value

func (e Env) lookup(name string) (Value, bool) {
	v, ok := e[name]
	return v, ok
}

type evalError struct{ msg string }

func (e *evalError) Error() string { return "eval error: " + e.msg }

func evalErr(format string, args ...any) error {
	return &evalError{msg: fmt.Sprintf(format, args...)}
}

func eval(node Expr, env Env) (Value, error) {
	switch n := node.(type) {
	case numberLit:
		return Number(n.v), nil
	case stringLit:
		return String(n.v), nil
	case boolLit:
		return Bool(n.v), nil
	case nullLit:
		return Null{}, nil
	case nameRef:
		if v, ok := env.lookup(n.name); ok {
			return v, nil
		}
		if fn := lookupBuiltin(n.name); fn != nil {
			return builtinRef{name: n.name, fn: fn}, nil
		}
		return Null{}, nil
	case pathExpr:
		target, err := eval(n.target, env)
		if err != nil {
			return nil, err
		}
		return navigatePath(target, n.name)
	case unaryExpr:
		rhs, err := eval(n.rhs, env)
		if err != nil {
			return nil, err
		}
		return evalUnary(n.op, rhs)
	case binaryExpr:
		return evalBinary(n.op, n.lhs, n.rhs, env)
	case ifExpr:
		c, err := eval(n.cond, env)
		if err != nil {
			return nil, err
		}
		if truthy(c) {
			return eval(n.then, env)
		}
		return eval(n.els, env)
	case listExpr:
		out := make(List, 0, len(n.items))
		for _, it := range n.items {
			v, err := eval(it, env)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case contextExpr:
		out := Context{}
		for _, e := range n.entries {
			v, err := eval(e.value, env)
			if err != nil {
				return nil, err
			}
			out[e.key] = v
		}
		return out, nil
	case callExpr:
		return evalCall(n, env)
	}
	return nil, evalErr("unknown node %T", node)
}

func navigatePath(target Value, name string) (Value, error) {
	switch t := target.(type) {
	case Context:
		if v, ok := t[name]; ok {
			return v, nil
		}
		return Null{}, nil
	case Date:
		switch name {
		case "year":
			return Number(t.T.Year()), nil
		case "month":
			return Number(int(t.T.Month())), nil
		case "day":
			return Number(t.T.Day()), nil
		}
	case DateTime:
		switch name {
		case "year":
			return Number(t.T.Year()), nil
		case "month":
			return Number(int(t.T.Month())), nil
		case "day":
			return Number(t.T.Day()), nil
		case "hour":
			return Number(t.T.Hour()), nil
		case "minute":
			return Number(t.T.Minute()), nil
		case "second":
			return Number(t.T.Second()), nil
		}
	}
	return Null{}, nil
}

func evalUnary(op string, v Value) (Value, error) {
	switch op {
	case "-":
		if n, ok := v.(Number); ok {
			return -n, nil
		}
		return nil, evalErr("'-' requires number, got %s", v.feelType())
	case "not":
		return Bool(!truthy(v)), nil
	}
	return nil, evalErr("unknown unary %q", op)
}

func evalBinary(op string, lhsN, rhsN Expr, env Env) (Value, error) {
	// Short-circuit for and/or.
	if op == "and" {
		l, err := eval(lhsN, env)
		if err != nil {
			return nil, err
		}
		if !truthy(l) {
			return Bool(false), nil
		}
		r, err := eval(rhsN, env)
		if err != nil {
			return nil, err
		}
		return Bool(truthy(r)), nil
	}
	if op == "or" {
		l, err := eval(lhsN, env)
		if err != nil {
			return nil, err
		}
		if truthy(l) {
			return Bool(true), nil
		}
		r, err := eval(rhsN, env)
		if err != nil {
			return nil, err
		}
		return Bool(truthy(r)), nil
	}
	lhs, err := eval(lhsN, env)
	if err != nil {
		return nil, err
	}
	rhs, err := eval(rhsN, env)
	if err != nil {
		return nil, err
	}
	switch op {
	case "+":
		return addOp(lhs, rhs)
	case "-":
		return arithOp(lhs, rhs, op)
	case "*":
		return arithOp(lhs, rhs, op)
	case "/":
		return arithOp(lhs, rhs, op)
	case "**":
		return arithOp(lhs, rhs, op)
	case "=":
		return Bool(equalValues(lhs, rhs)), nil
	case "!=":
		return Bool(!equalValues(lhs, rhs)), nil
	case "<", "<=", ">", ">=":
		return compareOp(lhs, rhs, op)
	}
	return nil, evalErr("unknown binary %q", op)
}

func addOp(a, b Value) (Value, error) {
	if as, ok := a.(String); ok {
		if bs, ok := b.(String); ok {
			return as + bs, nil
		}
	}
	if al, ok := a.(List); ok {
		if bl, ok := b.(List); ok {
			out := make(List, 0, len(al)+len(bl))
			out = append(out, al...)
			out = append(out, bl...)
			return out, nil
		}
	}
	return arithOp(a, b, "+")
}

func arithOp(a, b Value, op string) (Value, error) {
	an, ok1 := a.(Number)
	bn, ok2 := b.(Number)
	if !ok1 || !ok2 {
		return nil, evalErr("'%s' requires numbers, got %s and %s", op, a.feelType(), b.feelType())
	}
	switch op {
	case "+":
		return an + bn, nil
	case "-":
		return an - bn, nil
	case "*":
		return an * bn, nil
	case "/":
		if bn == 0 {
			return Null{}, nil
		}
		return an / bn, nil
	case "**":
		return Number(math.Pow(float64(an), float64(bn))), nil
	}
	return nil, evalErr("arith op %q not implemented", op)
}

func compareOp(a, b Value, op string) (Value, error) {
	switch av := a.(type) {
	case Number:
		bv, ok := b.(Number)
		if !ok {
			return nil, evalErr("compare type mismatch: number vs %s", b.feelType())
		}
		return Bool(cmpFloat(float64(av), float64(bv), op)), nil
	case String:
		bv, ok := b.(String)
		if !ok {
			return nil, evalErr("compare type mismatch: string vs %s", b.feelType())
		}
		return Bool(cmpStr(string(av), string(bv), op)), nil
	case Date:
		bv, ok := b.(Date)
		if !ok {
			return nil, evalErr("compare type mismatch: date vs %s", b.feelType())
		}
		return Bool(cmpTime(av.T, bv.T, op)), nil
	case DateTime:
		bv, ok := b.(DateTime)
		if !ok {
			return nil, evalErr("compare type mismatch: date and time vs %s", b.feelType())
		}
		return Bool(cmpTime(av.T, bv.T, op)), nil
	}
	return nil, evalErr("cannot compare %s", a.feelType())
}

func cmpFloat(a, b float64, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}
func cmpStr(a, b, op string) bool {
	switch op {
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}
func cmpTime(a, b time.Time, op string) bool {
	switch op {
	case "<":
		return a.Before(b)
	case "<=":
		return !a.After(b)
	case ">":
		return a.After(b)
	case ">=":
		return !a.Before(b)
	}
	return false
}

func equalValues(a, b Value) bool {
	if _, ok := a.(Null); ok {
		_, bn := b.(Null)
		return bn
	}
	if _, ok := b.(Null); ok {
		return false
	}
	switch av := a.(type) {
	case Number:
		bv, ok := b.(Number)
		return ok && av == bv
	case String:
		bv, ok := b.(String)
		return ok && av == bv
	case Bool:
		bv, ok := b.(Bool)
		return ok && av == bv
	case Date:
		bv, ok := b.(Date)
		return ok && av.T.Equal(bv.T)
	case DateTime:
		bv, ok := b.(DateTime)
		return ok && av.T.Equal(bv.T)
	case List:
		bv, ok := b.(List)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !equalValues(av[i], bv[i]) {
				return false
			}
		}
		return true
	case Context:
		bv, ok := b.(Context)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			bvv, ok := bv[k]
			if !ok || !equalValues(v, bvv) {
				return false
			}
		}
		return true
	}
	return false
}

func truthy(v Value) bool {
	switch t := v.(type) {
	case Bool:
		return bool(t)
	case Null:
		return false
	}
	return false
}

func evalCall(n callExpr, env Env) (Value, error) {
	// Special-case the synthetic indexing operator inserted by the parser.
	if nr, ok := n.callee.(nameRef); ok && nr.name == "__index" {
		target, err := eval(n.args[0], env)
		if err != nil {
			return nil, err
		}
		idx, err := eval(n.args[1], env)
		if err != nil {
			return nil, err
		}
		return indexInto(target, idx)
	}
	callee, err := eval(n.callee, env)
	if err != nil {
		return nil, err
	}
	fn, ok := callee.(builtinRef)
	if !ok {
		return nil, evalErr("not callable: %s", callee.feelType())
	}
	args := make([]Value, len(n.args))
	for i, a := range n.args {
		v, err := eval(a, env)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	return fn.fn(args)
}

func indexInto(target, idx Value) (Value, error) {
	switch t := target.(type) {
	case List:
		n, ok := idx.(Number)
		if !ok {
			return nil, evalErr("list index must be number, got %s", idx.feelType())
		}
		// FEEL lists are 1-based.
		i := int(n) - 1
		if i < 0 || i >= len(t) {
			return Null{}, nil
		}
		return t[i], nil
	case Context:
		s, ok := idx.(String)
		if !ok {
			return nil, evalErr("context index must be string, got %s", idx.feelType())
		}
		if v, ok := t[string(s)]; ok {
			return v, nil
		}
		return Null{}, nil
	}
	return nil, evalErr("cannot index into %s", target.feelType())
}

// numberFromString is a small helper used by built-ins and the date
// constructor.
func numberFromString(s string) (float64, bool) {
	n, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
