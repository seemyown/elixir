package elixir

import "testing"

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
		Bio   Column[string]
	}
	tbio := Bind(bioTable{
		TableRef: TableRef{Name: "users"},
		Email:    Column[string]{Name: "email"},
		Bio:      Column[string]{Name: "bio", Nullable: true},
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
