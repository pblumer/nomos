// Package feel implements a FEEL expression evaluator covering the
// "Common Subset" defined in ADR-0019 (Stufe 2):
// arithmetic, comparison, logic, path navigation, context/list literals,
// if/then/else and a curated catalogue of built-in functions.
//
// Out of scope (Stufe 3 in ADR-0019): boxed function definitions,
// BKM invocation, decision-service invocation, for/every/some/filter
// comprehensions over Relations, the full ~80-function built-in
// catalogue. Those will be added on demand.
package feel

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tkEOF tokenKind = iota
	tkNumber
	tkString
	tkIdent
	tkLParen
	tkRParen
	tkLBracket
	tkRBracket
	tkLBrace
	tkRBrace
	tkComma
	tkColon
	tkDot
	tkPlus
	tkMinus
	tkStar
	tkSlash
	tkPower
	tkEq
	tkNeq
	tkLt
	tkLte
	tkGt
	tkGte
	tkKwAnd
	tkKwOr
	tkKwNot
	tkKwTrue
	tkKwFalse
	tkKwNull
	tkKwIf
	tkKwThen
	tkKwElse
)

type token struct {
	kind tokenKind
	text string
	pos  int
}

func (t token) String() string {
	return fmt.Sprintf("%d(%q)", t.kind, t.text)
}

type lexError struct {
	pos int
	msg string
}

func (e *lexError) Error() string { return fmt.Sprintf("lex error at %d: %s", e.pos, e.msg) }

func tokenize(src string) ([]token, error) {
	var out []token
	i := 0
	for i < len(src) {
		c := src[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		start := i
		switch {
		case c == '(':
			out = append(out, token{tkLParen, "(", start})
			i++
		case c == ')':
			out = append(out, token{tkRParen, ")", start})
			i++
		case c == '[':
			out = append(out, token{tkLBracket, "[", start})
			i++
		case c == ']':
			out = append(out, token{tkRBracket, "]", start})
			i++
		case c == '{':
			out = append(out, token{tkLBrace, "{", start})
			i++
		case c == '}':
			out = append(out, token{tkRBrace, "}", start})
			i++
		case c == ',':
			out = append(out, token{tkComma, ",", start})
			i++
		case c == ':':
			out = append(out, token{tkColon, ":", start})
			i++
		case c == '.':
			out = append(out, token{tkDot, ".", start})
			i++
		case c == '+':
			out = append(out, token{tkPlus, "+", start})
			i++
		case c == '-':
			out = append(out, token{tkMinus, "-", start})
			i++
		case c == '*':
			if i+1 < len(src) && src[i+1] == '*' {
				out = append(out, token{tkPower, "**", start})
				i += 2
			} else {
				out = append(out, token{tkStar, "*", start})
				i++
			}
		case c == '/':
			out = append(out, token{tkSlash, "/", start})
			i++
		case c == '=':
			out = append(out, token{tkEq, "=", start})
			i++
		case c == '!':
			if i+1 < len(src) && src[i+1] == '=' {
				out = append(out, token{tkNeq, "!=", start})
				i += 2
			} else {
				return nil, &lexError{start, "unexpected '!' (use '!=')"}
			}
		case c == '<':
			if i+1 < len(src) && src[i+1] == '=' {
				out = append(out, token{tkLte, "<=", start})
				i += 2
			} else {
				out = append(out, token{tkLt, "<", start})
				i++
			}
		case c == '>':
			if i+1 < len(src) && src[i+1] == '=' {
				out = append(out, token{tkGte, ">=", start})
				i += 2
			} else {
				out = append(out, token{tkGt, ">", start})
				i++
			}
		case c == '"':
			j := i + 1
			var sb strings.Builder
			for j < len(src) && src[j] != '"' {
				if src[j] == '\\' && j+1 < len(src) {
					switch src[j+1] {
					case 'n':
						sb.WriteByte('\n')
					case 't':
						sb.WriteByte('\t')
					case 'r':
						sb.WriteByte('\r')
					case '"':
						sb.WriteByte('"')
					case '\\':
						sb.WriteByte('\\')
					default:
						sb.WriteByte(src[j+1])
					}
					j += 2
					continue
				}
				sb.WriteByte(src[j])
				j++
			}
			if j >= len(src) {
				return nil, &lexError{start, "unterminated string"}
			}
			out = append(out, token{tkString, sb.String(), start})
			i = j + 1
		case isDigit(c):
			j := i
			for j < len(src) && (isDigit(src[j]) || src[j] == '.') {
				j++
			}
			out = append(out, token{tkNumber, src[i:j], start})
			i = j
		case isIdentStart(c):
			j := i
			for j < len(src) && isIdentPart(src[j]) {
				j++
			}
			text := src[i:j]
			kind := tkIdent
			switch text {
			case "and":
				kind = tkKwAnd
			case "or":
				kind = tkKwOr
			case "not":
				kind = tkKwNot
			case "true":
				kind = tkKwTrue
			case "false":
				kind = tkKwFalse
			case "null":
				kind = tkKwNull
			case "if":
				kind = tkKwIf
			case "then":
				kind = tkKwThen
			case "else":
				kind = tkKwElse
			}
			out = append(out, token{kind, text, start})
			i = j
		default:
			return nil, &lexError{start, fmt.Sprintf("unexpected character %q", string(c))}
		}
	}
	out = append(out, token{tkEOF, "", len(src)})
	return out, nil
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isIdentStart(c byte) bool {
	return unicode.IsLetter(rune(c)) || c == '_'
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || isDigit(c)
}
