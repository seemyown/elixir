package elixir

import "github.com/seemyown/elixir/internal/ast"

// Assignment is a column = value pair used by UPDATE and named INSERT.
type Assignment struct {
	node ast.AssignNode
}

// Set builds a typed column = literal assignment for a non-NULL Column[T].
func Set[T any](col Column[T], value T) Assignment {
	return Assignment{
		node: ast.AssignNode{
			Column: col.exprNode(),
			Value:  ast.LiteralNode{Value: value},
		},
	}
}

// SetExpr builds a typed column = expression assignment for Column[T].
func SetExpr[T any](col Column[T], expr ExprOf[T]) Assignment {
	return Assignment{
		node: ast.AssignNode{
			Column: col.exprNode(),
			Value:  expr.asExpr().node,
		},
	}
}
