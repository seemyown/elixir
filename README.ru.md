# Elixir

[English](README.md) | [Русский](README.ru.md)

**Типобезопасный конструктор SQL-выражений для Go.**

> **Note:** Этот проект разработан с помощью ИИ.

Не ORM — таблицы описываются в Go, запросы собираются из типизированных выражений. Публичный API: `github.com/seemyown/elixir`. AST и компилятор — в `internal/`.

## Возможности

- Типизированные колонки и выражения (`Column[T]`, предикаты, функции, агрегаты)
- Fluent-билдеры `Select` / `Insert` / `Update` / `Delete`
- `All()` / `AllOf(table)` для `SELECT *` и `table.*`
- `Engine` с диалектом по умолчанию — `Compile()` без лишнего аргумента
- Диалекты Postgres, MySQL и SQLite
- Именованный insert (`Set` / `WithDefaults`), `RETURNING`, `INSERT … SELECT`, `ON CONFLICT`
- `Null[T]` / `NullColumn[T]` для nullable-значений (`*T`, совместим с `sql.Null[T]`)
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
	Active:   elixir.Column[bool]{Name: "active", Default: true},
})

u := elixir.As(Users, "u") // копия с алиасом для JOIN
```

`Column[T]` — NOT NULL: `SetNull` / `Null[T]` к нему не скомпилируются.  
`NullColumn[T]` принимает `SetPtr(*T)`, `SetOpt(Null[T])`, `SetSQLNull(sql.Null[T])`, `SetNull()`.

`Default` на любой колонке используется в `Insert.WithDefaults`.

```go
Users.Bio.SetPtr(nil)                 // NULL
Users.Bio.SetPtr(&s)                  // значение через указатель
Users.Bio.SetOpt(elixir.Some("x"))
Users.Bio.EqOpt(elixir.None[string]()) // IS NULL
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

// SELECT *
elixir.Select(elixir.All()).From(Users)

// SELECT "users".* (удобно при JOIN)
elixir.Select(elixir.AllOf(Users), Orders.Amount).From(Users)
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
	elixir.Set(Users.Active, true),
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

`ON CONFLICT` — синтаксис Postgres/SQLite. MySQL его не поддерживает — при необходимости используйте сырой `ON DUPLICATE KEY UPDATE`. `RETURNING` эмитится для всех диалектов; поддержка в MySQL зависит от версии.

## Статус

Экспериментальный API; возможны изменения.

**Реализовано:** типизированные выражения и предикаты; SELECT/INSERT/UPDATE/DELETE; `All`/`AllOf`; `Bind`/`As`; `Null`/`NullColumn`; `Default` у колонок; именованный insert + `WithDefaults`; `RETURNING`; `INSERT … SELECT`; `ON CONFLICT`; JOIN, агрегаты, функции, подзапросы, CTE, окна, CASE; диалекты Postgres/MySQL/SQLite; `Engine`; `sqlerr`.

**Пока нет:** рекурсивные CTE; хелперы интеграции с pgx.

## Лицензия

[MIT](LICENSE)
