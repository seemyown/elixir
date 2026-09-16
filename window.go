package elixir

import "github.com/seemyown/elixir/internal/ast"

// Window describes an OVER (...) specification for window functions.
type Window struct {
	partitionBy []ast.Node
	orderBy     []ast.OrderNode
	frame       *ast.FrameNode
}

// FrameBound is a window frame bound (UNBOUNDED PRECEDING, CURRENT ROW, …).
type FrameBound struct {
	node ast.FrameBoundNode
}

// UnboundedPreceding is UNBOUNDED PRECEDING.
func UnboundedPreceding() FrameBound {
	return FrameBound{node: ast.FrameBoundNode{Kind: ast.FrameUnboundedPreceding}}
}

// CurrentRow is CURRENT ROW.
func CurrentRow() FrameBound {
	return FrameBound{node: ast.FrameBoundNode{Kind: ast.FrameCurrentRow}}
}

// UnboundedFollowing is UNBOUNDED FOLLOWING.
func UnboundedFollowing() FrameBound {
	return FrameBound{node: ast.FrameBoundNode{Kind: ast.FrameUnboundedFollowing}}
}

// Preceding is n PRECEDING.
func Preceding(n int) FrameBound {
	return FrameBound{node: ast.FrameBoundNode{Kind: ast.FramePreceding, Offset: n}}
}

// Following is n FOLLOWING.
func Following(n int) FrameBound {
	return FrameBound{node: ast.FrameBoundNode{Kind: ast.FrameFollowing, Offset: n}}
}

// PartitionBy starts a window with PARTITION BY expressions.
func PartitionBy(exprs ...Expression) Window {
	return Window{}.PartitionBy(exprs...)
}

// WinOrderBy starts a window with only ORDER BY items.
func WinOrderBy(orders ...OrderExpr) Window {
	return Window{}.OrderBy(orders...)
}

// PartitionBy appends PARTITION BY expressions.
func (w Window) PartitionBy(exprs ...Expression) Window {
	for _, e := range exprs {
		w.partitionBy = append(w.partitionBy, e.exprNode())
	}
	return w
}

// OrderBy appends ORDER BY items inside OVER.
func (w Window) OrderBy(orders ...OrderExpr) Window {
	for _, o := range orders {
		w.orderBy = append(w.orderBy, o.node)
	}
	return w
}

// Rows sets a ROWS BETWEEN start AND end frame.
func (w Window) Rows(start, end FrameBound) Window {
	w.frame = &ast.FrameNode{Unit: ast.FrameRows, Start: start.node, End: end.node}
	return w
}

// Range sets a RANGE BETWEEN start AND end frame.
func (w Window) Range(start, end FrameBound) Window {
	w.frame = &ast.FrameNode{Unit: ast.FrameRange, Start: start.node, End: end.node}
	return w
}

// Over attaches an OVER (window) clause to an expression (typically an aggregate).
func (e Expr[T]) Over(w Window) Expr[T] {
	return Expr[T]{
		node: ast.WindowNode{
			Expr:        e.node,
			PartitionBy: w.partitionBy,
			OrderBy:     w.orderBy,
			Frame:       w.frame,
		},
	}
}

// RowNumber builds ROW_NUMBER().
func RowNumber() Expr[int64] {
	return Expr[int64]{node: ast.FunctionNode{Name: "ROW_NUMBER"}}
}

// Rank builds RANK().
func Rank() Expr[int64] {
	return Expr[int64]{node: ast.FunctionNode{Name: "RANK"}}
}

// DenseRank builds DENSE_RANK().
func DenseRank() Expr[int64] {
	return Expr[int64]{node: ast.FunctionNode{Name: "DENSE_RANK"}}
}
