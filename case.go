package elixir

import "github.com/seemyown/elixir/internal/ast"

// CaseBuilder builds a searched CASE WHEN ... THEN ... END expression.
type CaseBuilder[T any] struct {
	node ast.CaseNode
}

// Case starts a searched CASE expression of result type T.
func Case[T any]() CaseBuilder[T] {
	return CaseBuilder[T]{}
}

// When adds WHEN predicate THEN result.
func (c CaseBuilder[T]) When(pred Predicate, then ExprOf[T]) CaseBuilder[T] {
	c.node.Whens = append(c.node.Whens, ast.CaseWhenNode{
		When: pred.node,
		Then: then.asExpr().node,
	})
	return c
}

// Else sets ELSE result and finishes the CASE as an Expr.
func (c CaseBuilder[T]) Else(v ExprOf[T]) Expr[T] {
	c.node.Else = v.asExpr().node
	return c.End()
}

// End finishes the CASE without an ELSE clause.
func (c CaseBuilder[T]) End() Expr[T] {
	return Expr[T]{node: c.node}
}

// SimpleCaseBuilder builds CASE expr WHEN val THEN ... END.
type SimpleCaseBuilder[V, T any] struct {
	node ast.CaseNode
}

// CaseOn starts a simple CASE on the given expression.
func CaseOn[V, T any](expr ExprOf[V]) SimpleCaseBuilder[V, T] {
	return SimpleCaseBuilder[V, T]{
		node: ast.CaseNode{Value: expr.asExpr().node},
	}
}

// When adds WHEN match THEN result (match compared to the CASE operand).
func (c SimpleCaseBuilder[V, T]) When(match ExprOf[V], then ExprOf[T]) SimpleCaseBuilder[V, T] {
	c.node.Whens = append(c.node.Whens, ast.CaseWhenNode{
		When: match.asExpr().node,
		Then: then.asExpr().node,
	})
	return c
}

// Else sets ELSE result and finishes the CASE as an Expr.
func (c SimpleCaseBuilder[V, T]) Else(v ExprOf[T]) Expr[T] {
	c.node.Else = v.asExpr().node
	return c.End()
}

// End finishes the simple CASE without an ELSE clause.
func (c SimpleCaseBuilder[V, T]) End() Expr[T] {
	return Expr[T]{node: c.node}
}
