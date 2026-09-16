package elixir

import "github.com/seemyown/elixir/internal/ast"

// Lower wraps an expression in LOWER(...).
func Lower(e ExprOf[string]) Expr[string] {
	return Expr[string]{node: ast.FunctionNode{Name: "LOWER", Args: []ast.Node{e.asExpr().node}}}
}

// Upper wraps an expression in UPPER(...).
func Upper(e ExprOf[string]) Expr[string] {
	return Expr[string]{node: ast.FunctionNode{Name: "UPPER", Args: []ast.Node{e.asExpr().node}}}
}

// Coalesce builds COALESCE(arg0, arg1, ...).
func Coalesce[T any](first ExprOf[T], rest ...ExprOf[T]) Expr[T] {
	args := make([]ast.Node, 0, 1+len(rest))
	args = append(args, first.asExpr().node)
	for _, e := range rest {
		args = append(args, e.asExpr().node)
	}
	return Expr[T]{node: ast.FunctionNode{Name: "COALESCE", Args: args}}
}

// Func builds an arbitrary typed SQL function call.
func Func[T any](name string, args ...Expression) Expr[T] {
	nodes := make([]ast.Node, len(args))
	for i, a := range args {
		nodes[i] = a.exprNode()
	}
	return Expr[T]{node: ast.FunctionNode{Name: name, Args: nodes}}
}
