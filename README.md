# Elixir

[English](README.md) | [Русский](README.ru.md)

**A type-safe SQL expression builder for Go.**

> **Note:** This project was developed with AI assistance.

Not an ORM — you define tables in Go and compose typed expressions into SQL. Public API: `github.com/seemyown/elixir`. AST and compiler live under `internal/`.

## Features

- Typed columns and expressions (`Column[T]`, predicates, functions, aggregates)
- Fluent `Select` / `Insert` / `Update` / `Delete` builders
- `All()` / `AllOf(table)` for `SELECT *` and `table.*`
- `Engine` with a default dialect so `Compile()` needs no extra argument
- Postgres, MySQL, and SQLite dialects
- Named inserts (`Set` / `WithDefaults`), `RETURNING`, `INSERT … SELECT`, `ON CONFLICT`
- Driver-agnostic SQL error helpers in [`sqlerr`](./sqlerr)

## Installation

```bash
go get github.com/seemyown/elixir
```

## Quick start

```go
package main

import (
	"fmt"

	"github.com/seemyown/elixir"
)

type UserTable struct {
	elixir.TableRef
	ID    elixir.Column[int64]
	Email elixir.Column[string]
}

var Users = elixir.Bind(UserTable{
	TableRef: elixir.TableRef{Name: "users"},
	ID:       elixir.Column[int64]{Name: "id"},
	Email:    elixir.Column[string]{Name: "email"},
})

func main() {
	e := elixir.New(elixir.Postgres())

	q := e.Select(Users.ID, Users.Email).
		From(Users).
		Where(Users.ID.Eq(1))

	sql, args, err := q.Compile()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)  // SELECT "users"."id", "users"."email" FROM "users" WHERE ("users"."id" = $1)
	fmt.Println(args) // [1]
}
```

## Basic API

### Tables and columns

Embed `TableRef`, define `Column[T]` fields, wrap with `Bind` so columns inherit the table name:

```go
var Users = elixir.Bind(UserTable{
	TableRef: elixir.TableRef{Name: "users"},
	ID:       elixir.Column[int64]{Name: "id"},
	Email:    elixir.Column[string]{Name: "email"},
	Bio:      elixir.Column[string]{Name: "bio", Nullable: true},
	Active:   elixir.Column[bool]{Name: "active", Default: true},
})

u := elixir.As(Users, "u") // aliased copy for JOINs
```

Optional metadata: `Nullable`, `Default` (used by `Insert.WithDefaults`).

### Select

```go
q := elixir.Select(Users.ID, Users.Email).
	From(Users).
	Where(Users.Email.Eq("a@example.com")).
	OrderBy(Users.ID.Desc()).
	Limit(10)

sql, args, err := q.Compile(elixir.Postgres())

// SELECT *
elixir.Select(elixir.All()).From(Users)

// SELECT "users".* (handy with joins)
elixir.Select(elixir.AllOf(Users), Orders.Amount).From(Users)
```

Joins, `GROUP BY` / `HAVING`, subqueries, CTEs, window functions, and `CASE` are also supported.

### Insert / Update / Delete

```go
// Named insert
elixir.Insert(Users).Set(
	Users.Email.Set("a@example.com"),
	Users.Active.Set(true),
).Returning(Users.ID)

// INSERT … SELECT
elixir.Insert(Users).
	Columns(Users.Email, Users.Active).
	Select(elixir.Select(Other.Email, Other.Active).From(Other))

// ON CONFLICT (Postgres / SQLite)
elixir.Insert(Users).Set(Users.Email.Set("a@example.com")).
	OnConflict(Users.Email).DoNothing()

elixir.Insert(Users).Set(
	Users.Email.Set("a@example.com"),
	Users.Active.Set(false),
).OnConflict(Users.Email).DoUpdate(
	elixir.Set(Users.Active, true),
)

elixir.Update(Users).
	Set(Users.Active.Set(false)).
	Where(Users.ID.Eq(1)).
	Returning(Users.ID)

elixir.Delete(Users).Where(Users.ID.Eq(1))
```

### Engine and dialects

```go
e := elixir.New(elixir.Postgres()) // or MySQL(), SQLite()
sql, args, err := e.Compile(q)
// or: e.Select(...).From(...).Compile()
```

`Compile(ds ...Dialect)`: zero args use the embedded dialect (`Engine` / `WithDialect`); one arg overrides.

### sqlerr

```go
import "github.com/seemyown/elixir/sqlerr"

if sqlerr.IsUniqueViolation(err) { /* … */ }
if sqlerr.IsForeignKeyViolation(err) { /* … */ }
if sqlerr.IsNoRows(err) { /* … */ }
_ = sqlerr.Code(err) // SQLSTATE / driver code when available
```

Helpers inspect SQLSTATE, soft driver error shapes, and message patterns for Postgres / MySQL / SQLite without importing those drivers.

## Dialects

| Dialect | Factory | Placeholders | Quoting |
|---------|---------|--------------|---------|
| PostgreSQL | `Postgres()` | `$1`, `$2`, … | `"ident"` |
| MySQL | `MySQL()` | `?` | `` `ident` `` |
| SQLite | `SQLite()` | `?` | `"ident"` |

`ON CONFLICT` is Postgres/SQLite syntax. MySQL does not support it — use raw `ON DUPLICATE KEY UPDATE` if needed. `RETURNING` is emitted for all dialects; MySQL support depends on version.

## Status

Experimental; the API may change.

**Implemented:** typed expressions and predicates; SELECT/INSERT/UPDATE/DELETE; `All`/`AllOf`; `Bind`/`As`; column metadata; named insert + `WithDefaults`; `RETURNING`; `INSERT … SELECT`; `ON CONFLICT`; joins, aggregates, functions, subqueries, CTEs, windows, CASE; Postgres/MySQL/SQLite dialects; `Engine`; `sqlerr`.

**Not yet:** recursive CTE; pgx integration helpers.

## License

[MIT](LICENSE)
