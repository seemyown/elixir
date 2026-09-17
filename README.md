# Elixir

**A type-safe SQL expression builder for Go.**

> **Note:** This project was developed with AI assistance.

Elixir is a SQL toolkit for Go built around typed expressions, composable AST nodes and explicit query construction.

It is inspired by the expression model of SQLAlchemy, but designed around Go's type system and generics.

Elixir is **not an ORM**. It does not manage entities, hide SQL semantics, or generate code from your database schema.

Instead, you define your database tables explicitly and build SQL queries from typed Go expressions.

The public API lives in package `github.com/seemyown/elixir`. AST nodes and the SQL compiler live under `internal/` and are not part of the stable surface.

## Idea

Instead of writing SQL as strings:

```go
db.Query(
    ctx,
    `SELECT id, email
     FROM users
     WHERE id = $1 AND email = $2`,
    id,
    email,
)
```

Elixir lets you describe the same query using typed table definitions:

```go
Users.ID.Eq(id)
Users.Email.Eq(email)
```

and compose them into a query:

```go
query := Select(
    Users.ID,
    Users.Email,
).
    From(Users).
    Where(
        Users.ID.Eq(id),
        Users.Email.Eq(email),
    )

sql, args, err := query.Compile(Postgres())

// Or set the dialect once via Engine:
e := New(Postgres())
sql, args, err = e.Compile(query)
// or: e.Select(...).From(...).Compile()  // no dialect argument
```

The expressions are represented as an AST and compiled into SQL with its arguments.

```text
Go expressions
      │
      ▼
   Typed AST
      │
      ▼
 SQL compiler
      │
      ▼
 SQL + arguments
      │
      ▼
 database driver
```

## Design goals

Elixir aims to provide:

* **Compile-time type safety**
* **No code generation**
* **No reflection-based query generation**
* **Composable SQL expressions**
* **Explicit SQL semantics**
* **A small and predictable API**
* **Database-driver independence**
* **Dialect support**
* **First-class support for SQL expressions and functions**

The library should make invalid operations difficult to express rather than discovering them at runtime.

For example:

```go
Users.ID.Eq(42)       // valid
Users.Email.Eq("foo") // valid

Users.ID.Eq("42")     // compile-time error
Users.Email.Eq(42)    // compile-time error
```

## Tables and columns

Tables are explicitly described in Go with struct literals. Embed `TableRef` so the table value can be passed to `From` / `Join`. Define columns with `Column[T]{Name: "..."}` and wrap the value in `Bind` so each column inherits the parent table name (no need to repeat `Table:` on every field).

```go
type UserTable struct {
    TableRef

    ID        Column[int64]
    Email     Column[string]
    Password  Column[string]
    Bio       Column[string]
    Active    Column[bool]
    CreatedAt Column[time.Time]
}

var Users = Bind(UserTable{
    TableRef:  TableRef{Name: "users"},
    ID:        Column[int64]{Name: "id"},
    Email:     Column[string]{Name: "email"},
    Password:  Column[string]{Name: "password"},
    Bio:       Column[string]{Name: "bio", Nullable: true},
    Active:    Column[bool]{Name: "active", Default: true},
    CreatedAt: Column[time.Time]{Name: "created_at"},
})
// After Bind: Users.ID.Table == "users", etc.
```

You can still set `Table` manually when you prefer not to use `Bind`. Manual values are left untouched by `Bind` when already set.

Aliased table copy for joins (rebinds all column qualifiers):

```go
u := As(Users, "u")
// u.Alias == "u", u.ID.Table == "u"
Select(u.ID).From(u).Where(u.ID.Eq(1))
```

Or a bare `TableRef` alias without rebinding columns:

```go
u := TableRef{Name: "users", Alias: "u"}
// or: TableRef{Name: "users"}.As("u")
```

Column metadata:

* `Nullable` — schema hint; use `SetNull(col)` / `col.SetNull()` to write SQL `NULL`
* `Default` — when non-nil, `Insert(...).WithDefaults().Set(...)` emits SQL `DEFAULT` for omitted columns

A table exposes typed columns:

```go
Users.ID
Users.Email
Users.CreatedAt
```

A column is not just a string containing a column name.

It is a typed SQL expression:

```go
Expr[int64]
Expr[string]
Expr[time.Time]
```

This allows the type system to participate in query construction.

## Expressions

Everything that produces a SQL value is an expression.

Examples:

```go
Users.ID
Users.Email
Count(Users.ID)
Sum(Orders.Amount)
Lower(Users.Email)
Value("unknown")
All()           // SQL *
AllOf(Users)    // "users".*
```

Expressions can be composed:

```go
Lower(Users.Email).Eq("foo@example.com")
```

or:

```go
Count(Users.ID).Gt(10)
```

This is one of the core ideas of Elixir:

> A column, function, aggregate, literal, CASE expression or other SQL value should be composable through the same expression model.

## Predicates

Predicates are expressions used for conditions (`WHERE`, `HAVING`, `JOIN ON`).

```go
Users.ID.Eq(42)
Users.ID.Gt(100)
Users.Email.Ne("test@example.com")
Users.Email.IsNull()
Users.Email.IsNotNull()
```

Compare two expressions with `EqExpr` / `NeExpr` / `GtExpr` / …:

```go
Orders.UserID.EqExpr(Users.ID)
```

Logical expressions can be composed:

```go
And(
    Users.Active.Eq(true),
    Users.ID.Gt(100),
)
```

```go
Or(
    Users.Email.Eq("a@example.com"),
    Users.Email.Eq("b@example.com"),
)
```

```go
Not(Users.Active.Eq(true))
```

With qualified columns, the resulting AST compiles (PostgreSQL) to:

```sql
(("users"."active" = $1) AND ("users"."id" > $2))
```

with:

```go
[]any{true, int64(100)}
```

Standalone predicates can also be rendered with `CompilePredicate(dialect, pred)`.

## Functions

Built-in helpers wrap common SQL functions as typed expressions:

```go
Lower(Users.Email)
Upper(Users.Email)
Coalesce(Users.Email, Value("unknown"))
```

Arbitrary functions:

```go
Func[string]("TRIM", Users.Email)
```

## Aggregations

Aggregations are ordinary typed expressions.

```go
Count(Users.ID)
CountAll()
Sum(Orders.Amount)
Avg(Orders.Amount)
Min(Orders.Amount)
Max(Orders.Amount)
```

Because aggregates produce expressions, they can participate in other expressions:

```go
Count(Users.ID).Gt(10)
```

which can be used in `HAVING`:

```go
Having(
    Count(Users.ID).Gt(10),
)
```

This model also makes it possible to build more advanced SQL constructs such as:

```go
Count(Distinct(Users.Email))
```

without introducing a separate aggregation-specific query API.

## SQL AST

Elixir does not construct SQL by concatenating strings throughout the API.

Expressions are represented internally as an AST.

For example:

```go
Users.ID.Eq(42)
```

can be represented conceptually as:

```text
Binary
├── Column("users", "id")
└── Value(42)
```

While:

```go
Count(Users.ID).Gt(10)
```

becomes:

```text
Binary
├── Function("COUNT")
│   └── Column("users", "id")
└── Value(10)
```

The compiler is responsible for turning this AST into SQL.

This keeps query construction separate from SQL rendering.

## Query building

The query layer is intentionally built on top of the expression system.

```go
// SELECT *
query := Select(All()).From(Users).Where(Users.ID.Eq(1))

// SELECT "users".* (useful with joins)
query := Select(AllOf(Users), Orders.Amount).
    From(Users).
    Join(Orders, Users.ID.EqExpr(Orders.UserID))
```

```go
query := Select(
    Users.Email,
    Count(Users.ID).As("total"),
).
    From(Users).
    Where(
        Users.Active.Eq(true),
    ).
    GroupBy(
        Users.Email,
    ).
    Having(
        Count(Users.ID).Gt(5),
    ).
    OrderBy(
        Users.Email.Asc(),
    ).
    Limit(10).
    Offset(20)

sql, args, err := query.Compile(Postgres())
// ToSQL is an alias for Compile
```

Joins:

```go
query := Select(Users.Email, Orders.Amount).
    From(Users).
    Join(Orders, Orders.UserID.EqExpr(Users.ID)).
    Where(Users.Active.Eq(true))

// also: LeftJoin, RightJoin, FullJoin
```

Ordering uses `Asc()` / `Desc()` on columns or expressions:

```go
OrderBy(Users.CreatedAt.Desc(), Users.ID.Asc())
```

The query itself is another AST structure composed from the same expression primitives.

## INSERT / UPDATE / DELETE

Named / partial INSERT (recommended) — only the columns you pass are written:

```go
Insert(Users).Set(
    Users.Email.Set("a@example.com"),
    Users.Active.Set(true),
)

// omit Active but emit DEFAULT from column metadata:
Insert(Users).WithDefaults().Set(
    Users.Email.Set("a@example.com"),
)

Insert(Users).Set(
    Users.Email.Set("a@example.com"),
    Users.Bio.SetNull(),
).Returning(Users.ID, Users.Email)
```

Positional INSERT (still supported):

```go
Insert(Users).
    Columns(Users.Email, Users.Active).
    Values(Value("a@example.com"), Value(true))

// multiple rows:
Insert(Users).
    Columns(Users.Email).
    Values(Value("a@example.com")).
    Values(Value("b@example.com"))
```

```go
Update(Users).
    Set(
        Set(Users.Email, "new@example.com"),
        Set(Users.Active, false),
    ).
    Where(Users.ID.Eq(1)).
    Returning(Users.ID)

// expression on the right-hand side:
Update(Users).
    Set(SetExpr(Users.Email, Lower(Users.Email))).
    Where(Users.ID.Eq(1))
```

```go
Delete(Users).Where(Users.ID.Eq(1)).Returning(Users.ID)
```

`RETURNING` is emitted for all dialects; PostgreSQL and SQLite support it broadly, MySQL support depends on version.
## Subqueries

Scalar subquery as a typed expression:

```go
maxID := Subquery[int64](Select(Max(Users.ID)).From(Users))
Select(Users.Email).From(Users).Where(Users.ID.EqExpr(maxID))
```

`IN` / `NOT IN` with a list or a query:

```go
Users.ID.In(1, 2, 3)
Users.ID.InQuery(Select(Orders.UserID).From(Orders).Where(Orders.Amount.Gt(100)))
```

`EXISTS` / `NOT EXISTS`:

```go
Exists(Select(Orders.ID).From(Orders).Where(Orders.UserID.EqExpr(Users.ID)))
```

Subquery in `FROM` / `JOIN` (alias required):

```go
sub := Select(Users.ID, Users.Email).
    From(Users).
    Where(Users.Active.Eq(true)).
    AsTable("u")

id := Column[int64]{Table: "u", Name: "id"}
Select(id).From(sub)
```

## CTE (WITH)

```go
active := TableRef{Name: "active_users"}
activeID := Column[int64]{Table: "active_users", Name: "id"}

query := With(
    CTE("active_users",
        Select(Users.ID).From(Users).Where(Users.Active.Eq(true)),
    ),
).
    Select(activeID).
    From(active)

sql, args, err := query.Compile(Postgres())
```

`With(...).Insert` / `Update` / `Delete` are also available when a CTE feeds a DML statement.

## Window functions

```go
Count(Users.ID).Over(
    PartitionBy(Users.Email).OrderBy(Users.CreatedAt.Desc()),
)

RowNumber().Over(
    WinOrderBy(Users.ID.Asc()).Rows(UnboundedPreceding(), CurrentRow()),
)

Rank().Over(PartitionBy(Users.Active).OrderBy(Users.ID.Desc()))
DenseRank().Over(WinOrderBy(Orders.Amount.Desc()))
```

Frames: `Rows` / `Range` with `UnboundedPreceding`, `CurrentRow`, `UnboundedFollowing`, `Preceding(n)`, `Following(n)`.

## CASE expressions

Searched `CASE`:

```go
Case[string]().
    When(Users.Active.Eq(true), Value("yes")).
    When(Users.ID.Gt(100), Value("vip")).
    Else(Value("no"))
```

Simple `CASE` on an expression:

```go
CaseOn[bool, string](Users.Active).
    When(Value(true), Value("active")).
    When(Value(false), Value("inactive")).
    End()
```

## No ORM

Elixir does not try to be an ORM.

There is no:

```go
User.Find(...)
User.Save(...)
User.Delete(...)
```

magic.

There are no implicit queries triggered by accessing a field.

There is no requirement to model your entire database as an object graph.

Elixir focuses on constructing SQL while keeping the database semantics visible to the developer.

## No code generation

Table definitions are written explicitly as ordinary Go structs and values:

```go
var Users = Bind(UserTable{
    TableRef: TableRef{Name: "users"},
    ID:       Column[int64]{Name: "id"},
    Email:    Column[string]{Name: "email"},
})
```

There is no schema introspection step and no generated model directory containing hundreds of files.

The goal is to keep the database representation close to the code that uses it.

## Database drivers

Elixir is intended to be independent from the database driver.

The core library is responsible for:

```text
Expression → AST → SQL + arguments
```

Execution is the responsibility of the application or an integration layer.

For example, a PostgreSQL integration can pass the compiled query directly to `pgx`.

## Dialects and Engine

SQL syntax differs between databases.

Elixir therefore separates SQL generation from the expression model through dialects:

```go
Postgres() // $1, $2, …  and double-quoted identifiers
MySQL()    // ? placeholders and backtick identifiers
SQLite()   // ? placeholders and double-quoted identifiers
```

One-off compile (unchanged):

```go
sql, args, err := query.Compile(Postgres())
sql, args, err := query.Compile(MySQL())
sql, args, err := query.Compile(SQLite())
```

When most queries target one database, create an `Engine` once:

```go
e := New(Postgres()) // or WithDialect(Postgres())

// Compile any Select/Insert/Update/Delete with the engine dialect:
sql, args, err := e.Compile(query)

// Or build from the engine so Compile()/ToSQL() need no dialect argument:
q := e.Select(Users.ID).From(Users).Where(Users.ID.Eq(1))
sql, args, err = q.Compile()

// Optional override still works:
sql, args, err = q.Compile(MySQL())
```

`Compile(ds ...Dialect)` / `ToSQL(ds ...Dialect)`: zero arguments use the dialect embedded via `Engine` / `WithDialect`; one argument overrides (or supplies a one-off dialect). Passing a dialect remains fully supported for existing callers.

## Status

Elixir is currently experimental.

Implemented:

* [x] Typed expressions
* [x] Columns (`Column[T]`)
* [x] Literals (`Value`)
* [x] Comparison predicates (`Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `*Expr` variants)
* [x] `IS NULL` / `IS NOT NULL`
* [x] AND / OR / NOT
* [x] `IN` / `NOT IN` (values and subqueries)
* [x] `EXISTS` / `NOT EXISTS`
* [x] SQL functions (`Lower`, `Upper`, `Coalesce`, `Func`)
* [x] Aggregations (`Count`, `CountAll`, `Sum`, `Avg`, `Min`, `Max`, `Distinct`)
* [x] SELECT / FROM / WHERE
* [x] `All()` / `AllOf(table)` (`SELECT *` / `table.*`)
* [x] GROUP BY / HAVING
* [x] ORDER BY / LIMIT / OFFSET
* [x] JOIN / LEFT / RIGHT / FULL
* [x] INSERT / UPDATE / DELETE
* [x] Named INSERT via `Set` / `WithDefaults` / `SetNull`
* [x] `RETURNING` (Insert / Update / Delete)
* [x] `Bind` / `As` (inherit or rebind column table qualifiers)
* [x] Column metadata (`Nullable`, `Default`)
* [x] Subqueries (scalar, `IN`, `FROM` / `JOIN`)
* [x] CTE (`WITH ... AS`)
* [x] Window functions (`OVER`, frames, `ROW_NUMBER` / `RANK` / `DENSE_RANK`)
* [x] CASE (searched and simple)
* [x] PostgreSQL, MySQL, SQLite dialects
* [x] `Engine` / `New` / session-style `Compile()` without repeating dialect
* [x] `Compile` / `ToSQL` / `CompilePredicate`

Not yet:

* [ ] Recursive CTE
* [ ] `INSERT ... SELECT`
* [ ] pgx integration

The API is expected to change while the expression model is being developed.

## Philosophy

Elixir follows a simple principle:

> **SQL should remain SQL. Go should make constructing it safer.**

The library should not hide the database behind an abstraction so large that developers need to learn the abstraction instead of SQL.

Instead, Elixir provides typed building blocks for expressing SQL directly in Go.

```text
Go
 │
 │ typed expressions
 ▼
Elixir
 │
 │ SQL AST
 ▼
SQL
 │
 ▼
Database
```

No magic.

No generated code.

No accidental ORM.

Just typed SQL.
