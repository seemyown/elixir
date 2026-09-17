# Elixir

[English](README.md) | [Русский](README.ru.md)

**Типобезопасный конструктор SQL-выражений для Go.**

> **Note:** Этот проект разработан с помощью ИИ.

Не ORM — таблицы описываются в Go, запросы собираются из типизированных выражений. Публичный API: `github.com/seemyown/elixir`. AST и компилятор — в `internal/`.

## Возможности

- Типизированные колонки и выражения (`Column[T]`, предикаты, функции, агрегаты)
- `LIKE` / `ILIKE` (Postgres) / `BETWEEN` / `CAST` / арифметика
- Fluent-билдеры `Select` / `Insert` / `Update` / `Delete`
- `DISTINCT`, `UNION` / `UNION ALL`, необязательный `FROM` (например `SELECT EXISTS(...)`)
- `All()` / `AllOf(table)` для `SELECT *` и `table.*`
- `Engine` с диалектом по умолчанию — `Compile()` без лишнего аргумента
- Диалекты Postgres, MySQL и SQLite
- Именованный insert (`Set` / `WithDefaults`), `RETURNING`, `INSERT … SELECT`, `ON CONFLICT`, MySQL `ON DUPLICATE KEY`
- Квалификация схемы (`TableRef.Schema`)
- `Null[T]` / `NullColumn[T]` для nullable-значений (`*T`, `sql.Scanner` / `driver.Valuer`, совместим с `sql.Null[T]`)
- Классификация ошибок БД в пакете [`sqlerr`](./sqlerr)

## Установка

```bash
go get github.com/seemyown/elixir
```

## Быстрый старт

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

## Базовый API

### Таблицы и колонки

Встраивайте `TableRef`, объявляйте поля `Column[T]`, оборачивайте в `Bind`, чтобы колонки унаследовали имя таблицы:

```go
var Users = elixir.Bind(UserTable{
	TableRef: elixir.TableRef{Name: "users"},
	ID:       elixir.Column[int64]{Name: "id"},
	Email:    elixir.Column[string]{Name: "email"},
	Bio:      elixir.NullColumn[string]{Name: "bio"},
	Active:   elixir.Column[bool]{Name: "active"}.WithDefault(true),
})

u := elixir.As(Users, "u") // копия с алиасом для JOIN
```

`Column[T]` — NOT NULL: `SetNull` / `Null[T]` к нему не скомпилируются.  
`NullColumn[T]` принимает `SetPtr(*T)`, `SetOpt(Null[T])`, `SetSQLNull(sql.Null[T])`, `SetNull()`, а также `IsNull` / `IsNotNull`.

`WithDefault` записывает типизированный DB default (включая нулевое значение `T`) для `Insert.WithDefaults`. Схему задаёт `TableRef{Schema: "app", Name: "users"}` (`FROM "app"."users"`); квалификаторы колонок по-прежнему alias/name. `Bind` / `As` не затирают `Schema`.

```go
Users.Bio.SetPtr(nil)                 // NULL
Users.Bio.SetPtr(&s)                  // значение через указатель
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

// UNION / UNION ALL (ORDER BY / LIMIT относятся ко всему составному запросу)
elixir.Select(Users.ID).From(Users).
	UnionAll(elixir.Select(Orders.UserID).From(Orders)).
	OrderBy(Users.ID.Asc()).
	Limit(10)

// SELECT EXISTS (...) — внешний запрос без FROM
elixir.Select(elixir.Exists(elixir.Select(Users.ID).From(Users).Where(Users.ID.Eq(1))))

// SELECT *
elixir.Select(elixir.All()).From(Users)

// SELECT "users".* (удобно при JOIN)
elixir.Select(elixir.AllOf(Users), Orders.Amount).From(Users)
```

`Like` / `ILike` / `Between` / `Cast` / арифметика:

```go
Users.Email.Like("%@example.com")
Users.Email.ILike("%@Example.com") // только Postgres; на MySQL/SQLite Compile вернёт ошибку
Users.ID.Between(1, 10)
elixir.Cast[int64](Users.Email, "integer")
Orders.Amount.Add(1.5)
```

### Joins

Для алиасов лучше `As`, чтобы квалификаторы колонок совпали. Условие колонка–колонка — через `EqExpr`.

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

// Несколько JOIN
P := elixir.As(Payments, "p")
elixir.Select(U.Email, O.Amount, P.ID).
	From(U).
	LeftJoin(O, U.ID.EqExpr(O.UserID)).
	LeftJoin(P, O.ID.EqExpr(P.OrderID))
```

Также поддерживаются `GROUP BY` / `HAVING`, подзапросы, CTE, оконные функции и `CASE`.

### Insert / Update / Delete

```go
// Именованный insert
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

### Engine и диалекты

```go
e := elixir.New(elixir.Postgres()) // или MySQL(), SQLite()
sql, args, err := e.Compile(q)
// или: e.Select(...).From(...).Compile()
```

`Compile(ds ...Dialect)`: без аргументов — встроенный диалект (`Engine` / `WithDialect`); один аргумент — переопределение.

### sqlerr

```go
import "github.com/seemyown/elixir/sqlerr"

if sqlerr.IsUniqueViolation(err) { /* … */ }
if sqlerr.IsForeignKeyViolation(err) { /* … */ }
if sqlerr.IsNoRows(err) { /* … */ }
_ = sqlerr.Code(err) // SQLSTATE / код драйвера, если доступен
```

Хелперы смотрят SQLSTATE, мягкие формы ошибок драйверов и текстовые паттерны для Postgres / MySQL / SQLite без импорта этих драйверов.

## Диалекты

| Диалект | Фабрика | Плейсхолдеры | Кавычки |
|---------|---------|--------------|---------|
| PostgreSQL | `Postgres()` | `$1`, `$2`, … | `"ident"` |
| MySQL | `MySQL()` | `?` | `` `ident` `` |
| SQLite | `SQLite()` | `?` | `"ident"` |

`ON CONFLICT` — синтаксис Postgres/SQLite. Для MySQL — `OnDuplicateKey` (`ON DUPLICATE KEY UPDATE`). `ILIKE` только в PostgreSQL (`Compile` возвращает `elixir: ILIKE is PostgreSQL-only` на других диалектах). `RETURNING` эмитится для всех диалектов; поддержка в MySQL зависит от версии.

Проверки на живых БД: `go test -tags=integration ./...` (DSN Postgres/MySQL — `ELIXIR_PG_DSN` / `ELIXIR_MYSQL_DSN`; SQLite in-process). Чтобы драйверы остались в `go.mod`: `GOFLAGS='-tags=integration' go mod tidy`.

## Статус

Экспериментальный API; возможны изменения.

**Реализовано:** типизированные выражения и предикаты (`LIKE`/`ILIKE`/`BETWEEN`/`CAST`/арифметика); SELECT/INSERT/UPDATE/DELETE; `DISTINCT`; `UNION`; необязательный FROM; `All`/`AllOf`; `Bind`/`As`; схемы; `Null`/`NullColumn` (`Scan`/`Value`); `WithDefault`; именованный insert + `WithDefaults`; `RETURNING`; `INSERT … SELECT`; `ON CONFLICT` + `EXCLUDED`; MySQL `ON DUPLICATE KEY` + `VALUES(col)`; JOIN, агрегаты, функции, подзапросы, CTE, окна, CASE; диалекты Postgres/MySQL/SQLite; `Engine`; `sqlerr`; CI integration job.

**Пока нет:** рекурсивные CTE; хелперы интеграции с pgx.

## Лицензия

[MIT](LICENSE)
