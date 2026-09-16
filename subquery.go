package elixir

import "github.com/seemyown/elixir/internal/ast"

// Subquery wraps a SELECT as a scalar expression (e.g. for comparisons).
func Subquery[T any](q SelectQuery) Expr[T] {
	return Expr[T]{node: ast.SubqueryExprNode{Query: q.node}}
}

// Exists builds EXISTS (SELECT ...).
func Exists(q SelectQuery) Predicate {
	return Predicate{node: ast.ExistsNode{Query: q.node}}
}

// NotExists builds NOT EXISTS (SELECT ...).
func NotExists(q SelectQuery) Predicate {
	return Predicate{node: ast.ExistsNode{Not: true, Query: q.node}}
}

// In builds expr IN (v1, v2, ...).
func (e Expr[T]) In(values ...T) Predicate {
	list := make([]ast.Node, len(values))
	for i, v := range values {
		list[i] = ast.LiteralNode{Value: v}
	}
	return Predicate{node: ast.InNode{Expr: e.node, List: list}}
}

// NotIn builds expr NOT IN (v1, v2, ...).
func (e Expr[T]) NotIn(values ...T) Predicate {
	list := make([]ast.Node, len(values))
	for i, v := range values {
		list[i] = ast.LiteralNode{Value: v}
	}
	return Predicate{node: ast.InNode{Expr: e.node, Not: true, List: list}}
}

// InQuery builds expr IN (SELECT ...).
func (e Expr[T]) InQuery(q SelectQuery) Predicate {
	query := q.node
	return Predicate{node: ast.InNode{Expr: e.node, Query: &query}}
}

// NotInQuery builds expr NOT IN (SELECT ...).
func (e Expr[T]) NotInQuery(q SelectQuery) Predicate {
	query := q.node
	return Predicate{node: ast.InNode{Expr: e.node, Not: true, Query: &query}}
}
