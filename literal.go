package elixir

import "github.com/seemyown/elixir/internal/ast"

// Value wraps a Go value as a typed SQL literal expression.
func Value[T any](value T) Expr[T] {
	return Expr[T]{node: ast.LiteralNode{Value: value}}
}
