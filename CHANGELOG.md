# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-09-17

### Added

- Predicates and expressions: `Like` / `NotLike` / `ILike` / `NotILike`, `Between` / `NotBetween`, `Cast`, arithmetic (`Add` / `Sub` / `Mul` / `Div`) with `*Expr` variants
- `SELECT DISTINCT` via `SelectQuery.Distinct`
- `UNION` / `UNION ALL` via `Union` / `UnionAll` (`ORDER BY` / `LIMIT` / `OFFSET` apply to the compound query)
- `FROM` is optional when there are no JOINs (e.g. `SELECT EXISTS(...)`, `SELECT 1`)
- `TableRef.Schema` for `schema.table` in FROM / DML (column qualifiers stay alias/name)
- `Excluded` (`EXCLUDED.col`) and `ValuesCol` (`VALUES(col)`) for upserts
- `InsertQuery.OnDuplicateKey` (MySQL `ON DUPLICATE KEY UPDATE`)
- `Null[T]` implements `sql.Scanner` and `driver.Valuer` via `sql.Null[T]`
- `Column[T].WithDefault` / `HasDefault` (and the same on `NullColumn[T]`)
- Integration tests (`go test -tags=integration`) against Postgres, MySQL, and SQLite; CI job with live database services

### Changed

- **Breaking:** `Column[T].IsNull` / `IsNotNull` are removed. Use `Expr[T]` or `NullColumn[T]` (nullable columns).
- **Breaking:** `Default` is typed as `T` instead of `any`. A zero value no longer means “no default”; set `HasDefault` via `WithDefault(v)`.

## [v0.1.4] - 2026-09-17

### Added

- `Null[T]` optional value (`Some` / `None` / `FromPtr` / `Ptr` / `FromSQL` / `SQL`) compatible with `database/sql.Null[T]`
- `NullColumn[T]` for NULL-able columns: `SetPtr`, `SetOpt`, `SetSQLNull`, `SetNull`, `EqOpt` / `EqPtr`
- `ValueOpt` / `ValuePtr` for INSERT `Values` cells

### Changed

- `Column[T]` is non-NULL only: removed `Nullable` flag and `Column.SetNull` (use `NullColumn` — compile-time safe)

## [v0.1.3] - 2026-09-17

### Added

- `All()` / `AllOf(table)` for `SELECT *` and qualified `table.*`
- GitHub Actions: CI (`go test` / `go vet`) and Release workflow (tag + GitHub Release from CHANGELOG)

## [v0.1.2] - 2026-09-17

### Fixed

- Bug fixes and small corrections after v0.1.1

## [v0.1.1] - 2026-09-17

### Added

- `INSERT … SELECT` via `InsertQuery.Select`
- `ON CONFLICT` / `ON CONFLICT ON CONSTRAINT` with `DoNothing` and `DoUpdate` (Postgres / SQLite)
- Package [`github.com/seemyown/elixir/sqlerr`](./sqlerr) for driver-agnostic SQL error classification:
  - `IsUniqueViolation`, `IsForeignKeyViolation`, `IsNotNullViolation`, `IsCheckViolation`
  - `IsSerializationFailure` (includes common deadlocks / lock contention)
  - `IsNoRows`, `Code`
- Rewritten English and Russian READMEs in a standard Go library layout

### Changed

- Compiler and AST support for insert-select sources and conflict clauses

## [v0.1.0] - 2026-09-17

### Added

- Initial release of `github.com/seemyown/elixir`
- Typed expression model (`Column[T]`, predicates, literals, functions, aggregates)
- Fluent `Select` / `Insert` / `Update` / `Delete` builders
- `Engine` (`New` / `WithDialect`) with default dialect and dialect-free `Compile()`
- Dialects: PostgreSQL, MySQL, SQLite
- Table helpers: `Bind`, `As`, column metadata (`Nullable`, `Default`)
- Named inserts (`Set` / `SetExpr` / `SetNull` / `WithDefaults`) and `RETURNING`
- Joins, subqueries, CTEs, window functions, searched/simple `CASE`
- Internal AST and SQL compiler under `internal/`
- Mirror unit and example tests
- MIT license; bilingual README (EN / RU)
