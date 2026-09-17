# Elixir

**Типобезопасный конструктор SQL-выражений для Go.**

> **Note:** Этот проект разработан с помощью ИИ.

Elixir — это SQL toolkit для Go, построенный вокруг типизированных выражений, компонуемого AST и явного построения запросов.

Идея вдохновлена моделью выражений SQLAlchemy, но адаптирована под систему типов и generics в Go.

Elixir — **не ORM**. Он не управляет сущностями, не скрывает семантику SQL и не генерирует код на основе схемы базы данных.

Вместо этого таблицы и их колонки явно описываются в Go, после чего из типизированных выражений собираются SQL-запросы.

Публичный API — пакет `github.com/seemyown/elixir`. Узлы AST и SQL-компилятор находятся в `internal/` и не входят в стабильную поверхность API.

## Идея

Вместо SQL-строк:

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

Elixir позволяет описать тот же запрос через типизированные таблицы:

```go
Users.ID.Eq(id)
Users.Email.Eq(email)
```

и собрать их в запрос:

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

// Или задать диалект один раз через Engine:
e := New(Postgres())
sql, args, err = e.Compile(query)
// или: e.Select(...).From(...).Compile()  // без аргумента dialect
```

Выражения представляются в виде AST, после чего компилируются в SQL и набор аргументов.

```text
Go-выражения
      │
      ▼
 Типизированный AST
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

## Цели

Elixir стремится предоставить:

* **типобезопасность на этапе компиляции;**
* **отсутствие code generation;**
* **отсутствие reflection для построения запросов;**
* **компонуемые SQL-выражения;**
* **явную семантику SQL;**
* **небольшой и предсказуемый API;**
* **независимость от конкретного database driver;**
* **поддержку SQL dialects;**
* **полноценную работу с SQL-выражениями и функциями.**

Библиотека должна делать некорректные операции невозможными или затруднительными ещё на этапе компиляции, а не обнаруживать их во время выполнения.

Например:

```go
Users.ID.Eq(42)       // OK
Users.Email.Eq("foo") // OK

Users.ID.Eq("42")     // ошибка компиляции
Users.Email.Eq(42)    // ошибка компиляции
```

## Таблицы и колонки

Таблицы явно описываются в Go через struct literals. Встраивайте `TableRef`, чтобы значение таблицы можно было передать в `From` / `Join`. Колонки задаются как `Column[T]{Name: "..."}`; оберните значение в `Bind`, чтобы каждая колонка унаследовала имя родительской таблицы (без дублирования `Table:` на каждом поле).

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
// После Bind: Users.ID.Table == "users" и т.д.
```

Поле `Table` по-прежнему можно задать вручную; `Bind` не перезаписывает уже заполненные значения.

Копия таблицы с алиасом для JOIN (перебиндивает квалификаторы колонок):

```go
u := As(Users, "u")
// u.Alias == "u", u.ID.Table == "u"
Select(u.ID).From(u).Where(u.ID.Eq(1))
```

Или голый `TableRef` без перебинда колонок:

```go
u := TableRef{Name: "users", Alias: "u"}
// или: TableRef{Name: "users"}.As("u")
```

Метаданные колонок:

* `Nullable` — подсказка схемы; для записи SQL `NULL` используйте `SetNull(col)` / `col.SetNull()`
* `Default` — если не `nil`, то `Insert(...).WithDefaults().Set(...)` добавит SQL `DEFAULT` для пропущенных колонок

Таблица предоставляет типизированные колонки:

```go
Users.ID
Users.Email
Users.CreatedAt
```

Колонка — это не просто строка с названием поля.

Она является типизированным SQL-выражением:

```go
Expr[int64]
Expr[string]
Expr[time.Time]
```

Благодаря этому система типов Go может участвовать в построении SQL-запросов.

## Выражения

Любое значение, которое может быть представлено в SQL, является выражением.

Например:

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

Выражения можно комбинировать:

```go
Lower(Users.Email).Eq("foo@example.com")
```

или:

```go
Count(Users.ID).Gt(10)
```

Это одна из основных идей Elixir:

> Колонка, функция, агрегат, литерал, CASE-выражение и другие SQL-конструкции должны использовать единую модель выражений и свободно комбинироваться между собой.

## Предикаты

Предикаты — это выражения, используемые в условиях (`WHERE`, `HAVING`, `JOIN ON`).

```go
Users.ID.Eq(42)
Users.ID.Gt(100)
Users.Email.Ne("test@example.com")
Users.Email.IsNull()
Users.Email.IsNotNull()
```

Сравнение двух выражений — через `EqExpr` / `NeExpr` / `GtExpr` / …:

```go
Orders.UserID.EqExpr(Users.ID)
```

Логические выражения можно объединять:

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

С квалифицированными колонками полученный AST компилируется (PostgreSQL) в:

```sql
(("users"."active" = $1) AND ("users"."id" > $2))
```

с аргументами:

```go
[]any{true, int64(100)}
```

Отдельные предикаты можно рендерить через `CompilePredicate(dialect, pred)`.

## Функции

Встроенные хелперы оборачивают распространённые SQL-функции в типизированные выражения:

```go
Lower(Users.Email)
Upper(Users.Email)
Coalesce(Users.Email, Value("unknown"))
```

Произвольные функции:

```go
Func[string]("TRIM", Users.Email)
```

## Агрегации

Агрегации являются обычными типизированными выражениями.

```go
Count(Users.ID)
CountAll()
Sum(Orders.Amount)
Avg(Orders.Amount)
Min(Orders.Amount)
Max(Orders.Amount)
```

Поскольку агрегаты являются выражениями, они могут участвовать в других выражениях:

```go
Count(Users.ID).Gt(10)
```

Например, использоваться в `HAVING`:

```go
Having(
    Count(Users.ID).Gt(10),
)
```

Такая модель позволяет строить более сложные SQL-конструкции без создания отдельного API для каждого вида агрегации:

```go
Count(Distinct(Users.Email))
```

## SQL AST

Elixir не собирает SQL простым конкатенированием строк во всех слоях API.

Выражения внутри представлены в виде AST.

Например:

```go
Users.ID.Eq(42)
```

концептуально может выглядеть так:

```text
Binary
├── Column("users", "id")
└── Value(42)
```

А:

```go
Count(Users.ID).Gt(10)
```

превращается в:

```text
Binary
├── Function("COUNT")
│   └── Column("users", "id")
└── Value(10)
```

Компилятор отвечает за преобразование этого AST в SQL.

Таким образом, построение запроса отделено от его компиляции.

## Построение запросов

Query layer строится поверх системы выражений.

```go
// SELECT *
query := Select(All()).From(Users).Where(Users.ID.Eq(1))

// SELECT "users".* (удобно при JOIN)
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
// ToSQL — алиас для Compile
```

JOIN:

```go
query := Select(Users.Email, Orders.Amount).
    From(Users).
    Join(Orders, Orders.UserID.EqExpr(Users.ID)).
    Where(Users.Active.Eq(true))

// также: LeftJoin, RightJoin, FullJoin
```

Сортировка через `Asc()` / `Desc()` на колонках или выражениях:

```go
OrderBy(Users.CreatedAt.Desc(), Users.ID.Asc())
```

Сам запрос также является AST, состоящим из тех же базовых компонентов.

## INSERT / UPDATE / DELETE

Именованный / частичный INSERT (рекомендуется) — записываются только переданные колонки:

```go
Insert(Users).Set(
    Users.Email.Set("a@example.com"),
    Users.Active.Set(true),
)

// Active не передали — WithDefaults добавит DEFAULT из метаданных колонки:
Insert(Users).WithDefaults().Set(
    Users.Email.Set("a@example.com"),
)

Insert(Users).Set(
    Users.Email.Set("a@example.com"),
    Users.Bio.SetNull(),
).Returning(Users.ID, Users.Email)
```

Позиционный INSERT (по-прежнему поддерживается):

```go
Insert(Users).
    Columns(Users.Email, Users.Active).
    Values(Value("a@example.com"), Value(true))

// несколько строк:
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

// выражение справа:
Update(Users).
    Set(SetExpr(Users.Email, Lower(Users.Email))).
    Where(Users.ID.Eq(1))
```

```go
Delete(Users).Where(Users.ID.Eq(1)).Returning(Users.ID)
```

`RETURNING` генерируется для всех диалектов; PostgreSQL и SQLite поддерживают его широко, поддержка в MySQL зависит от версии.
## Подзапросы

Скалярный подзапрос как типизированное выражение:

```go
maxID := Subquery[int64](Select(Max(Users.ID)).From(Users))
Select(Users.Email).From(Users).Where(Users.ID.EqExpr(maxID))
```

`IN` / `NOT IN` со списком или запросом:

```go
Users.ID.In(1, 2, 3)
Users.ID.InQuery(Select(Orders.UserID).From(Orders).Where(Orders.Amount.Gt(100)))
```

`EXISTS` / `NOT EXISTS`:

```go
Exists(Select(Orders.ID).From(Orders).Where(Orders.UserID.EqExpr(Users.ID)))
```

Подзапрос в `FROM` / `JOIN` (алиас обязателен):

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

Также доступны `With(...).Insert` / `Update` / `Delete`, если CTE используется в DML.

## Оконные функции

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

Фреймы: `Rows` / `Range` с `UnboundedPreceding`, `CurrentRow`, `UnboundedFollowing`, `Preceding(n)`, `Following(n)`.

## CASE-выражения

Searched `CASE`:

```go
Case[string]().
    When(Users.Active.Eq(true), Value("yes")).
    When(Users.ID.Gt(100), Value("vip")).
    Else(Value("no"))
```

Simple `CASE` по выражению:

```go
CaseOn[bool, string](Users.Active).
    When(Value(true), Value("active")).
    When(Value(false), Value("inactive")).
    End()
```

## Не ORM

Elixir не пытается быть ORM.

Здесь не будет:

```go
User.Find(...)
User.Save(...)
User.Delete(...)
```

магии.

Нет неявных запросов, которые внезапно выполняются при обращении к полю.

Нет необходимости представлять всю базу данных в виде объектного графа.

Elixir занимается построением SQL, сохраняя семантику базы данных видимой разработчику.

## Без code generation

Описание таблиц пишется непосредственно в Go как обычные структуры и значения:

```go
var Users = Bind(UserTable{
    TableRef: TableRef{Name: "users"},
    ID:       Column[int64]{Name: "id"},
    Email:    Column[string]{Name: "email"},
})
```

Не требуется отдельный этап интроспекции схемы и генерации сотен файлов с моделями.

Цель — держать описание базы данных максимально близко к коду, который с ней работает.

## Database drivers

Elixir не должен зависеть от конкретного database driver.

Ядро библиотеки отвечает за:

```text
Expression → AST → SQL + arguments
```

А выполнение запроса остаётся ответственностью приложения или отдельной интеграции.

Например, PostgreSQL-интеграция может передавать скомпилированный запрос непосредственно в `pgx`.

## Диалекты и Engine

Синтаксис SQL отличается между базами данных.

Поэтому Elixir отделяет модель выражений от конкретного способа генерации SQL с помощью dialects:

```go
Postgres() // $1, $2, …  и идентификаторы в двойных кавычках
MySQL()    // плейсхолдеры ? и идентификаторы в backticks
SQLite()   // плейсхолдеры ? и идентификаторы в двойных кавычках
```

Разовая компиляция (как раньше):

```go
sql, args, err := query.Compile(Postgres())
sql, args, err := query.Compile(MySQL())
sql, args, err := query.Compile(SQLite())
```

Если большинство запросов идут в одну БД, создайте `Engine` один раз:

```go
e := New(Postgres()) // или WithDialect(Postgres())

// Скомпилировать любой Select/Insert/Update/Delete с диалектом engine:
sql, args, err := e.Compile(query)

// Или строить через engine — тогда Compile()/ToSQL() без аргумента dialect:
q := e.Select(Users.ID).From(Users).Where(Users.ID.Eq(1))
sql, args, err = q.Compile()

// Переопределение по-прежнему возможно:
sql, args, err = q.Compile(MySQL())
```

`Compile(ds ...Dialect)` / `ToSQL(ds ...Dialect)`: ноль аргументов — встроенный диалект (`Engine` / `WithDialect`); один аргумент — override или разовый диалект. Передача dialect явно полностью поддерживается для существующих вызовов.

## Статус

Elixir находится на экспериментальной стадии.

Реализовано:

* [x] Типизированные выражения
* [x] Колонки (`Column[T]`)
* [x] Литералы (`Value`)
* [x] Сравнения (`Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, варианты `*Expr`)
* [x] `IS NULL` / `IS NOT NULL`
* [x] AND / OR / NOT
* [x] `IN` / `NOT IN` (значения и подзапросы)
* [x] `EXISTS` / `NOT EXISTS`
* [x] SQL-функции (`Lower`, `Upper`, `Coalesce`, `Func`)
* [x] Агрегации (`Count`, `CountAll`, `Sum`, `Avg`, `Min`, `Max`, `Distinct`)
* [x] SELECT / FROM / WHERE
* [x] `All()` / `AllOf(table)` (`SELECT *` / `table.*`)
* [x] GROUP BY / HAVING
* [x] ORDER BY / LIMIT / OFFSET
* [x] JOIN / LEFT / RIGHT / FULL
* [x] INSERT / UPDATE / DELETE
* [x] Именованный INSERT через `Set` / `WithDefaults` / `SetNull`
* [x] `RETURNING` (Insert / Update / Delete)
* [x] `Bind` / `As` (наследование / перебинд квалификатора таблицы у колонок)
* [x] Метаданные колонок (`Nullable`, `Default`)
* [x] Подзапросы (скаляр, `IN`, `FROM` / `JOIN`)
* [x] CTE (`WITH ... AS`)
* [x] Оконные функции (`OVER`, фреймы, `ROW_NUMBER` / `RANK` / `DENSE_RANK`)
* [x] CASE (searched и simple)
* [x] Диалекты PostgreSQL, MySQL, SQLite
* [x] `Engine` / `New` / session-style `Compile()` без повторной передачи dialect
* [x] `Compile` / `ToSQL` / `CompilePredicate`

Ещё нет:

* [ ] Рекурсивные CTE
* [ ] `INSERT ... SELECT`
* [ ] Интеграция с pgx

API будет изменяться по мере разработки и стабилизации модели выражений.

## Философия

Elixir придерживается простого принципа:

> **SQL должен оставаться SQL. Go должен сделать его построение безопаснее.**

Библиотека не должна скрывать базу данных за настолько большим слоем абстракций, что разработчику приходится изучать абстракцию вместо SQL.

Вместо этого Elixir предоставляет типизированные строительные блоки для непосредственного описания SQL в Go.

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

Без магии.

Без генерируемого кода.

Без случайно превратившегося в ORM проекта.

Просто типизированный SQL.
