package elixir

import (
	"fmt"

	"github.com/seemyown/elixir/internal/ast"
	"github.com/seemyown/elixir/internal/compiler"
)

// InsertQuery is a fluent INSERT statement builder.
type InsertQuery struct {
	node         ast.InsertNode
	dialect      Dialect
	src          any
	withDefaults bool
	assigns      []ast.AssignNode
}

// Insert starts an INSERT INTO table statement.
func Insert(table Table) InsertQuery {
	return InsertQuery{
		node: ast.InsertNode{Table: table.tableRef()},
		src:  table,
	}
}

// WithDialect embeds a dialect so Compile() / ToSQL() can omit the dialect argument.
func (q InsertQuery) WithDialect(d Dialect) InsertQuery {
	q.dialect = d
	return q
}

// Columns optionally lists target columns for INSERT (positional Values or Select mode).
func (q InsertQuery) Columns(cols ...Expression) InsertQuery {
	for _, c := range cols {
		q.node.Columns = append(q.node.Columns, c.exprNode())
	}
	return q
}

// Values appends one VALUES row. Expressions are typically Value(...) literals.
func (q InsertQuery) Values(vals ...Expression) InsertQuery {
	row := make([]ast.Node, len(vals))
	for i, v := range vals {
		row[i] = v.exprNode()
	}
	q.node.Rows = append(q.node.Rows, row)
	return q
}

// Select sets an INSERT ... SELECT source. Mutually exclusive with Values / Set.
func (q InsertQuery) Select(sel SelectQuery) InsertQuery {
	node := sel.node
	q.node.Select = &node
	return q
}

// Set appends named column assignments (partial / full insert without positional Values).
// Use Set / SetExpr / SetNull helpers or Column.Set methods.
func (q InsertQuery) Set(assigns ...Assignment) InsertQuery {
	for _, a := range assigns {
		q.assigns = append(q.assigns, a.node)
	}
	return q
}

// WithDefaults, when used with Set, appends explicit DEFAULT for schema columns
// that declare Default and were omitted from Set.
func (q InsertQuery) WithDefaults() InsertQuery {
	q.withDefaults = true
	return q
}

// OnConflict starts an ON CONFLICT clause with optional inference columns.
// With no columns, only DoNothing is valid (SQLite / Postgres catch-all).
// MySQL does not support ON CONFLICT (use ON DUPLICATE KEY UPDATE in raw SQL).
func (q InsertQuery) OnConflict(cols ...Expression) OnConflictBuilder {
	columns := make([]ast.Node, len(cols))
	for i, c := range cols {
		columns[i] = c.exprNode()
	}
	return OnConflictBuilder{q: q, columns: columns}
}

// OnConflictConstraint starts ON CONFLICT ON CONSTRAINT name (Postgres).
// SQLite does not support ON CONSTRAINT; MySQL does not support ON CONFLICT.
func (q InsertQuery) OnConflictConstraint(name string) OnConflictBuilder {
	return OnConflictBuilder{q: q, constraint: name}
}

// OnConflictBuilder configures the action for an ON CONFLICT clause.
type OnConflictBuilder struct {
	q          InsertQuery
	columns    []ast.Node
	constraint string
}

// DoNothing finishes the clause as ON CONFLICT ... DO NOTHING.
func (b OnConflictBuilder) DoNothing() InsertQuery {
	b.q.node.Conflict = &ast.ConflictNode{
		Columns:    append([]ast.Node(nil), b.columns...),
		Constraint: b.constraint,
		DoNothing:  true,
	}
	return b.q
}

// DoUpdate finishes the clause as ON CONFLICT ... DO UPDATE SET ....
// A conflict target (columns or constraint) is required.
func (b OnConflictBuilder) DoUpdate(assigns ...Assignment) InsertQuery {
	updates := make([]ast.AssignNode, len(assigns))
	for i, a := range assigns {
		updates[i] = a.node
	}
	b.q.node.Conflict = &ast.ConflictNode{
		Columns:    append([]ast.Node(nil), b.columns...),
		Constraint: b.constraint,
		Updates:    updates,
	}
	return b.q
}

// Returning appends a RETURNING clause (Postgres / SQLite; MySQL support varies).
func (q InsertQuery) Returning(cols ...Expression) InsertQuery {
	for _, c := range cols {
		q.node.Returning = append(q.node.Returning, c.exprNode())
	}
	return q
}

// Compile renders the query.
//
// With no arguments, uses the dialect embedded via Engine.Insert / WithDialect.
// With one Dialect argument, that dialect is used (one-off or override).
func (q InsertQuery) Compile(ds ...Dialect) (string, []any, error) {
	d, err := resolveDialect(q.dialect, ds...)
	if err != nil {
		return "", nil, err
	}
	node, err := q.materialize()
	if err != nil {
		return "", nil, err
	}
	return compiler.CompileInsert(asCompilerDialect(d), node)
}

// ToSQL is an alias for Compile.
func (q InsertQuery) ToSQL(ds ...Dialect) (string, []any, error) {
	return q.Compile(ds...)
}

func (q InsertQuery) materialize() (ast.InsertNode, error) {
	node := q.node

	if node.Select != nil {
		if len(q.assigns) > 0 {
			return ast.InsertNode{}, fmt.Errorf("elixir: Insert.Select cannot be mixed with Set")
		}
		if len(node.Rows) > 0 {
			return ast.InsertNode{}, fmt.Errorf("elixir: Insert.Select cannot be mixed with Values")
		}
		return node, nil
	}

	if len(q.assigns) == 0 {
		return node, nil
	}
	if len(node.Columns) > 0 || len(node.Rows) > 0 {
		return ast.InsertNode{}, fmt.Errorf("elixir: Insert.Set cannot be mixed with Columns/Values")
	}

	assigns := append([]ast.AssignNode(nil), q.assigns...)
	if q.withDefaults {
		assigned := make(map[string]struct{}, len(assigns))
		for _, a := range assigns {
			if col, ok := a.Column.(ast.ColumnNode); ok {
				assigned[col.Name] = struct{}{}
			}
		}
		for _, def := range columnDefaults(q.src) {
			col, ok := def.Column.(ast.ColumnNode)
			if !ok {
				continue
			}
			if _, exists := assigned[col.Name]; exists {
				continue
			}
			assigns = append(assigns, def)
			assigned[col.Name] = struct{}{}
		}
	}

	if len(assigns) == 0 {
		return ast.InsertNode{}, fmt.Errorf("elixir: INSERT Set requires at least one assignment")
	}

	cols := make([]ast.Node, len(assigns))
	row := make([]ast.Node, len(assigns))
	for i, a := range assigns {
		cols[i] = a.Column
		row[i] = a.Value
	}
	node.Columns = cols
	node.Rows = [][]ast.Node{row}
	return node, nil
}
