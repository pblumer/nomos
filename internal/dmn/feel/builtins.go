package feel

import (
	"math"
	"strings"
	"time"
)

// builtinFn is the signature of a FEEL built-in. Each implementation
// validates its own argument count and types.
type builtinFn func(args []Value) (Value, error)

type builtinRef struct {
	name string
	fn   builtinFn
}

func (builtinRef) feelType() string { return "function" }
func (b builtinRef) GoValue() any   { return b.name }

// Stufe-2 catalogue. Names follow the DMN-1.5 FEEL spec; spaces in
// official names are not supported by our identifier syntax yet, so
// underscored aliases are exposed where needed.
var builtins = map[string]builtinFn{
	// String
	"string length": bStringLength,
	"upper case":    bUpperCase,
	"lower case":    bLowerCase,
	"contains":      bContains,
	"starts with":   bStartsWith,
	"ends with":     bEndsWith,
	"substring":     bSubstring,
	// Number
	"abs":     bAbs,
	"ceiling": bCeiling,
	"floor":   bFloor,
	"modulo":  bModulo,
	// List
	"count":         bCount,
	"sum":           bSum,
	"min":           bMin,
	"max":           bMax,
	"mean":          bMean,
	"list contains": bListContains,
	"all":           bAll,
	"any":           bAny,
	// Date / time
	"today": bToday,
	"now":   bNow,
	"date":  bDate,
}

// lookupBuiltin matches a name first as written, then with
// underscores replaced by spaces (so "string_length" finds
// "string length"). FEEL spec names are space-separated.
func lookupBuiltin(name string) builtinFn {
	if fn, ok := builtins[name]; ok {
		return fn
	}
	if fn, ok := builtins[strings.ReplaceAll(name, "_", " ")]; ok {
		return fn
	}
	return nil
}

func argCount(args []Value, want int) error {
	if len(args) != want {
		return evalErr("expected %d arg(s), got %d", want, len(args))
	}
	return nil
}

func argString(args []Value, idx int) (string, error) {
	v, ok := args[idx].(String)
	if !ok {
		return "", evalErr("arg %d: expected string, got %s", idx, args[idx].feelType())
	}
	return string(v), nil
}

func argNumber(args []Value, idx int) (float64, error) {
	v, ok := args[idx].(Number)
	if !ok {
		return 0, evalErr("arg %d: expected number, got %s", idx, args[idx].feelType())
	}
	return float64(v), nil
}

func argList(args []Value, idx int) (List, error) {
	v, ok := args[idx].(List)
	if !ok {
		return nil, evalErr("arg %d: expected list, got %s", idx, args[idx].feelType())
	}
	return v, nil
}

// String built-ins
func bStringLength(args []Value) (Value, error) {
	if err := argCount(args, 1); err != nil {
		return nil, err
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	return Number(len([]rune(s))), nil
}

func bUpperCase(args []Value) (Value, error) {
	if err := argCount(args, 1); err != nil {
		return nil, err
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	return String(strings.ToUpper(s)), nil
}

func bLowerCase(args []Value) (Value, error) {
	if err := argCount(args, 1); err != nil {
		return nil, err
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	return String(strings.ToLower(s)), nil
}

func bContains(args []Value) (Value, error) {
	if err := argCount(args, 2); err != nil {
		return nil, err
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	sub, err := argString(args, 1)
	if err != nil {
		return nil, err
	}
	return Bool(strings.Contains(s, sub)), nil
}

func bStartsWith(args []Value) (Value, error) {
	if err := argCount(args, 2); err != nil {
		return nil, err
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	pre, err := argString(args, 1)
	if err != nil {
		return nil, err
	}
	return Bool(strings.HasPrefix(s, pre)), nil
}

func bEndsWith(args []Value) (Value, error) {
	if err := argCount(args, 2); err != nil {
		return nil, err
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	suf, err := argString(args, 1)
	if err != nil {
		return nil, err
	}
	return Bool(strings.HasSuffix(s, suf)), nil
}

func bSubstring(args []Value) (Value, error) {
	// substring(s, start) or substring(s, start, length); FEEL is 1-based.
	if len(args) != 2 && len(args) != 3 {
		return nil, evalErr("substring expects 2 or 3 args, got %d", len(args))
	}
	s, err := argString(args, 0)
	if err != nil {
		return nil, err
	}
	startF, err := argNumber(args, 1)
	if err != nil {
		return nil, err
	}
	runes := []rune(s)
	start := int(startF) - 1
	if start < 0 {
		start = len(runes) + int(startF)
	}
	if start < 0 {
		start = 0
	}
	if start > len(runes) {
		start = len(runes)
	}
	end := len(runes)
	if len(args) == 3 {
		lenF, err := argNumber(args, 2)
		if err != nil {
			return nil, err
		}
		end = start + int(lenF)
		if end > len(runes) {
			end = len(runes)
		}
		if end < start {
			end = start
		}
	}
	return String(string(runes[start:end])), nil
}

// Number built-ins
func bAbs(args []Value) (Value, error) {
	if err := argCount(args, 1); err != nil {
		return nil, err
	}
	n, err := argNumber(args, 0)
	if err != nil {
		return nil, err
	}
	return Number(math.Abs(n)), nil
}

func bCeiling(args []Value) (Value, error) {
	if err := argCount(args, 1); err != nil {
		return nil, err
	}
	n, err := argNumber(args, 0)
	if err != nil {
		return nil, err
	}
	return Number(math.Ceil(n)), nil
}

func bFloor(args []Value) (Value, error) {
	if err := argCount(args, 1); err != nil {
		return nil, err
	}
	n, err := argNumber(args, 0)
	if err != nil {
		return nil, err
	}
	return Number(math.Floor(n)), nil
}

func bModulo(args []Value) (Value, error) {
	if err := argCount(args, 2); err != nil {
		return nil, err
	}
	a, err := argNumber(args, 0)
	if err != nil {
		return nil, err
	}
	b, err := argNumber(args, 1)
	if err != nil {
		return nil, err
	}
	if b == 0 {
		return Null{}, nil
	}
	return Number(math.Mod(a, b)), nil
}

// List built-ins
//
// FEEL allows both list(...) calls and varargs: count([1,2,3]) and
// count(1, 2, 3). We accept either: a single List arg, or a flat
// arglist.
func collectNumbers(args []Value) ([]float64, error) {
	var nums []float64
	if len(args) == 1 {
		if l, ok := args[0].(List); ok {
			for i, v := range l {
				n, ok := v.(Number)
				if !ok {
					return nil, evalErr("list element %d: expected number, got %s", i, v.feelType())
				}
				nums = append(nums, float64(n))
			}
			return nums, nil
		}
	}
	for i, v := range args {
		n, ok := v.(Number)
		if !ok {
			return nil, evalErr("arg %d: expected number, got %s", i, v.feelType())
		}
		nums = append(nums, float64(n))
	}
	return nums, nil
}

func bCount(args []Value) (Value, error) {
	if len(args) == 1 {
		if l, ok := args[0].(List); ok {
			return Number(len(l)), nil
		}
	}
	return Number(len(args)), nil
}

func bSum(args []Value) (Value, error) {
	nums, err := collectNumbers(args)
	if err != nil {
		return nil, err
	}
	var s float64
	for _, n := range nums {
		s += n
	}
	return Number(s), nil
}

func bMin(args []Value) (Value, error) {
	nums, err := collectNumbers(args)
	if err != nil {
		return nil, err
	}
	if len(nums) == 0 {
		return Null{}, nil
	}
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return Number(m), nil
}

func bMax(args []Value) (Value, error) {
	nums, err := collectNumbers(args)
	if err != nil {
		return nil, err
	}
	if len(nums) == 0 {
		return Null{}, nil
	}
	m := nums[0]
	for _, n := range nums[1:] {
		if n > m {
			m = n
		}
	}
	return Number(m), nil
}

func bMean(args []Value) (Value, error) {
	nums, err := collectNumbers(args)
	if err != nil {
		return nil, err
	}
	if len(nums) == 0 {
		return Null{}, nil
	}
	var s float64
	for _, n := range nums {
		s += n
	}
	return Number(s / float64(len(nums))), nil
}

func bListContains(args []Value) (Value, error) {
	if err := argCount(args, 2); err != nil {
		return nil, err
	}
	l, err := argList(args, 0)
	if err != nil {
		return nil, err
	}
	for _, v := range l {
		if equalValues(v, args[1]) {
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func bAll(args []Value) (Value, error) {
	values := args
	if len(args) == 1 {
		if l, ok := args[0].(List); ok {
			values = l
		}
	}
	for _, v := range values {
		if !truthy(v) {
			return Bool(false), nil
		}
	}
	return Bool(true), nil
}

func bAny(args []Value) (Value, error) {
	values := args
	if len(args) == 1 {
		if l, ok := args[0].(List); ok {
			values = l
		}
	}
	for _, v := range values {
		if truthy(v) {
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

// Date / time built-ins
func bToday(args []Value) (Value, error) {
	if err := argCount(args, 0); err != nil {
		return nil, err
	}
	t := time.Now().UTC()
	return Date{T: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}, nil
}

func bNow(args []Value) (Value, error) {
	if err := argCount(args, 0); err != nil {
		return nil, err
	}
	return DateTime{T: time.Now().UTC()}, nil
}

// date(string) or date(year, month, day)
func bDate(args []Value) (Value, error) {
	switch len(args) {
	case 1:
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, evalErr("invalid date %q", s)
		}
		return Date{T: t}, nil
	case 3:
		y, err := argNumber(args, 0)
		if err != nil {
			return nil, err
		}
		m, err := argNumber(args, 1)
		if err != nil {
			return nil, err
		}
		d, err := argNumber(args, 2)
		if err != nil {
			return nil, err
		}
		return Date{T: time.Date(int(y), time.Month(int(m)), int(d), 0, 0, 0, 0, time.UTC)}, nil
	}
	return nil, evalErr("date expects 1 or 3 args, got %d", len(args))
}
