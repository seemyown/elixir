package elixir

import (
	"fmt"

	"github.com/seemyown/elixir/internal/compiler"
)

// Engine holds a default Dialect for compiling queries.
//
// Prefer Engine when most queries target one database:
//
//	e := elixir.New(elixir.Postgres())
//	sql, args, err := e.Compile(query)
//
// Or build queries via Engine so Compile() needs no dialect argument:
//
//	q := e.Select(Users.ID).From(Users)
//	sql, args, err := q.Compile()
type Engine struct {
	dialect Dialect
}

// New returns an Engine with the given default dialect.
func New(d Dialect) *Engine {
	return &Engine{dialect: d}
}

// WithDialect is an alias for New.
func WithDialect(d Dialect) *Engine {
	return New(d)
}

// Dialect returns the Engine's default dialect.
func (e *Engine) Dialect() Dialect {
	if e == nil {
		return nil
	}
	return e.dialect
}

// Compilable is implemented by SelectQuery, InsertQuery, UpdateQuery, and DeleteQuery.
type Compilable interface {
	Compile(ds ...Dialect) (string, []any, error)
}

// Compile renders q using the Engine's dialect (equivalent to q.Compile(e.Dialect())).
func (e *Engine) Compile(q Compilable) (string, []any, error) {
	if e == nil || e.dialect == nil {
		return "", nil, fmt.Errorf("elixir: dialect is nil")
	}
	return q.Compile(e.dialect)
}

// Select starts a SELECT with this Engine's dialect embedded.
func (e *Engine) Select(cols ...Expression) SelectQuery {
	q := Select(cols...)
	if e != nil {
		q.dialect = e.dialect
	}
	return q
}

// Insert starts an INSERT with this Engine's dialect embedded.
func (e *Engine) Insert(table Table) InsertQuery {
	q := Insert(table)
	if e != nil {
		q.dialect = e.dialect
	}
	return q
}

// Update starts an UPDATE with this Engine's dialect embedded.
func (e *Engine) Update(table Table) UpdateQuery {
	q := Update(table)
	if e != nil {
		q.dialect = e.dialect
	}
	return q
}

// Delete starts a DELETE with this Engine's dialect embedded.
func (e *Engine) Delete(table Table) DeleteQuery {
	q := Delete(table)
	if e != nil {
		q.dialect = e.dialect
	}
	return q
}

// With starts a WITH clause; following Select/Insert/Update/Delete inherit the Engine dialect.
func (e *Engine) With(ctes ...CTEDef) WithBuilder {
	w := With(ctes...)
	if e != nil {
		w.dialect = e.dialect
	}
	return w
}

// resolveDialect picks an override dialect, else the embedded one.
// Compile(ds ...Dialect): 0 args use embedded; 1 arg overrides; more than 1 is an error.
func resolveDialect(embedded Dialect, ds ...Dialect) (Dialect, error) {
	switch len(ds) {
	case 0:
		if embedded == nil {
			return nil, fmt.Errorf("elixir: dialect required (pass a Dialect or build via Engine / WithDialect)")
		}
		return embedded, nil
	case 1:
		if ds[0] == nil {
			return nil, fmt.Errorf("elixir: dialect is nil")
		}
		return ds[0], nil
	default:
		return nil, fmt.Errorf("elixir: Compile accepts at most one Dialect argument")
	}
}

func asCompilerDialect(d Dialect) compiler.Dialect {
	return d
}
