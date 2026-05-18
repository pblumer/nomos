package dmn

import (
	"fmt"
	"math"
)

// matchedRow holds a matched rule and its evaluated outputs.
type matchedRow struct {
	rule    RuleRow
	outputs []Value
}

// Result is the outcome of evaluating a decision table.
type Result struct {
	// Outputs holds the evaluated output values (one entry per output column).
	// For COLLECT without aggregation, each value is a []any slice.
	Outputs map[string]any `json:"outputs"`

	// MatchedRules lists the IDs of all matched rules (in evaluation order).
	MatchedRules []string `json:"matched_rules,omitempty"`

	// HitPolicy is the hit policy that was applied.
	HitPolicy string `json:"hit_policy"`

	// Aggregation is the aggregation function applied (COLLECT only).
	Aggregation string `json:"aggregation,omitempty"`
}

// Evaluate runs the decision table against the provided input map and returns the result.
// inputTypeHints maps input column names to their DMN typeRef (used for type coercion).
func Evaluate(table *Table, inputs map[string]any) (*Result, error) {
	// Build type hints map: expression → typeRef
	typeHints := make(map[string]string, len(table.Inputs))
	for _, col := range table.Inputs {
		typeHints[col.Expression] = col.TypeRef
	}

	// Coerce raw inputs to FEEL Values
	feelInputs := make(map[string]Value, len(inputs))
	for k, v := range inputs {
		hint := typeHints[k]
		feelInputs[k] = ToValue(v, hint)
	}

	// Find matching rules
	var matched []matchedRow

	for _, rule := range table.Rules {
		ok, err := ruleMatches(table, rule, feelInputs)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", rule.ID, err)
		}
		if !ok {
			continue
		}
		outs, err := evaluateOutputs(table, rule)
		if err != nil {
			return nil, fmt.Errorf("rule %q output evaluation: %w", rule.ID, err)
		}
		matched = append(matched, matchedRow{rule, outs})
	}

	res := &Result{
		HitPolicy:   table.HitPolicy,
		Aggregation: table.Aggregation,
	}

	switch table.HitPolicy {
	case "UNIQUE":
		if len(matched) == 0 {
			res.Outputs = map[string]any{}
			return res, nil
		}
		if len(matched) > 1 {
			ids := make([]string, len(matched))
			for i, m := range matched {
				ids[i] = m.rule.ID
			}
			return nil, fmt.Errorf("UNIQUE hit policy violated: %d rules matched: %v", len(matched), ids)
		}
		res.MatchedRules = []string{matched[0].rule.ID}
		res.Outputs = outputsToMap(table, matched[0].outputs)

	case "FIRST":
		if len(matched) == 0 {
			res.Outputs = map[string]any{}
			return res, nil
		}
		res.MatchedRules = []string{matched[0].rule.ID}
		res.Outputs = outputsToMap(table, matched[0].outputs)

	case "ANY":
		if len(matched) == 0 {
			res.Outputs = map[string]any{}
			return res, nil
		}
		// All matched outputs must be identical.
		first := outputsToMap(table, matched[0].outputs)
		for _, m := range matched[1:] {
			other := outputsToMap(table, m.outputs)
			if !mapsEqual(first, other) {
				return nil, fmt.Errorf("ANY hit policy violated: matched rules have different outputs")
			}
		}
		for _, m := range matched {
			res.MatchedRules = append(res.MatchedRules, m.rule.ID)
		}
		res.Outputs = first

	case "COLLECT":
		for _, m := range matched {
			res.MatchedRules = append(res.MatchedRules, m.rule.ID)
		}
		res.Outputs = collectOutputs(table, matched, table.Aggregation)

	case "RULE ORDER":
		for _, m := range matched {
			res.MatchedRules = append(res.MatchedRules, m.rule.ID)
		}
		res.Outputs = collectOutputs(table, matched, "")

	default:
		// Fallback: behave like FIRST.
		if len(matched) == 0 {
			res.Outputs = map[string]any{}
			return res, nil
		}
		res.MatchedRules = []string{matched[0].rule.ID}
		res.Outputs = outputsToMap(table, matched[0].outputs)
	}

	return res, nil
}

// ruleMatches returns true if all input entries of the rule match the provided inputs.
func ruleMatches(table *Table, rule RuleRow, inputs map[string]Value) (bool, error) {
	for i, col := range table.Inputs {
		var entryText string
		if i < len(rule.InputEntries) {
			entryText = rule.InputEntries[i]
		}
		test, err := ParseUnaryTest(entryText)
		if err != nil {
			return false, fmt.Errorf("input %q: %w", col.Expression, err)
		}
		inputVal, ok := inputs[col.Expression]
		if !ok {
			inputVal = NullVal{}
		}
		if !test.Matches(inputVal) {
			return false, nil
		}
	}
	return true, nil
}

// evaluateOutputs parses and returns the output values for a matched rule.
func evaluateOutputs(table *Table, rule RuleRow) ([]Value, error) {
	vals := make([]Value, len(table.Outputs))
	for i, col := range table.Outputs {
		var entryText string
		if i < len(rule.OutputEntries) {
			entryText = rule.OutputEntries[i]
		}
		v, err := ParseOutputExpression(entryText)
		if err != nil {
			return nil, fmt.Errorf("output %q: %w", col.Name, err)
		}
		// Coerce to declared type
		vals[i] = coerceOutput(v, col.TypeRef)
	}
	return vals, nil
}

// coerceOutput converts an output Value to the declared typeRef where needed.
func coerceOutput(v Value, typeRef string) Value {
	if _, isNull := v.(NullVal); isNull {
		return v
	}
	switch typeRef {
	case "boolean":
		if sv, ok := v.(StringVal); ok {
			if sv.V == "true" {
				return BoolVal{true}
			}
			if sv.V == "false" {
				return BoolVal{false}
			}
		}
	case "number", "integer":
		if sv, ok := v.(StringVal); ok {
			var n float64
			if _, err := fmt.Sscanf(sv.V, "%f", &n); err == nil {
				return NumberVal{n}
			}
		}
	}
	return v
}

// outputsToMap converts a slice of Values to a name→GoValue map.
func outputsToMap(table *Table, vals []Value) map[string]any {
	m := make(map[string]any, len(table.Outputs))
	for i, col := range table.Outputs {
		if i < len(vals) {
			m[col.Name] = vals[i].GoValue()
		}
	}
	return m
}

// collectOutputs builds COLLECT results, optionally applying an aggregation.
func collectOutputs(table *Table, matched []matchedRow, aggregation string) map[string]any {
	if len(matched) == 0 {
		return map[string]any{}
	}
	if aggregation == "" {
		// No aggregation: each output column returns a list.
		m := make(map[string]any, len(table.Outputs))
		for i, col := range table.Outputs {
			list := make([]any, 0, len(matched))
			for _, row := range matched {
				if i < len(row.outputs) {
					list = append(list, row.outputs[i].GoValue())
				}
			}
			m[col.Name] = list
		}
		return m
	}

	// Aggregation over numeric outputs.
	m := make(map[string]any, len(table.Outputs))
	for i, col := range table.Outputs {
		var nums []float64
		for _, row := range matched {
			if i < len(row.outputs) {
				if nv, ok := row.outputs[i].(NumberVal); ok {
					nums = append(nums, nv.V)
				}
			}
		}
		switch aggregation {
		case "SUM":
			sum := 0.0
			for _, n := range nums {
				sum += n
			}
			m[col.Name] = sum
		case "MIN":
			if len(nums) == 0 {
				m[col.Name] = nil
			} else {
				min := nums[0]
				for _, n := range nums[1:] {
					if n < min {
						min = n
					}
				}
				m[col.Name] = min
			}
		case "MAX":
			if len(nums) == 0 {
				m[col.Name] = nil
			} else {
				max := nums[0]
				for _, n := range nums[1:] {
					if n > max {
						max = n
					}
				}
				m[col.Name] = max
			}
		case "COUNT":
			m[col.Name] = float64(len(nums))
		default:
			m[col.Name] = nil
		}
	}
	return m
}

// mapsEqual performs a shallow string-equal comparison of two output maps.
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}

// ensure math is used (for potential future use of Inf/NaN checks)
var _ = math.IsNaN
