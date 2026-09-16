package elixir

import "github.com/seemyown/elixir/internal/ast"

// Expression is any typed SQL value that can appear in a query clause.
type Expression interface {
	exprNode() ast.Node
}

// ExprOf is implemented by Expr[T] and Column[T].
type ExprOf[T any] interface {
	asExpr() Expr[T]
}

// Expr is a typed SQL expression of Go type T.
type Expr[T any] struct {
	node ast.Node
}

func (e Expr[T]) exprNode() ast.Node { return e.node }
func (e Expr[T]) asExpr() Expr[T]    { return e }

// As aliases the expression in SELECT lists (e.g. COUNT(id) AS "total").
func (e Expr[T]) As(alias string) Expr[T] {
	return Expr[T]{node: ast.AliasNode{Expr: e.node, Alias: alias}}
}

// Eq builds an equality predicate against a literal value.
func (e Expr[T]) Eq(value T) Predicate {
	return binaryPred(e.node, OpEq, value)
}

// EqExpr builds an equality predicate against another expression.
func (e Expr[T]) EqExpr(other ExprOf[T]) Predicate {
	return binaryNodePred(e.node, OpEq, other.asExpr().node)
}

// Ne builds a not-equal predicate against a literal value.
func (e Expr[T]) Ne(value T) Predicate {
	return binaryPred(e.node, OpNe, value)
}

// NeExpr builds a not-equal predicate against another expression.
func (e Expr[T]) NeExpr(other ExprOf[T]) Predicate {
	return binaryNodePred(e.node, OpNe, other.asExpr().node)
}

// Gt builds a greater-than predicate against a literal value.
func (e Expr[T]) Gt(value T) Predicate {
	return binaryPred(e.node, OpGt, value)
}

// GtExpr builds a greater-than predicate against another expression.
func (e Expr[T]) GtExpr(other ExprOf[T]) Predicate {
	return binaryNodePred(e.node, OpGt, other.asExpr().node)
}

// Gte builds a greater-or-equal predicate against a literal value.
func (e Expr[T]) Gte(value T) Predicate {
	return binaryPred(e.node, OpGte, value)
}

// GteExpr builds a greater-or-equal predicate against another expression.
func (e Expr[T]) GteExpr(other ExprOf[T]) Predicate {
	return binaryNodePred(e.node, OpGte, other.asExpr().node)
}

// Lt builds a less-than predicate against a literal value.
func (e Expr[T]) Lt(value T) Predicate {
	return binaryPred(e.node, OpLt, value)
}

// LtExpr builds a less-than predicate against another expression.
func (e Expr[T]) LtExpr(other ExprOf[T]) Predicate {
	return binaryNodePred(e.node, OpLt, other.asExpr().node)
}

// Lte builds a less-or-equal predicate against a literal value.
func (e Expr[T]) Lte(value T) Predicate {
	return binaryPred(e.node, OpLte, value)
}

// LteExpr builds a less-or-equal predicate against another expression.
func (e Expr[T]) LteExpr(other ExprOf[T]) Predicate {
	return binaryNodePred(e.node, OpLte, other.asExpr().node)
}

// IsNull builds an IS NULL predicate.
func (e Expr[T]) IsNull() Predicate {
	return Predicate{node: ast.UnaryNode{Op: OpIsNull, Expr: e.node}}
}

// IsNotNull builds an IS NOT NULL predicate.
func (e Expr[T]) IsNotNull() Predicate {
	return Predicate{node: ast.UnaryNode{Op: OpIsNotNull, Expr: e.node}}
}

// Asc marks the expression for ascending ORDER BY.
func (e Expr[T]) Asc() OrderExpr {
	return OrderExpr{node: ast.OrderNode{Expr: e.node, Desc: false}}
}

// Desc marks the expression for descending ORDER BY.
func (e Expr[T]) Desc() OrderExpr {
	return OrderExpr{node: ast.OrderNode{Expr: e.node, Desc: true}}
}

func binaryPred[T any](left ast.Node, op BinaryOperator, value T) Predicate {
	return binaryNodePred(left, op, ast.LiteralNode{Value: value})
}

func binaryNodePred(left ast.Node, op BinaryOperator, right ast.Node) Predicate {
	return Predicate{
		node: ast.BinaryNode{
			Left:  left,
			Op:    op,
			Right: right,
		},
	}
}
