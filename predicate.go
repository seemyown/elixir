package elixir

import (
	"github.com/seemyown/elixir/internal/ast"
	"github.com/seemyown/elixir/internal/compiler"
)

// Predicate is a boolean SQL expression used in WHERE / HAVING / JOIN ON.
type Predicate struct {
	node ast.Node
}

func (p Predicate) exprNode() ast.Node { return p.node }

// CompilePredicate renders a standalone predicate (useful for tests and debugging).
func CompilePredicate(d Dialect, p Predicate) (string, []any, error) {
	return compiler.CompileExpr(asCompilerDialect(d), p.node)
}

// And combines predicates with AND. A single predicate is returned unchanged.
// Zero predicates yield a no-op TRUE literal.
func And(preds ...Predicate) Predicate {
	switch len(preds) {
	case 0:
		return Predicate{node: ast.LiteralNode{Value: true}}
	case 1:
		return preds[0]
	default:
		nodes := make([]ast.Node, len(preds))
		for i, p := range preds {
			nodes[i] = p.node
		}
		return Predicate{node: ast.LogicalNode{Op: OpAnd, Exprs: nodes}}
	}
}

// Or combines predicates with OR. A single predicate is returned unchanged.
// Zero predicates yield a no-op FALSE literal.
func Or(preds ...Predicate) Predicate {
	switch len(preds) {
	case 0:
		return Predicate{node: ast.LiteralNode{Value: false}}
	case 1:
		return preds[0]
	default:
		nodes := make([]ast.Node, len(preds))
		for i, p := range preds {
			nodes[i] = p.node
		}
		return Predicate{node: ast.LogicalNode{Op: OpOr, Exprs: nodes}}
	}
}

// Not negates a predicate.
func Not(pred Predicate) Predicate {
	return Predicate{node: ast.UnaryNode{Op: OpNot, Expr: pred.node}}
}
