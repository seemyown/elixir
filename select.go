package elixir

import (
	"github.com/seemyown/elixir/internal/ast"
	"github.com/seemyown/elixir/internal/compiler"
)

// OrderExpr is an expression used in ORDER BY, optionally descending.
type OrderExpr struct {
	node ast.OrderNode
}

// SelectQuery is a fluent SELECT statement builder.
type SelectQuery struct {
	node    ast.SelectNode
	dialect Dialect
}

// Select starts a SELECT query with the given projection expressions.
func Select(cols ...Expression) SelectQuery {
	columns := make([]ast.Node, len(cols))
	for i, c := range cols {
		columns[i] = c.exprNode()
	}
	return SelectQuery{node: ast.SelectNode{Columns: columns}}
}

// WithDialect embeds a dialect so Compile() / ToSQL() can omit the dialect argument.
func (q SelectQuery) WithDialect(d Dialect) SelectQuery {
	q.dialect = d
	return q
}

// From sets the FROM table (or subquery via SelectQuery.AsTable).
func (q SelectQuery) From(table Table) SelectQuery {
	ref := table.tableRef()
	q.node.From = &ref
	return q
}

// Where appends predicates combined with AND.
func (q SelectQuery) Where(preds ...Predicate) SelectQuery {
	for _, p := range preds {
		q.node.Where = append(q.node.Where, p.node)
	}
	return q
}

// GroupBy appends GROUP BY expressions.
func (q SelectQuery) GroupBy(exprs ...Expression) SelectQuery {
	for _, e := range exprs {
		q.node.GroupBy = append(q.node.GroupBy, e.exprNode())
	}
	return q
}

// Having appends HAVING predicates combined with AND.
func (q SelectQuery) Having(preds ...Predicate) SelectQuery {
	for _, p := range preds {
		q.node.Having = append(q.node.Having, p.node)
	}
	return q
}

// OrderBy appends ORDER BY items. Use Expr.Asc() / Expr.Desc().
func (q SelectQuery) OrderBy(orders ...OrderExpr) SelectQuery {
	for _, o := range orders {
		q.node.OrderBy = append(q.node.OrderBy, o.node)
	}
	return q
}

// Limit sets LIMIT n.
func (q SelectQuery) Limit(n int) SelectQuery {
	q.node.Limit = &n
	return q
}

// Offset sets OFFSET n.
func (q SelectQuery) Offset(n int) SelectQuery {
	q.node.Offset = &n
	return q
}

// Distinct emits SELECT DISTINCT.
func (q SelectQuery) Distinct() SelectQuery {
	q.node.Distinct = true
	return q
}

// Union appends UNION other. ORDER BY / LIMIT / OFFSET apply to the compound query.
func (q SelectQuery) Union(other SelectQuery) SelectQuery {
	q.node.Unions = append(q.node.Unions, ast.UnionNode{Query: other.node})
	return q
}

// UnionAll appends UNION ALL other. ORDER BY / LIMIT / OFFSET apply to the compound query.
func (q SelectQuery) UnionAll(other SelectQuery) SelectQuery {
	q.node.Unions = append(q.node.Unions, ast.UnionNode{All: true, Query: other.node})
	return q
}

// Join adds an INNER JOIN.
func (q SelectQuery) Join(table Table, on Predicate) SelectQuery {
	return q.addJoin(ast.JoinInner, table, on)
}

// LeftJoin adds a LEFT JOIN.
func (q SelectQuery) LeftJoin(table Table, on Predicate) SelectQuery {
	return q.addJoin(ast.JoinLeft, table, on)
}

// RightJoin adds a RIGHT JOIN.
func (q SelectQuery) RightJoin(table Table, on Predicate) SelectQuery {
	return q.addJoin(ast.JoinRight, table, on)
}

// FullJoin adds a FULL JOIN.
func (q SelectQuery) FullJoin(table Table, on Predicate) SelectQuery {
	return q.addJoin(ast.JoinFull, table, on)
}

func (q SelectQuery) addJoin(kind ast.JoinKind, table Table, on Predicate) SelectQuery {
	q.node.Joins = append(q.node.Joins, ast.JoinNode{
		Kind:     kind,
		Relation: table.tableRef(),
		On:       on.node,
	})
	return q
}

// AsTable uses this SELECT as a FROM / JOIN subquery with the given alias.
func (q SelectQuery) AsTable(alias string) Table {
	return subqueryTable{query: q.node, alias: alias}
}

// Compile renders the query.
//
// With no arguments, uses the dialect embedded via Engine.Select / WithDialect.
// With one Dialect argument, that dialect is used (one-off or override).
func (q SelectQuery) Compile(ds ...Dialect) (string, []any, error) {
	d, err := resolveDialect(q.dialect, ds...)
	if err != nil {
		return "", nil, err
	}
	return compiler.CompileSelect(asCompilerDialect(d), q.node)
}

// ToSQL is an alias for Compile.
func (q SelectQuery) ToSQL(ds ...Dialect) (string, []any, error) {
	return q.Compile(ds...)
}
