package elixir

import "github.com/seemyown/elixir/internal/ast"

// Column is a typed non-NULL table column expression.
//
// For NULL-able columns use NullColumn[T]. Assigning Null[T] / calling SetNull
// on Column[T] is a Go compile-time error.
//
//	Column[int64]{Name: "id"}
//	Column[bool]{Name: "active"}.WithDefault(true)
type Column[T any] struct {
	// Table is an optional SQL qualifier (table name or alias).
	Table string
	// Name is the column name.
	Name string
	// Default records a DB default when HasDefault is true. Used by
	// Insert.WithDefaults to emit an explicit DEFAULT for omitted columns.
	// A zero T does not mean "no default" — set HasDefault / WithDefault.
	Default    T
	HasDefault bool
}

// WithDefault records a typed DB default (including a zero value of T).
func (c Column[T]) WithDefault(v T) Column[T] {
	c.Default = v
	c.HasDefault = true
	return c
}

// isColumn marks Column for Bind / As reflection (not used in query compile).
func (Column[T]) isColumn() {}

// Set builds a typed column = literal assignment for INSERT/UPDATE.
func (c Column[T]) Set(value T) Assignment {
	return Set(c, value)
}

// SetExpr builds a typed column = expression assignment for INSERT/UPDATE.
func (c Column[T]) SetExpr(expr ExprOf[T]) Assignment {
	return SetExpr(c, expr)
}

func (c Column[T]) exprNode() ast.Node {
	return ast.ColumnNode{Table: c.Table, Name: c.Name}
}

func (c Column[T]) asExpr() Expr[T] {
	return Expr[T]{node: c.exprNode()}
}

// As aliases the column in SELECT lists.
func (c Column[T]) As(alias string) Expr[T] {
	return c.asExpr().As(alias)
}

// Eq builds an equality predicate against a literal value.
func (c Column[T]) Eq(value T) Predicate {
	return c.asExpr().Eq(value)
}

// EqExpr builds an equality predicate against another expression.
func (c Column[T]) EqExpr(other ExprOf[T]) Predicate {
	return c.asExpr().EqExpr(other)
}

// Ne builds a not-equal predicate against a literal value.
func (c Column[T]) Ne(value T) Predicate {
	return c.asExpr().Ne(value)
}

// NeExpr builds a not-equal predicate against another expression.
func (c Column[T]) NeExpr(other ExprOf[T]) Predicate {
	return c.asExpr().NeExpr(other)
}

// Gt builds a greater-than predicate against a literal value.
func (c Column[T]) Gt(value T) Predicate {
	return c.asExpr().Gt(value)
}

// GtExpr builds a greater-than predicate against another expression.
func (c Column[T]) GtExpr(other ExprOf[T]) Predicate {
	return c.asExpr().GtExpr(other)
}

// Gte builds a greater-or-equal predicate against a literal value.
func (c Column[T]) Gte(value T) Predicate {
	return c.asExpr().Gte(value)
}

// GteExpr builds a greater-or-equal predicate against another expression.
func (c Column[T]) GteExpr(other ExprOf[T]) Predicate {
	return c.asExpr().GteExpr(other)
}

// Lt builds a less-than predicate against a literal value.
func (c Column[T]) Lt(value T) Predicate {
	return c.asExpr().Lt(value)
}

// LtExpr builds a less-than predicate against another expression.
func (c Column[T]) LtExpr(other ExprOf[T]) Predicate {
	return c.asExpr().LtExpr(other)
}

// Lte builds a less-or-equal predicate against a literal value.
func (c Column[T]) Lte(value T) Predicate {
	return c.asExpr().Lte(value)
}

// LteExpr builds a less-or-equal predicate against another expression.
func (c Column[T]) LteExpr(other ExprOf[T]) Predicate {
	return c.asExpr().LteExpr(other)
}

// Like builds column LIKE pattern.
func (c Column[T]) Like(pattern string) Predicate {
	return c.asExpr().Like(pattern)
}

// LikeExpr builds column LIKE other.
func (c Column[T]) LikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().LikeExpr(other)
}

// NotLike builds column NOT LIKE pattern.
func (c Column[T]) NotLike(pattern string) Predicate {
	return c.asExpr().NotLike(pattern)
}

// NotLikeExpr builds column NOT LIKE other.
func (c Column[T]) NotLikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().NotLikeExpr(other)
}

// ILike builds column ILIKE pattern (PostgreSQL only).
func (c Column[T]) ILike(pattern string) Predicate {
	return c.asExpr().ILike(pattern)
}

// ILikeExpr builds column ILIKE other (PostgreSQL only).
func (c Column[T]) ILikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().ILikeExpr(other)
}

// NotILike builds column NOT ILIKE pattern (PostgreSQL only).
func (c Column[T]) NotILike(pattern string) Predicate {
	return c.asExpr().NotILike(pattern)
}

// NotILikeExpr builds column NOT ILIKE other (PostgreSQL only).
func (c Column[T]) NotILikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().NotILikeExpr(other)
}

// Between builds column BETWEEN low AND high.
func (c Column[T]) Between(low, high T) Predicate {
	return c.asExpr().Between(low, high)
}

// BetweenExpr builds column BETWEEN low AND high using expressions.
func (c Column[T]) BetweenExpr(low, high ExprOf[T]) Predicate {
	return c.asExpr().BetweenExpr(low, high)
}

// NotBetween builds column NOT BETWEEN low AND high.
func (c Column[T]) NotBetween(low, high T) Predicate {
	return c.asExpr().NotBetween(low, high)
}

// NotBetweenExpr builds column NOT BETWEEN low AND high using expressions.
func (c Column[T]) NotBetweenExpr(low, high ExprOf[T]) Predicate {
	return c.asExpr().NotBetweenExpr(low, high)
}

// Add builds column + value.
func (c Column[T]) Add(value T) Expr[T] {
	return c.asExpr().Add(value)
}

// AddExpr builds column + other.
func (c Column[T]) AddExpr(other ExprOf[T]) Expr[T] {
	return c.asExpr().AddExpr(other)
}

// Sub builds column - value.
func (c Column[T]) Sub(value T) Expr[T] {
	return c.asExpr().Sub(value)
}

// SubExpr builds column - other.
func (c Column[T]) SubExpr(other ExprOf[T]) Expr[T] {
	return c.asExpr().SubExpr(other)
}

// Mul builds column * value.
func (c Column[T]) Mul(value T) Expr[T] {
	return c.asExpr().Mul(value)
}

// MulExpr builds column * other.
func (c Column[T]) MulExpr(other ExprOf[T]) Expr[T] {
	return c.asExpr().MulExpr(other)
}

// Div builds column / value.
func (c Column[T]) Div(value T) Expr[T] {
	return c.asExpr().Div(value)
}

// DivExpr builds column / other.
func (c Column[T]) DivExpr(other ExprOf[T]) Expr[T] {
	return c.asExpr().DivExpr(other)
}

// Asc marks the column for ascending ORDER BY.
func (c Column[T]) Asc() OrderExpr {
	return c.asExpr().Asc()
}

// Desc marks the column for descending ORDER BY.
func (c Column[T]) Desc() OrderExpr {
	return c.asExpr().Desc()
}

// In builds column IN (v1, v2, ...).
func (c Column[T]) In(values ...T) Predicate {
	return c.asExpr().In(values...)
}

// NotIn builds column NOT IN (v1, v2, ...).
func (c Column[T]) NotIn(values ...T) Predicate {
	return c.asExpr().NotIn(values...)
}

// InQuery builds column IN (SELECT ...).
func (c Column[T]) InQuery(q SelectQuery) Predicate {
	return c.asExpr().InQuery(q)
}

// NotInQuery builds column NOT IN (SELECT ...).
func (c Column[T]) NotInQuery(q SelectQuery) Predicate {
	return c.asExpr().NotInQuery(q)
}

// Over attaches an OVER (window) clause (typically unused on bare columns).
func (c Column[T]) Over(w Window) Expr[T] {
	return c.asExpr().Over(w)
}
