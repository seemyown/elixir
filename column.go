package elixir

import "github.com/seemyown/elixir/internal/ast"

// Column is a typed table column expression.
//
// Define columns with struct literals. Name is required; Table is an optional
// qualifier used for joins and disambiguation (or filled by Bind):
//
//	Column[int64]{Name: "id"}
//	Column[int64]{Table: "users", Name: "id"}
//	Column[string]{Name: "bio", Nullable: true}
//	Column[bool]{Name: "active", Default: true}
type Column[T any] struct {
	// Table is an optional SQL qualifier (table name or alias).
	Table string
	// Name is the column name.
	Name string
	// Nullable marks the column as NULL-able (schema metadata).
	Nullable bool
	// Default, when non-nil, records a DB default. Used by Insert.WithDefaults
	// to emit an explicit DEFAULT for omitted columns.
	Default any
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

// SetNull builds a typed column = NULL assignment for INSERT/UPDATE.
func (c Column[T]) SetNull() Assignment {
	return SetNull(c)
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

// IsNull builds an IS NULL predicate.
func (c Column[T]) IsNull() Predicate {
	return c.asExpr().IsNull()
}

// IsNotNull builds an IS NOT NULL predicate.
func (c Column[T]) IsNotNull() Predicate {
	return c.asExpr().IsNotNull()
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
