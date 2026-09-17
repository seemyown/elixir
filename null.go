package elixir

import (
	"database/sql"
	"database/sql/driver"

	"github.com/seemyown/elixir/internal/ast"
)

// Null is an optional value for nullable columns.
// Prefer Null / *T over database/sql.Null[T] in application code;
// convert at the edges with FromSQL / SQL.
//
//	n := elixir.Some("hello")
//	p := n.Ptr()          // *string
//	n2 := elixir.FromPtr(p)
type Null[T any] struct {
	V     T
	Valid bool
}

// Some wraps a non-NULL value.
func Some[T any](v T) Null[T] {
	return Null[T]{V: v, Valid: true}
}

// None is a NULL value.
func None[T any]() Null[T] {
	return Null[T]{}
}

// FromPtr converts a pointer to Null. nil → NULL; non-nil → Some(*p).
func FromPtr[T any](p *T) Null[T] {
	if p == nil {
		return None[T]()
	}
	return Some(*p)
}

// Ptr returns a pointer to the value, or nil when Valid is false.
func (n Null[T]) Ptr() *T {
	if !n.Valid {
		return nil
	}
	v := n.V
	return &v
}

// Or returns V when Valid, otherwise fallback.
func (n Null[T]) Or(fallback T) T {
	if n.Valid {
		return n.V
	}
	return fallback
}

// FromSQL converts database/sql.Null[T] into elixir.Null[T].
func FromSQL[T any](n sql.Null[T]) Null[T] {
	return Null[T]{V: n.V, Valid: n.Valid}
}

// SQL converts to database/sql.Null[T] for drivers / scanners that expect it.
func (n Null[T]) SQL() sql.Null[T] {
	return sql.Null[T]{V: n.V, Valid: n.Valid}
}

// Scan implements database/sql.Scanner via sql.Null[T].
func (n *Null[T]) Scan(src any) error {
	var s sql.Null[T]
	if err := s.Scan(src); err != nil {
		return err
	}
	*n = FromSQL(s)
	return nil
}

// Value implements database/sql/driver.Valuer via sql.Null[T].
func (n Null[T]) Value() (driver.Value, error) {
	return n.SQL().Value()
}

func nullToNode[T any](n Null[T]) ast.Node {
	if !n.Valid {
		return ast.NullNode{}
	}
	return ast.LiteralNode{Value: n.V}
}
