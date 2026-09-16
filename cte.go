package elixir

import "github.com/seemyown/elixir/internal/ast"

// CTEDef is a named common table expression usable with With(...).
type CTEDef struct {
	name  string
	query ast.SelectNode
}

// CTE starts a CTE named name from a SELECT query.
func CTE(name string, q SelectQuery) CTEDef {
	return CTEDef{name: name, query: q.node}
}

// WithBuilder attaches WITH clauses to a following statement.
type WithBuilder struct {
	ctes    []ast.CTENode
	dialect Dialect
}

// With starts a WITH clause containing one or more CTEs.
func With(ctes ...CTEDef) WithBuilder {
	nodes := make([]ast.CTENode, len(ctes))
	for i, c := range ctes {
		nodes[i] = ast.CTENode{Name: c.name, Query: c.query}
	}
	return WithBuilder{ctes: nodes}
}

// WithDialect embeds a dialect so Compile() on the following statement can omit it.
func (w WithBuilder) WithDialect(d Dialect) WithBuilder {
	w.dialect = d
	return w
}

// Select starts a SELECT that uses the WITH clause.
func (w WithBuilder) Select(cols ...Expression) SelectQuery {
	q := Select(cols...)
	q.node.With = append([]ast.CTENode(nil), w.ctes...)
	q.dialect = w.dialect
	return q
}

// Insert starts an INSERT that uses the WITH clause.
func (w WithBuilder) Insert(table Table) InsertQuery {
	q := Insert(table)
	q.node.With = append([]ast.CTENode(nil), w.ctes...)
	q.dialect = w.dialect
	return q
}

// Update starts an UPDATE that uses the WITH clause.
func (w WithBuilder) Update(table Table) UpdateQuery {
	q := Update(table)
	q.node.With = append([]ast.CTENode(nil), w.ctes...)
	q.dialect = w.dialect
	return q
}

// Delete starts a DELETE that uses the WITH clause.
func (w WithBuilder) Delete(table Table) DeleteQuery {
	q := Delete(table)
	q.node.With = append([]ast.CTENode(nil), w.ctes...)
	q.dialect = w.dialect
	return q
}
