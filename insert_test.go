package elixir

import (
	"strings"
	"testing"
)

func TestInsert(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email, Users.Active).
		Values(Value("a@example.com"), Value(true))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "active") VALUES ($1, $2)`,
		[]any{"a@example.com", true},
	)
}

func TestInsertMultiRow(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email).
		Values(Value("a@example.com")).
		Values(Value("b@example.com"))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") VALUES ($1), ($2)`,
		[]any{"a@example.com", "b@example.com"},
	)
}

func TestInsertMySQL(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email).
		Values(Value("x"))

	sql, args, err := q.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"INSERT INTO `users` (`email`) VALUES (?)",
		[]any{"x"},
	)
}

func TestInsertErrors(t *testing.T) {
	if _, _, err := Insert(Users).Compile(Postgres()); err == nil {
		t.Fatal("expected values error")
	}
	if _, _, err := Insert(Users).Columns(Users.Email).Values(Value("a"), Value(true)).Compile(Postgres()); err == nil {
		t.Fatal("expected column/value mismatch error")
	}
}

func TestInsertSet(t *testing.T) {
	q := Insert(Users).Set(
		Users.Email.Set("a@example.com"),
		Users.Active.Set(true),
	)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "active") VALUES ($1, $2)`,
		[]any{"a@example.com", true},
	)
}

func TestInsertSetWithDefaultZero(t *testing.T) {
	type tdef struct {
		TableRef
		Name Column[string]
		N    Column[int]
	}
	tbl := Bind(tdef{
		TableRef: TableRef{Name: "t"},
		Name:     Column[string]{Name: "name"},
		N:        Column[int]{Name: "n"}.WithDefault(0),
	})
	if !tbl.N.HasDefault || tbl.N.Default != 0 {
		t.Fatalf("WithDefault(0): %+v", tbl.N)
	}
	q := Insert(tbl).WithDefaults().Set(tbl.Name.Set("x"))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "t" ("name", "n") VALUES ($1, DEFAULT)`,
		[]any{"x"},
	)
}

func TestInsertSetWithDefaults(t *testing.T) {
	q := Insert(Users).WithDefaults().Set(
		Users.Email.Set("a@example.com"),
	)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "active") VALUES ($1, DEFAULT)`,
		[]any{"a@example.com"},
	)
}

func TestInsertSetNull(t *testing.T) {
	type bioTable struct {
		TableRef
		Email Column[string]
		Bio   NullColumn[string]
	}
	tbio := Bind(bioTable{
		TableRef: TableRef{Name: "users"},
		Email:    Column[string]{Name: "email"},
		Bio:      NullColumn[string]{Name: "bio"},
	})
	q := Insert(tbio).Set(
		tbio.Email.Set("a@example.com"),
		tbio.Bio.SetNull(),
	)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "bio") VALUES ($1, NULL)`,
		[]any{"a@example.com"},
	)
}

func TestInsertSetMixError(t *testing.T) {
	_, _, err := Insert(Users).
		Columns(Users.Email).
		Set(Users.Email.Set("x")).
		Compile(Postgres())
	if err == nil {
		t.Fatal("expected mix error")
	}
}

func TestInsertReturning(t *testing.T) {
	q := Insert(Users).
		Set(Users.Email.Set("a@example.com")).
		Returning(Users.ID, Users.Email)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") VALUES ($1) RETURNING "users"."id", "users"."email"`,
		[]any{"a@example.com"},
	)
}

func TestInsertWithDialectToSQL(t *testing.T) {
	sql, args, err := Insert(Users).
		Set(Users.Email.Set("x")).
		WithDialect(Postgres()).
		ToSQL()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `INSERT INTO "users" ("email") VALUES ($1)`, []any{"x"})
}

func TestCTEWithInsert(t *testing.T) {
	q := With(
		CTE("src", Select(Users.Email).From(Users)),
	).Insert(Users).Columns(Users.Email).Values(Value("z@example.com"))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "src" AS (SELECT "users"."email" FROM "users") INSERT INTO "users" ("email") VALUES ($1)`,
		[]any{"z@example.com"},
	)
}

func TestInsertSelect(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email, Users.Active).
		Select(
			Select(Orders.UserID, Value(true)).
				From(Orders).
				Where(Orders.Amount.Gt(100)),
		)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "active") SELECT "orders"."user_id", $1 FROM "orders" WHERE ("orders"."amount" > $2)`,
		[]any{true, 100.0},
	)
}

func TestInsertSelectReturning(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email).
		Select(Select(Users.Email).From(Users).Where(Users.Active.Eq(true))).
		Returning(Users.ID)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") SELECT "users"."email" FROM "users" WHERE ("users"."active" = $1) RETURNING "users"."id"`,
		[]any{true},
	)
}

func TestCTEWithInsertSelect(t *testing.T) {
	srcEmail := Column[string]{Name: "email"}
	q := With(
		CTE("src", Select(Users.Email).From(Users).Where(Users.Active.Eq(true))),
	).Insert(Users).
		Columns(Users.Email).
		Select(Select(srcEmail).From(TableRef{Name: "src"}))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "src" AS (SELECT "users"."email" FROM "users" WHERE ("users"."active" = $1)) INSERT INTO "users" ("email") SELECT "email" FROM "src"`,
		[]any{true},
	)
}

func TestInsertSelectMixErrors(t *testing.T) {
	sel := Select(Users.Email).From(Users)
	if _, _, err := Insert(Users).Columns(Users.Email).Values(Value("a")).Select(sel).Compile(Postgres()); err == nil {
		t.Fatal("expected Values/Select mix error")
	}
	if _, _, err := Insert(Users).Set(Users.Email.Set("a")).Select(sel).Compile(Postgres()); err == nil {
		t.Fatal("expected Set/Select mix error")
	}
}

func TestInsertOnConflictDoNothing(t *testing.T) {
	q := Insert(Users).
		Set(Users.Email.Set("a@example.com")).
		OnConflict(Users.Email).
		DoNothing()
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") VALUES ($1) ON CONFLICT ("email") DO NOTHING`,
		[]any{"a@example.com"},
	)
}

func TestInsertOnConflictDoNothingNoTarget(t *testing.T) {
	q := Insert(Users).
		Set(Users.Email.Set("a@example.com")).
		OnConflict().
		DoNothing()
	sql, args, err := q.Compile(SQLite())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") VALUES (?) ON CONFLICT DO NOTHING`,
		[]any{"a@example.com"},
	)
}

func TestInsertOnConflictDoUpdate(t *testing.T) {
	q := Insert(Users).
		Set(
			Users.Email.Set("a@example.com"),
			Users.Active.Set(false),
		).
		OnConflict(Users.Email).
		DoUpdate(
			Set(Users.Active, true),
			SetExpr(Users.Email, Lower(Users.Email)),
		)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "active") VALUES ($1, $2) ON CONFLICT ("email") DO UPDATE SET "active" = $3, "email" = LOWER("users"."email")`,
		[]any{"a@example.com", false, true},
	)
}

func TestInsertOnConflictConstraint(t *testing.T) {
	q := Insert(Users).
		Set(Users.Email.Set("a@example.com")).
		OnConflictConstraint("users_email_key").
		DoNothing()
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") VALUES ($1) ON CONFLICT ON CONSTRAINT "users_email_key" DO NOTHING`,
		[]any{"a@example.com"},
	)
}

func TestInsertOnConflictMultiColumn(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email, Users.ID).
		Values(Value("a@example.com"), Value(int64(1))).
		OnConflict(Users.Email, Users.ID).
		DoNothing()
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "id") VALUES ($1, $2) ON CONFLICT ("email", "id") DO NOTHING`,
		[]any{"a@example.com", int64(1)},
	)
}

func TestInsertOnConflictWithSelect(t *testing.T) {
	q := Insert(Users).
		Columns(Users.Email).
		Select(Select(Users.Email).From(Users)).
		OnConflict(Users.Email).
		DoNothing().
		Returning(Users.ID)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email") SELECT "users"."email" FROM "users" ON CONFLICT ("email") DO NOTHING RETURNING "users"."id"`,
		nil,
	)
}

func TestInsertOnConflictExcluded(t *testing.T) {
	q := Insert(Users).
		Set(
			Users.Email.Set("a@example.com"),
			Users.Active.Set(false),
		).
		OnConflict(Users.Email).
		DoUpdate(
			Users.Active.SetExpr(Excluded(Users.Active)),
		)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("email", "active") VALUES ($1, $2) ON CONFLICT ("email") DO UPDATE SET "active" = EXCLUDED."active"`,
		[]any{"a@example.com", false},
	)
}

func TestInsertOnDuplicateKey(t *testing.T) {
	q := Insert(Users).
		Set(
			Users.Email.Set("a@example.com"),
			Users.Active.Set(false),
		).
		OnDuplicateKey(
			Users.Active.SetExpr(ValuesCol(Users.Active)),
		)
	sql, args, err := q.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"INSERT INTO `users` (`email`, `active`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `active` = VALUES(`active`)",
		[]any{"a@example.com", false},
	)
}

func TestInsertOnDuplicateKeyEmptyError(t *testing.T) {
	if _, _, err := Insert(Users).
		Set(Users.Email.Set("a")).
		OnDuplicateKey().
		Compile(MySQL()); err == nil {
		t.Fatal("expected empty ON DUPLICATE KEY error")
	}
}

func TestInsertMixConflictAndDuplicateKey(t *testing.T) {
	_, _, err := Insert(Users).
		Set(Users.Email.Set("a")).
		OnConflict(Users.Email).
		DoNothing().
		OnDuplicateKey(Users.Active.Set(true)).
		Compile(Postgres())
	if err == nil {
		t.Fatal("expected mix error")
	}
	if !strings.Contains(err.Error(), "cannot mix ON CONFLICT and ON DUPLICATE KEY") {
		t.Fatalf("got %v", err)
	}
}

func TestInsertOnConflictErrors(t *testing.T) {
	if _, _, err := Insert(Users).
		Set(Users.Email.Set("a")).
		OnConflict().
		DoUpdate(Set(Users.Active, true)).
		Compile(Postgres()); err == nil {
		t.Fatal("expected DO UPDATE without target error")
	}
	if _, _, err := Insert(Users).
		Set(Users.Email.Set("a")).
		OnConflict(Users.Email).
		DoUpdate().
		Compile(Postgres()); err == nil {
		t.Fatal("expected empty DO UPDATE error")
	}
}
