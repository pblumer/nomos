package feel

import (
	"fmt"
	"strconv"
)

// Precedence levels (Pratt). Higher value binds tighter.
const (
	precNone     = 0
	precOr       = 10
	precAnd      = 20
	precCompare  = 30
	precAdditive = 40
	precMul      = 50
	precPower    = 60
	precUnary    = 70
	precCall     = 80
)

type parser struct {
	toks []token
	i    int
}

type parseError struct {
	pos int
	msg string
}

func (e *parseError) Error() string { return fmt.Sprintf("parse error at %d: %s", e.pos, e.msg) }

func parse(src string) (Expr, error) {
	toks, err := tokenize(src)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	expr, err := p.parseExpr(precNone)
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tkEOF {
		return nil, &parseError{p.peek().pos, fmt.Sprintf("unexpected token %q", p.peek().text)}
	}
	return expr, nil
}

func (p *parser) peek() token    { return p.toks[p.i] }
func (p *parser) advance() token { t := p.toks[p.i]; p.i++; return t }

func (p *parser) expect(k tokenKind, what string) (token, error) {
	t := p.advance()
	if t.kind != k {
		return t, &parseError{t.pos, fmt.Sprintf("expected %s, got %q", what, t.text)}
	}
	return t, nil
}

func infixPrec(k tokenKind) int {
	switch k {
	case tkKwOr:
		return precOr
	case tkKwAnd:
		return precAnd
	case tkEq, tkNeq, tkLt, tkLte, tkGt, tkGte:
		return precCompare
	case tkPlus, tkMinus:
		return precAdditive
	case tkStar, tkSlash:
		return precMul
	case tkPower:
		return precPower
	case tkLParen, tkDot, tkLBracket:
		return precCall
	}
	return precNone
}

func (p *parser) parseExpr(minPrec int) (Expr, error) {
	lhs, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		prec := infixPrec(t.kind)
		if prec == precNone || prec < minPrec {
			return lhs, nil
		}
		p.advance()
		switch t.kind {
		case tkLParen:
			args, err := p.parseArgList()
			if err != nil {
				return nil, err
			}
			lhs = callExpr{callee: lhs, args: args}
		case tkDot:
			id, err := p.expect(tkIdent, "identifier")
			if err != nil {
				return nil, err
			}
			lhs = pathExpr{target: lhs, name: id.text}
		case tkLBracket:
			idx, err := p.parseExpr(precNone)
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(tkRBracket, "']'"); err != nil {
				return nil, err
			}
			lhs = callExpr{callee: nameRef{name: "__index"}, args: []Expr{lhs, idx}}
		default:
			// Left-associative binary; right-associative power.
			nextPrec := prec + 1
			if t.kind == tkPower {
				nextPrec = prec
			}
			rhs, err := p.parseExpr(nextPrec)
			if err != nil {
				return nil, err
			}
			lhs = binaryExpr{op: t.text, lhs: lhs, rhs: rhs}
		}
	}
}

func (p *parser) parsePrefix() (Expr, error) {
	t := p.advance()
	switch t.kind {
	case tkNumber:
		n, err := strconv.ParseFloat(t.text, 64)
		if err != nil {
			return nil, &parseError{t.pos, fmt.Sprintf("invalid number %q", t.text)}
		}
		return numberLit{v: n}, nil
	case tkString:
		return stringLit{v: t.text}, nil
	case tkKwTrue:
		return boolLit{v: true}, nil
	case tkKwFalse:
		return boolLit{v: false}, nil
	case tkKwNull:
		return nullLit{}, nil
	case tkIdent:
		return nameRef{name: t.text}, nil
	case tkMinus:
		rhs, err := p.parseExpr(precUnary)
		if err != nil {
			return nil, err
		}
		return unaryExpr{op: "-", rhs: rhs}, nil
	case tkKwNot:
		rhs, err := p.parseExpr(precUnary)
		if err != nil {
			return nil, err
		}
		return unaryExpr{op: "not", rhs: rhs}, nil
	case tkLParen:
		expr, err := p.parseExpr(precNone)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tkRParen, "')'"); err != nil {
			return nil, err
		}
		return expr, nil
	case tkLBracket:
		var items []Expr
		if p.peek().kind != tkRBracket {
			for {
				it, err := p.parseExpr(precNone)
				if err != nil {
					return nil, err
				}
				items = append(items, it)
				if p.peek().kind != tkComma {
					break
				}
				p.advance()
			}
		}
		if _, err := p.expect(tkRBracket, "']'"); err != nil {
			return nil, err
		}
		return listExpr{items: items}, nil
	case tkLBrace:
		var entries []contextEntry
		if p.peek().kind != tkRBrace {
			for {
				key, err := p.parseContextKey()
				if err != nil {
					return nil, err
				}
				if _, err := p.expect(tkColon, "':'"); err != nil {
					return nil, err
				}
				val, err := p.parseExpr(precNone)
				if err != nil {
					return nil, err
				}
				entries = append(entries, contextEntry{key: key, value: val})
				if p.peek().kind != tkComma {
					break
				}
				p.advance()
			}
		}
		if _, err := p.expect(tkRBrace, "'}'"); err != nil {
			return nil, err
		}
		return contextExpr{entries: entries}, nil
	case tkKwIf:
		cond, err := p.parseExpr(precNone)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tkKwThen, "'then'"); err != nil {
			return nil, err
		}
		then, err := p.parseExpr(precNone)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(tkKwElse, "'else'"); err != nil {
			return nil, err
		}
		els, err := p.parseExpr(precNone)
		if err != nil {
			return nil, err
		}
		return ifExpr{cond: cond, then: then, els: els}, nil
	}
	return nil, &parseError{t.pos, fmt.Sprintf("unexpected token %q", t.text)}
}

func (p *parser) parseContextKey() (string, error) {
	t := p.advance()
	switch t.kind {
	case tkIdent:
		return t.text, nil
	case tkString:
		return t.text, nil
	}
	return "", &parseError{t.pos, fmt.Sprintf("expected context key, got %q", t.text)}
}

func (p *parser) parseArgList() ([]Expr, error) {
	var args []Expr
	if p.peek().kind == tkRParen {
		p.advance()
		return args, nil
	}
	for {
		a, err := p.parseExpr(precNone)
		if err != nil {
			return nil, err
		}
		args = append(args, a)
		if p.peek().kind != tkComma {
			break
		}
		p.advance()
	}
	if _, err := p.expect(tkRParen, "')'"); err != nil {
		return nil, err
	}
	return args, nil
}
