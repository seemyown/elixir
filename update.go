package elixir

import (
	"github.com/seemyown/elixir/internal/ast"
	"github.com/seemyown/elixir/internal/compiler"
)

// UpdateQuery is a fluent UPDATE statement builder.
type UpdateQuery struct {
	node    ast.UpdateNode
	dialect Dialect
}

// Update starts an UPDATE table statement.
func Update(table Table) UpdateQuery {
	return UpdateQuery{node: ast.UpdateNode{Table: table.tableRef()}}
}

// WithDialect embeds a dialect so Compile() / ToSQL() can omit the dialect argument.
func (q UpdateQuery) WithDialect(d Dialect) UpdateQuery {
	q.dialect = d
	return q
}

// Set appends SET assignments. Use Set / SetExpr / SetNull helpers or Column.Set.
func (q UpdateQuery) Set(assigns ...Assignment) UpdateQuery {
	for _, a := range assigns {
		q.node.Sets = append(q.node.Sets, a.node)
	}
	return q
}

// Where appends predicates combined with AND.
func (q UpdateQuery) Where(preds ...Predicate) UpdateQuery {
	for _, p := range preds {
		q.node.Where = append(q.node.Where, p.node)
	}
	return q
}

// Returning appends a RETURNING clause (Postgres / SQLite; MySQL support varies).
func (q UpdateQuery) Returning(cols ...Expression) UpdateQuery {
	for _, c := range cols {
		q.node.Returning = append(q.node.Returning, c.exprNode())
	}
	return q
}

// Compile renders the query.
//
// With no arguments, uses the dialect embedded via Engine.Update / WithDialect.
// With one Dialect argument, that dialect is used (one-off or override).
func (q UpdateQuery) Compile(ds ...Dialect) (string, []any, error) {
	d, err := resolveDialect(q.dialect, ds...)
	if err != nil {
		return "", nil, err
	}
	return compiler.CompileUpdate(asCompilerDialect(d), q.node)
}

// ToSQL is an alias for Compile.
func (q UpdateQuery) ToSQL(ds ...Dialect) (string, []any, error) {
	return q.Compile(ds...)
}
