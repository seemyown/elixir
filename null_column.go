package elixir

import (
	"database/sql"

	"github.com/seemyown/elixir/internal/ast"
)

// NullColumn is a NULL-able table column.
//
// Use NullColumn for nullable fields. Non-null columns stay as Column[T]:
// passing Null[T] / SetNull / SetPtr to Column[T] is a Go compile-time error.
//
//	Bio: elixir.NullColumn[string]{Name: "bio"}
//
//	Users.Bio.SetPtr(nil)                 // NULL
//	Users.Bio.SetPtr(&s)                  // value
//	Users.Bio.SetOpt(elixir.Some("x"))
//	Users.Bio.SetOpt(elixir.None[string]())
//	Users.Bio.SetSQLNull(sql.Null[string]{…})
type NullColumn[T any] struct {
	Table      string
	Name       string
	Default    T
	HasDefault bool
}

func (NullColumn[T]) isColumn() {}

// WithDefault records a typed DB default (including a zero value of T).
func (c NullColumn[T]) WithDefault(v T) NullColumn[T] {
	c.Default = v
	c.HasDefault = true
	return c
}

func (c NullColumn[T]) asColumn() Column[T] {
	return Column[T]{Table: c.Table, Name: c.Name, Default: c.Default, HasDefault: c.HasDefault}
}

func (c NullColumn[T]) exprNode() ast.Node { return c.asColumn().exprNode() }
func (c NullColumn[T]) asExpr() Expr[T]    { return c.asColumn().asExpr() }

// Set assigns a non-NULL value.
func (c NullColumn[T]) Set(value T) Assignment { return Set(c.asColumn(), value) }

// SetExpr assigns an expression.
func (c NullColumn[T]) SetExpr(expr ExprOf[T]) Assignment {
	return SetExpr(c.asColumn(), expr)
}

// SetOpt assigns from elixir.Null[T] (NULL when !Valid).
func (c NullColumn[T]) SetOpt(n Null[T]) Assignment {
	return Assignment{
		node: ast.AssignNode{
			Column: c.exprNode(),
			Value:  nullToNode(n),
		},
	}
}

// SetPtr assigns from a pointer. nil → SQL NULL.
func (c NullColumn[T]) SetPtr(p *T) Assignment { return c.SetOpt(FromPtr(p)) }

// SetSQLNull assigns from database/sql.Null[T].
func (c NullColumn[T]) SetSQLNull(n sql.Null[T]) Assignment {
	return c.SetOpt(FromSQL(n))
}

// SetNull assigns SQL NULL.
func (c NullColumn[T]) SetNull() Assignment { return c.SetOpt(None[T]()) }

// As aliases the column in SELECT lists.
func (c NullColumn[T]) As(alias string) Expr[T] { return c.asExpr().As(alias) }

func (c NullColumn[T]) Eq(value T) Predicate              { return c.asExpr().Eq(value) }
func (c NullColumn[T]) EqExpr(other ExprOf[T]) Predicate  { return c.asExpr().EqExpr(other) }
func (c NullColumn[T]) Ne(value T) Predicate              { return c.asExpr().Ne(value) }
func (c NullColumn[T]) NeExpr(other ExprOf[T]) Predicate  { return c.asExpr().NeExpr(other) }
func (c NullColumn[T]) Gt(value T) Predicate              { return c.asExpr().Gt(value) }
func (c NullColumn[T]) GtExpr(other ExprOf[T]) Predicate  { return c.asExpr().GtExpr(other) }
func (c NullColumn[T]) Gte(value T) Predicate             { return c.asExpr().Gte(value) }
func (c NullColumn[T]) GteExpr(other ExprOf[T]) Predicate { return c.asExpr().GteExpr(other) }
func (c NullColumn[T]) Lt(value T) Predicate              { return c.asExpr().Lt(value) }
func (c NullColumn[T]) LtExpr(other ExprOf[T]) Predicate  { return c.asExpr().LtExpr(other) }
func (c NullColumn[T]) Lte(value T) Predicate             { return c.asExpr().Lte(value) }
func (c NullColumn[T]) LteExpr(other ExprOf[T]) Predicate { return c.asExpr().LteExpr(other) }
func (c NullColumn[T]) Like(pattern string) Predicate     { return c.asExpr().Like(pattern) }
func (c NullColumn[T]) LikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().LikeExpr(other)
}
func (c NullColumn[T]) NotLike(pattern string) Predicate { return c.asExpr().NotLike(pattern) }
func (c NullColumn[T]) NotLikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().NotLikeExpr(other)
}
func (c NullColumn[T]) ILike(pattern string) Predicate { return c.asExpr().ILike(pattern) }
func (c NullColumn[T]) ILikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().ILikeExpr(other)
}
func (c NullColumn[T]) NotILike(pattern string) Predicate { return c.asExpr().NotILike(pattern) }
func (c NullColumn[T]) NotILikeExpr(other ExprOf[string]) Predicate {
	return c.asExpr().NotILikeExpr(other)
}
func (c NullColumn[T]) Between(low, high T) Predicate { return c.asExpr().Between(low, high) }
func (c NullColumn[T]) BetweenExpr(low, high ExprOf[T]) Predicate {
	return c.asExpr().BetweenExpr(low, high)
}
func (c NullColumn[T]) NotBetween(low, high T) Predicate {
	return c.asExpr().NotBetween(low, high)
}
func (c NullColumn[T]) NotBetweenExpr(low, high ExprOf[T]) Predicate {
	return c.asExpr().NotBetweenExpr(low, high)
}
func (c NullColumn[T]) Add(value T) Expr[T]             { return c.asExpr().Add(value) }
func (c NullColumn[T]) AddExpr(other ExprOf[T]) Expr[T] { return c.asExpr().AddExpr(other) }
func (c NullColumn[T]) Sub(value T) Expr[T]             { return c.asExpr().Sub(value) }
func (c NullColumn[T]) SubExpr(other ExprOf[T]) Expr[T] { return c.asExpr().SubExpr(other) }
func (c NullColumn[T]) Mul(value T) Expr[T]             { return c.asExpr().Mul(value) }
func (c NullColumn[T]) MulExpr(other ExprOf[T]) Expr[T] { return c.asExpr().MulExpr(other) }
func (c NullColumn[T]) Div(value T) Expr[T]             { return c.asExpr().Div(value) }
func (c NullColumn[T]) DivExpr(other ExprOf[T]) Expr[T] { return c.asExpr().DivExpr(other) }
func (c NullColumn[T]) IsNull() Predicate               { return c.asExpr().IsNull() }
func (c NullColumn[T]) IsNotNull() Predicate            { return c.asExpr().IsNotNull() }
func (c NullColumn[T]) Asc() OrderExpr                  { return c.asExpr().Asc() }
func (c NullColumn[T]) Desc() OrderExpr                 { return c.asExpr().Desc() }
func (c NullColumn[T]) In(values ...T) Predicate        { return c.asExpr().In(values...) }
func (c NullColumn[T]) NotIn(values ...T) Predicate     { return c.asExpr().NotIn(values...) }
func (c NullColumn[T]) InQuery(q SelectQuery) Predicate { return c.asExpr().InQuery(q) }
func (c NullColumn[T]) NotInQuery(q SelectQuery) Predicate {
	return c.asExpr().NotInQuery(q)
}
func (c NullColumn[T]) Over(w Window) Expr[T] { return c.asExpr().Over(w) }

// EqOpt compares to elixir.Null[T]. Invalid → IS NULL; valid → = value.
func (c NullColumn[T]) EqOpt(n Null[T]) Predicate {
	if !n.Valid {
		return c.IsNull()
	}
	return c.Eq(n.V)
}

// NeOpt compares to elixir.Null[T]. Invalid → IS NOT NULL; valid → <> value.
func (c NullColumn[T]) NeOpt(n Null[T]) Predicate {
	if !n.Valid {
		return c.IsNotNull()
	}
	return c.Ne(n.V)
}

// EqPtr compares to a pointer. nil → IS NULL.
func (c NullColumn[T]) EqPtr(p *T) Predicate { return c.EqOpt(FromPtr(p)) }

// NePtr compares to a pointer. nil → IS NOT NULL.
func (c NullColumn[T]) NePtr(p *T) Predicate { return c.NeOpt(FromPtr(p)) }

// EqSQLNull compares to database/sql.Null[T].
func (c NullColumn[T]) EqSQLNull(n sql.Null[T]) Predicate {
	return c.EqOpt(FromSQL(n))
}

// ValueOpt wraps Null[T] as a VALUES / expression cell (NULL or literal).
func ValueOpt[T any](n Null[T]) Expr[T] {
	return Expr[T]{node: nullToNode(n)}
}

// ValuePtr wraps *T as a VALUES / expression cell. nil → NULL.
func ValuePtr[T any](p *T) Expr[T] {
	return ValueOpt(FromPtr(p))
}

// SetNull assigns SQL NULL to a NullColumn.
func SetNull[T any](col NullColumn[T]) Assignment { return col.SetNull() }

// SetOpt assigns elixir.Null[T] to a NullColumn.
func SetOpt[T any](col NullColumn[T], n Null[T]) Assignment { return col.SetOpt(n) }

// SetPtr assigns *T to a NullColumn. nil → NULL.
func SetPtr[T any](col NullColumn[T], p *T) Assignment { return col.SetPtr(p) }

// SetSQLNull assigns database/sql.Null[T] to a NullColumn.
func SetSQLNull[T any](col NullColumn[T], n sql.Null[T]) Assignment {
	return col.SetSQLNull(n)
}

// SetNullExpr assigns an expression to a NullColumn.
func SetNullExpr[T any](col NullColumn[T], expr ExprOf[T]) Assignment {
	return SetExpr(col.asColumn(), expr)
}
