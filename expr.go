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

// Like builds expr LIKE pattern.
func (e Expr[T]) Like(pattern string) Predicate {
	return binaryNodePred(e.node, OpLike, ast.LiteralNode{Value: pattern})
}

// LikeExpr builds expr LIKE other.
func (e Expr[T]) LikeExpr(other ExprOf[string]) Predicate {
	return binaryNodePred(e.node, OpLike, other.asExpr().node)
}

// NotLike builds expr NOT LIKE pattern.
func (e Expr[T]) NotLike(pattern string) Predicate {
	return binaryNodePred(e.node, OpNotLike, ast.LiteralNode{Value: pattern})
}

// NotLikeExpr builds expr NOT LIKE other.
func (e Expr[T]) NotLikeExpr(other ExprOf[string]) Predicate {
	return binaryNodePred(e.node, OpNotLike, other.asExpr().node)
}

// ILike builds expr ILIKE pattern (PostgreSQL only; Compile errors on other dialects).
func (e Expr[T]) ILike(pattern string) Predicate {
	return binaryNodePred(e.node, OpILike, ast.LiteralNode{Value: pattern})
}

// ILikeExpr builds expr ILIKE other (PostgreSQL only).
func (e Expr[T]) ILikeExpr(other ExprOf[string]) Predicate {
	return binaryNodePred(e.node, OpILike, other.asExpr().node)
}

// NotILike builds expr NOT ILIKE pattern (PostgreSQL only).
func (e Expr[T]) NotILike(pattern string) Predicate {
	return binaryNodePred(e.node, OpNotILike, ast.LiteralNode{Value: pattern})
}

// NotILikeExpr builds expr NOT ILIKE other (PostgreSQL only).
func (e Expr[T]) NotILikeExpr(other ExprOf[string]) Predicate {
	return binaryNodePred(e.node, OpNotILike, other.asExpr().node)
}

// Between builds expr BETWEEN low AND high.
func (e Expr[T]) Between(low, high T) Predicate {
	return betweenPred(e.node, false, ast.LiteralNode{Value: low}, ast.LiteralNode{Value: high})
}

// BetweenExpr builds expr BETWEEN low AND high using expressions.
func (e Expr[T]) BetweenExpr(low, high ExprOf[T]) Predicate {
	return betweenPred(e.node, false, low.asExpr().node, high.asExpr().node)
}

// NotBetween builds expr NOT BETWEEN low AND high.
func (e Expr[T]) NotBetween(low, high T) Predicate {
	return betweenPred(e.node, true, ast.LiteralNode{Value: low}, ast.LiteralNode{Value: high})
}

// NotBetweenExpr builds expr NOT BETWEEN low AND high using expressions.
func (e Expr[T]) NotBetweenExpr(low, high ExprOf[T]) Predicate {
	return betweenPred(e.node, true, low.asExpr().node, high.asExpr().node)
}

// Add builds expr + value.
func (e Expr[T]) Add(value T) Expr[T] {
	return binaryArith[T](e.node, OpAdd, ast.LiteralNode{Value: value})
}

// AddExpr builds expr + other.
func (e Expr[T]) AddExpr(other ExprOf[T]) Expr[T] {
	return binaryArith[T](e.node, OpAdd, other.asExpr().node)
}

// Sub builds expr - value.
func (e Expr[T]) Sub(value T) Expr[T] {
	return binaryArith[T](e.node, OpSub, ast.LiteralNode{Value: value})
}

// SubExpr builds expr - other.
func (e Expr[T]) SubExpr(other ExprOf[T]) Expr[T] {
	return binaryArith[T](e.node, OpSub, other.asExpr().node)
}

// Mul builds expr * value.
func (e Expr[T]) Mul(value T) Expr[T] {
	return binaryArith[T](e.node, OpMul, ast.LiteralNode{Value: value})
}

// MulExpr builds expr * other.
func (e Expr[T]) MulExpr(other ExprOf[T]) Expr[T] {
	return binaryArith[T](e.node, OpMul, other.asExpr().node)
}

// Div builds expr / value.
func (e Expr[T]) Div(value T) Expr[T] {
	return binaryArith[T](e.node, OpDiv, ast.LiteralNode{Value: value})
}

// DivExpr builds expr / other.
func (e Expr[T]) DivExpr(other ExprOf[T]) Expr[T] {
	return binaryArith[T](e.node, OpDiv, other.asExpr().node)
}

// Cast converts an expression to another Go type, emitting CAST(x AS sqlType).
// sqlType is raw SQL (e.g. "integer", "DOUBLE PRECISION") and is not quoted.
func Cast[To any](e Expression, sqlType string) Expr[To] {
	return Expr[To]{node: ast.CastNode{Expr: e.exprNode(), Type: sqlType}}
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

func betweenPred(expr ast.Node, not bool, low, high ast.Node) Predicate {
	return Predicate{
		node: ast.BetweenNode{Expr: expr, Low: low, High: high, Not: not},
	}
}

func binaryArith[T any](left ast.Node, op BinaryOperator, right ast.Node) Expr[T] {
	return Expr[T]{
		node: ast.BinaryNode{Left: left, Op: op, Right: right},
	}
}
