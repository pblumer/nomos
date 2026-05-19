package dmn

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UnaryTest is a parsed FEEL unary-test expression (decision table input cell).
type UnaryTest interface {
	Matches(input Value) bool
	String() string
}

// --- concrete test types ---

type anyTest struct{}

func (anyTest) Matches(Value) bool { return true }
func (anyTest) String() string     { return "-" }

type nullTest struct{}

func (nullTest) Matches(v Value) bool { _, ok := v.(NullVal); return ok }
func (nullTest) String() string       { return "null" }

type literalTest struct{ val Value }

func (t literalTest) Matches(v Value) bool { return feelEqual(t.val, v) }
func (t literalTest) String() string       { return t.val.String() }

type compareTest struct {
	op  string
	val Value
}

func (t compareTest) Matches(v Value) bool { return feelCompare(v, t.op, t.val) }
func (t compareTest) String() string       { return t.op + " " + t.val.String() }

type rangeTest struct {
	lower, upper         Value
	lowerIncl, upperIncl bool
}

func (t rangeTest) Matches(v Value) bool {
	lowerOp := ">"
	if t.lowerIncl {
		lowerOp = ">="
	}
	upperOp := "<"
	if t.upperIncl {
		upperOp = "<="
	}
	return feelCompare(v, lowerOp, t.lower) && feelCompare(v, upperOp, t.upper)
}

func (t rangeTest) String() string {
	l := "("
	if t.lowerIncl {
		l = "["
	}
	r := ")"
	if t.upperIncl {
		r = "]"
	}
	return l + t.lower.String() + ".." + t.upper.String() + r
}

// listTest is a set of alternatives with OR semantics (comma-separated in DMN cells).
type listTest struct{ tests []UnaryTest }

func (t listTest) Matches(v Value) bool {
	for _, ut := range t.tests {
		if ut.Matches(v) {
			return true
		}
	}
	return false
}

func (t listTest) String() string {
	parts := make([]string, len(t.tests))
	for i, ut := range t.tests {
		parts[i] = ut.String()
	}
	return strings.Join(parts, ", ")
}

// notTest negates a set of alternatives.
type notTest struct{ tests []UnaryTest }

func (t notTest) Matches(v Value) bool {
	for _, ut := range t.tests {
		if ut.Matches(v) {
			return false
		}
	}
	return true
}

func (t notTest) String() string {
	parts := make([]string, len(t.tests))
	for i, ut := range t.tests {
		parts[i] = ut.String()
	}
	return "not(" + strings.Join(parts, ", ") + ")"
}

// --- parser ---

// ParseUnaryTest parses a FEEL unary-test string (decision table input entry).
// Empty string or "-" → anyTest (matches anything).
// Comma-separated entries → OR semantics.
func ParseUnaryTest(expr string) (UnaryTest, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" || expr == "-" {
		return anyTest{}, nil
	}
	p := &feelParser{input: expr}
	tests, err := p.parseUnaryTests()
	if err != nil {
		return nil, fmt.Errorf("FEEL parse error in %q: %w", expr, err)
	}
	if len(tests) == 1 {
		return tests[0], nil
	}
	return listTest{tests}, nil
}

// ParseOutputExpression parses a FEEL output entry (a simple literal).
// Returns nil for empty / "-" (means null/no value).
func ParseOutputExpression(expr string) (Value, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" || expr == "-" || expr == "null" {
		return NullVal{}, nil
	}
	p := &feelParser{input: expr}
	return p.parseLiteral()
}

// --- feel lexer/parser ---

type feelParser struct {
	input string
	pos   int
}

func (p *feelParser) remaining() string { return p.input[p.pos:] }

func (p *feelParser) skipWS() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func (p *feelParser) peek() byte {
	p.skipWS()
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *feelParser) hasPrefix(s string) bool {
	p.skipWS()
	return strings.HasPrefix(p.input[p.pos:], s)
}

func (p *feelParser) parseUnaryTests() ([]UnaryTest, error) {
	var tests []UnaryTest
	for {
		t, err := p.parseOneUnaryTest()
		if err != nil {
			return nil, err
		}
		tests = append(tests, t)
		p.skipWS()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++
		} else {
			break
		}
	}
	return tests, nil
}

func (p *feelParser) parseOneUnaryTest() (UnaryTest, error) {
	p.skipWS()
	if p.pos >= len(p.input) {
		return anyTest{}, nil
	}
	ch := p.input[p.pos]

	// not(...)
	if p.hasPrefix("not(") {
		p.pos += 4
		inner, err := p.parseUnaryTests()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if p.pos < len(p.input) && p.input[p.pos] == ')' {
			p.pos++
		}
		return notTest{inner}, nil
	}

	// Interval: [a..b], (a..b), [a..b), (a..b]
	if ch == '[' || ch == '(' {
		return p.parseInterval()
	}

	// Comparison operators: <, >, <=, >=, !=, =
	if ch == '<' || ch == '>' || ch == '!' || ch == '=' {
		return p.parseComparison()
	}

	// null
	if p.hasPrefix("null") && (len(p.remaining()) == 4 || !isIdentChar(p.input[p.pos+4])) {
		p.pos += 4
		return nullTest{}, nil
	}

	// literal value (string, bool, number, date)
	val, err := p.parseLiteral()
	if err != nil {
		return nil, err
	}
	return literalTest{val}, nil
}

func (p *feelParser) parseComparison() (UnaryTest, error) {
	var op string
	switch {
	case p.hasPrefix("<="):
		op = "<="
		p.pos += 2
	case p.hasPrefix(">="):
		op = ">="
		p.pos += 2
	case p.hasPrefix("!="):
		op = "!="
		p.pos += 2
	case p.hasPrefix("<"):
		op = "<"
		p.pos++
	case p.hasPrefix(">"):
		op = ">"
		p.pos++
	case p.hasPrefix("="):
		op = "="
		p.pos++
	default:
		return nil, fmt.Errorf("expected comparison operator at pos %d in %q", p.pos, p.input)
	}
	p.skipWS()
	val, err := p.parseLiteral()
	if err != nil {
		return nil, err
	}
	return compareTest{op, val}, nil
}

func (p *feelParser) parseInterval() (UnaryTest, error) {
	lowerIncl := p.input[p.pos] == '['
	p.pos++
	p.skipWS()
	lower, err := p.parseLiteral()
	if err != nil {
		return nil, fmt.Errorf("interval lower bound: %w", err)
	}
	p.skipWS()
	if !strings.HasPrefix(p.remaining(), "..") {
		return nil, fmt.Errorf("expected '..' in interval at pos %d in %q", p.pos, p.input)
	}
	p.pos += 2
	p.skipWS()
	upper, err := p.parseLiteral()
	if err != nil {
		return nil, fmt.Errorf("interval upper bound: %w", err)
	}
	p.skipWS()
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unclosed interval in %q", p.input)
	}
	upperIncl := p.input[p.pos] == ']'
	p.pos++
	return rangeTest{lower, upper, lowerIncl, upperIncl}, nil
}

func (p *feelParser) parseLiteral() (Value, error) {
	p.skipWS()
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unexpected end of expression")
	}
	ch := p.input[p.pos]

	// String literal "..."
	if ch == '"' {
		return p.parseStringLiteral()
	}

	// Boolean
	if p.hasPrefix("true") && (len(p.remaining()) == 4 || !isIdentChar(p.input[p.pos+4])) {
		p.pos += 4
		return BoolVal{true}, nil
	}
	if p.hasPrefix("false") && (len(p.remaining()) == 5 || !isIdentChar(p.input[p.pos+5])) {
		p.pos += 5
		return BoolVal{false}, nil
	}

	// date("...") — FEEL date constructor
	if p.hasPrefix(`date("`) {
		p.pos += 6
		s, err := p.readUntil('"')
		if err != nil {
			return nil, err
		}
		if p.pos < len(p.input) && p.input[p.pos] == '"' {
			p.pos++
		}
		if p.pos < len(p.input) && p.input[p.pos] == ')' {
			p.pos++
		}
		for _, layout := range []string{"2006-01-02", "01/02/2006"} {
			if d, err2 := time.Parse(layout, s); err2 == nil {
				return DateVal{d}, nil
			}
		}
		return nil, fmt.Errorf("invalid date %q", s)
	}

	// Number (with optional leading sign)
	if ch == '-' || ch == '+' || (ch >= '0' && ch <= '9') {
		return p.parseNumber()
	}

	return nil, fmt.Errorf("unexpected character %q at pos %d in %q", string(ch), p.pos, p.input)
}

func (p *feelParser) parseStringLiteral() (Value, error) {
	p.pos++ // consume opening "
	var sb strings.Builder
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == '\\' && p.pos+1 < len(p.input) {
			p.pos++
			switch p.input[p.pos] {
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			default:
				// Preserve unknown escapes verbatim — see lexer.go.
				sb.WriteByte('\\')
				sb.WriteByte(p.input[p.pos])
			}
		} else if ch == '"' {
			p.pos++ // consume closing "
			return StringVal{sb.String()}, nil
		} else {
			sb.WriteByte(ch)
		}
		p.pos++
	}
	return nil, fmt.Errorf("unterminated string literal in %q", p.input)
}

func (p *feelParser) parseNumber() (Value, error) {
	start := p.pos
	if p.input[p.pos] == '-' || p.input[p.pos] == '+' {
		p.pos++
	}
	for p.pos < len(p.input) && (p.input[p.pos] >= '0' && p.input[p.pos] <= '9') {
		p.pos++
	}
	if p.pos < len(p.input) && p.input[p.pos] == '.' {
		// Make sure this isn't the start of ".." (range separator)
		if p.pos+1 < len(p.input) && p.input[p.pos+1] == '.' {
			// Stop before ".."
		} else {
			p.pos++
			for p.pos < len(p.input) && (p.input[p.pos] >= '0' && p.input[p.pos] <= '9') {
				p.pos++
			}
		}
	}
	numStr := p.input[start:p.pos]
	if numStr == "" || numStr == "-" || numStr == "+" {
		return nil, fmt.Errorf("invalid number at pos %d in %q", start, p.input)
	}
	n, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number %q: %w", numStr, err)
	}
	return NumberVal{n}, nil
}

func (p *feelParser) readUntil(stop byte) (string, error) {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != stop {
		p.pos++
	}
	return p.input[start:p.pos], nil
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
