package elixir

import (
	"github.com/seemyown/elixir/internal/ast"
	"github.com/seemyown/elixir/internal/compiler"
)

// DeleteQuery is a fluent DELETE statement builder.
type DeleteQuery struct {
	node    ast.DeleteNode
	dialect Dialect
}

// Delete starts a DELETE FROM table statement.
func Delete(table Table) DeleteQuery {
	return DeleteQuery{node: ast.DeleteNode{Table: table.tableRef()}}
}

// WithDialect embeds a dialect so Compile() / ToSQL() can omit the dialect argument.
func (q DeleteQuery) WithDialect(d Dialect) DeleteQuery {
	q.dialect = d
	return q
}

// Where appends predicates combined with AND.
func (q DeleteQuery) Where(preds ...Predicate) DeleteQuery {
	for _, p := range preds {
		q.node.Where = append(q.node.Where, p.node)
	}
	return q
}

// Returning appends a RETURNING clause (Postgres / SQLite; MySQL support varies).
func (q DeleteQuery) Returning(cols ...Expression) DeleteQuery {
	for _, c := range cols {
		q.node.Returning = append(q.node.Returning, c.exprNode())
	}
	return q
}

// Compile renders the query.
//
// With no arguments, uses the dialect embedded via Engine.Delete / WithDialect.
// With one Dialect argument, that dialect is used (one-off or override).
func (q DeleteQuery) Compile(ds ...Dialect) (string, []any, error) {
	d, err := resolveDialect(q.dialect, ds...)
	if err != nil {
		return "", nil, err
	}
	return compiler.CompileDelete(asCompilerDialect(d), q.node)
}

// ToSQL is an alias for Compile.
func (q DeleteQuery) ToSQL(ds ...Dialect) (string, []any, error) {
	return q.Compile(ds...)
}
