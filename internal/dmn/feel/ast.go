package feel

// Expr is a parsed FEEL expression node. Implementations are unexported
// because evaluation goes through (*Expression).Eval — callers do not
// need to walk the AST.
type Expr interface {
	exprNode()
}

type numberLit struct{ v float64 }
type stringLit struct{ v string }
type boolLit struct{ v bool }
type nullLit struct{}
type nameRef struct{ name string }

type pathExpr struct {
	target Expr
	name   string
}

type binaryExpr struct {
	op       string
	lhs, rhs Expr
}

type unaryExpr struct {
	op  string
	rhs Expr
}

type ifExpr struct {
	cond, then, els Expr
}

type listExpr struct{ items []Expr }

type contextEntry struct {
	key   string
	value Expr
}
type contextExpr struct{ entries []contextEntry }

type callExpr struct {
	callee Expr
	args   []Expr
}

func (numberLit) exprNode()   {}
func (stringLit) exprNode()   {}
func (boolLit) exprNode()     {}
func (nullLit) exprNode()     {}
func (nameRef) exprNode()     {}
func (pathExpr) exprNode()    {}
func (binaryExpr) exprNode()  {}
func (unaryExpr) exprNode()   {}
func (ifExpr) exprNode()      {}
func (listExpr) exprNode()    {}
func (contextExpr) exprNode() {}
func (callExpr) exprNode()    {}
