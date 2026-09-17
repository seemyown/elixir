# Elixir

[English](README.md) | [Русский](README.ru.md)

**A type-safe SQL expression builder for Go.**

> **Note:** This project was developed with AI assistance.

Not an ORM — you define tables in Go and compose typed expressions into SQL. Public API: `github.com/seemyown/elixir`. AST and compiler live under `internal/`.

## Features

- Typed columns and expressions (`Column[T]`, predicates, functions, aggregates)
- `LIKE` / `ILIKE` (Postgres) / `BETWEEN` / `CAST` / arithmetic
- Fluent `Select` / `Insert` / `Update` / `Delete` builders
- `DISTINCT`, `UNION` / `UNION ALL`, optional `FROM` (e.g. `SELECT EXISTS(...)`)
- `All()` / `AllOf(table)` for `SELECT *` and `table.*`
- `Engine` with a default dialect so `Compile()` needs no extra argument
- Postgres, MySQL, and SQLite dialects
- Named inserts (`Set` / `WithDefaults`), `RETURNING`, `INSERT … SELECT`, `ON CONFLICT`, MySQL `ON DUPLICATE KEY`
- Schema-qualified tables (`TableRef.Schema`)
- `Null[T]` / `NullColumn[T]` for nullable values (`*T`, `sql.Scanner` / `driver.Valuer`, compatible with `sql.Null[T]`)
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
	Bio:      elixir.NullColumn[string]{Name: "bio"},
	Active:   elixir.Column[bool]{Name: "active"}.WithDefault(true),
})

u := elixir.As(Users, "u") // aliased copy for JOINs
```

`Column[T]` is NOT NULL — `SetNull` / `Null[T]` do not compile against it.  
`NullColumn[T]` accepts `SetPtr(*T)`, `SetOpt(Null[T])`, `SetSQLNull(sql.Null[T])`, `SetNull()`, and `IsNull` / `IsNotNull`.

`WithDefault` records a typed DB default (including a zero value of `T`) for `Insert.WithDefaults`. Schema-qualify a table with `TableRef{Schema: "app", Name: "users"}` (`FROM "app"."users"`); column qualifiers stay alias/name. `Bind` / `As` do not clear `Schema`.

```go
Users.Bio.SetPtr(nil)                 // NULL
Users.Bio.SetPtr(&s)                  // value via pointer
Users.Bio.SetOpt(elixir.Some("x"))
Users.Bio.EqOpt(elixir.None[string]()) // IS NULL
Users.Bio.IsNull()
p := elixir.Some("x").Ptr()           // *string
```

### Select

```go
q := elixir.Select(Users.ID, Users.Email).
	From(Users).
	Where(Users.Email.Eq("a@example.com")).
	OrderBy(Users.ID.Desc()).
	Limit(10)

sql, args, err := q.Compile(elixir.Postgres())

// SELECT DISTINCT
elixir.Select(Users.Email).Distinct().From(Users)

// UNION / UNION ALL (ORDER BY / LIMIT apply to the whole compound)
elixir.Select(Users.ID).From(Users).
	UnionAll(elixir.Select(Orders.UserID).From(Orders)).
	OrderBy(Users.ID.Asc()).
	Limit(10)

// SELECT EXISTS (...) — no FROM on the outer query
elixir.Select(elixir.Exists(elixir.Select(Users.ID).From(Users).Where(Users.ID.Eq(1))))

// SELECT *
elixir.Select(elixir.All()).From(Users)

// SELECT "users".* (handy with joins)
elixir.Select(elixir.AllOf(Users), Orders.Amount).From(Users)
```

`Like` / `ILike` / `Between` / `Cast` / arithmetic:

```go
Users.Email.Like("%@example.com")
Users.Email.ILike("%@Example.com") // Postgres only; Compile errors on MySQL/SQLite
Users.ID.Between(1, 10)
elixir.Cast[int64](Users.Email, "integer")
Orders.Amount.Add(1.5)
```

### Joins

Prefer `As` so column qualifiers match the alias. Use `EqExpr` for column–column conditions.

```go
U := elixir.As(Users, "u")
O := elixir.As(Orders, "o")

on := U.ID.EqExpr(O.UserID)

// INNER JOIN
elixir.Select(U.Email, O.Amount).
	From(U).
	Join(O, on)

// LEFT JOIN
elixir.Select(U.Email, O.Amount).
	From(U).
	LeftJoin(O, on)

// RIGHT JOIN
elixir.Select(U.Email, O.Amount).
	From(U).
	RightJoin(O, on)

// FULL JOIN
elixir.Select(U.Email, O.Amount).
	From(U).
	FullJoin(O, on)

// Multiple joins
P := elixir.As(Payments, "p")
elixir.Select(U.Email, O.Amount, P.ID).
	From(U).
	LeftJoin(O, U.ID.EqExpr(O.UserID)).
	LeftJoin(P, O.ID.EqExpr(P.OrderID))
```

`GROUP BY` / `HAVING`, subqueries, CTEs, window functions, and `CASE` are also supported.

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
	Users.Active.SetExpr(elixir.Excluded(Users.Active)),
)

// ON DUPLICATE KEY (MySQL)
elixir.Insert(Users).Set(
	Users.Email.Set("a@example.com"),
	Users.Active.Set(false),
).OnDuplicateKey(
	Users.Active.SetExpr(elixir.ValuesCol(Users.Active)),
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

`ON CONFLICT` is Postgres/SQLite syntax. MySQL uses `OnDuplicateKey` (`ON DUPLICATE KEY UPDATE`). `ILIKE` is PostgreSQL-only (`Compile` returns `elixir: ILIKE is PostgreSQL-only` on other dialects). `RETURNING` is emitted for all dialects; MySQL support depends on version.

Live-database checks: `go test -tags=integration ./...` (Postgres/MySQL DSNs via `ELIXIR_PG_DSN` / `ELIXIR_MYSQL_DSN`; SQLite in-process). Keep drivers in `go.mod` with `GOFLAGS='-tags=integration' go mod tidy`.

## Status

Experimental; the API may change.

**Implemented:** typed expressions and predicates (`LIKE`/`ILIKE`/`BETWEEN`/`CAST`/arithmetic); SELECT/INSERT/UPDATE/DELETE; `DISTINCT`; `UNION`; optional FROM; `All`/`AllOf`; `Bind`/`As`; schema-qualified tables; `Null`/`NullColumn` (`Scan`/`Value`); `WithDefault`; named insert + `WithDefaults`; `RETURNING`; `INSERT … SELECT`; `ON CONFLICT` + `EXCLUDED`; MySQL `ON DUPLICATE KEY` + `VALUES(col)`; joins, aggregates, functions, subqueries, CTEs, windows, CASE; Postgres/MySQL/SQLite dialects; `Engine`; `sqlerr`; CI integration job.

**Not yet:** recursive CTE; pgx integration helpers.

## License

[MIT](LICENSE)
