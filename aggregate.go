package elixir

import "github.com/seemyown/elixir/internal/ast"

// Numeric constrains aggregate helpers that require numeric input.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Distinct marks an expression so aggregates can emit COUNT(DISTINCT ...).
func Distinct[T any](e ExprOf[T]) Expr[T] {
	return Expr[T]{node: ast.DistinctNode{Expr: e.asExpr().node}}
}

// Count builds COUNT(expr). When expr is Distinct(...), emits COUNT(DISTINCT ...).
func Count[T any](e ExprOf[T]) Expr[int64] {
	n := e.asExpr().node
	if d, ok := n.(ast.DistinctNode); ok {
		return Expr[int64]{node: ast.FunctionNode{Name: "COUNT", Args: []ast.Node{d.Expr}, Distinct: true}}
	}
	return Expr[int64]{node: ast.FunctionNode{Name: "COUNT", Args: []ast.Node{n}}}
}

// CountAll builds COUNT(*).
func CountAll() Expr[int64] {
	return Expr[int64]{node: ast.FunctionNode{Name: "COUNT", Args: []ast.Node{ast.StarNode{}}}}
}

// Sum builds SUM(expr).
func Sum[T Numeric](e ExprOf[T]) Expr[T] {
	return Expr[T]{node: ast.FunctionNode{Name: "SUM", Args: []ast.Node{e.asExpr().node}}}
}

// Avg builds AVG(expr) as float64.
func Avg[T Numeric](e ExprOf[T]) Expr[float64] {
	return Expr[float64]{node: ast.FunctionNode{Name: "AVG", Args: []ast.Node{e.asExpr().node}}}
}

// Min builds MIN(expr).
func Min[T any](e ExprOf[T]) Expr[T] {
	return Expr[T]{node: ast.FunctionNode{Name: "MIN", Args: []ast.Node{e.asExpr().node}}}
}

// Max builds MAX(expr).
func Max[T any](e ExprOf[T]) Expr[T] {
	return Expr[T]{node: ast.FunctionNode{Name: "MAX", Args: []ast.Node{e.asExpr().node}}}
}
