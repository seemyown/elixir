package elixir

import "github.com/seemyown/elixir/internal/ast"

// Assignment is a column = value pair used by UPDATE and named INSERT.
type Assignment struct {
	node ast.AssignNode
}

// Set builds a typed column = literal assignment.
func Set[T any](col Column[T], value T) Assignment {
	return Assignment{
		node: ast.AssignNode{
			Column: col.exprNode(),
			Value:  ast.LiteralNode{Value: value},
		},
	}
}

// SetExpr builds a typed column = expression assignment.
func SetExpr[T any](col Column[T], expr ExprOf[T]) Assignment {
	return Assignment{
		node: ast.AssignNode{
			Column: col.exprNode(),
			Value:  expr.asExpr().node,
		},
	}
}

// SetNull builds a typed column = NULL assignment.
func SetNull[T any](col Column[T]) Assignment {
	return Assignment{
		node: ast.AssignNode{
			Column: col.exprNode(),
			Value:  ast.NullNode{},
		},
	}
}
