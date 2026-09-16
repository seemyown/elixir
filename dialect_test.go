package elixir

import "testing"

func TestMySQLDialect(t *testing.T) {
	query := Select(Users.ID).From(Users).Where(Users.ID.Eq(7))
	sql, args, err := query.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"SELECT `users`.`id` FROM `users` WHERE (`users`.`id` = ?)",
		[]any{int64(7)},
	)
}

func TestSQLiteDialect(t *testing.T) {
	query := Select(Users.ID).From(Users).Where(Users.Email.Eq("x"))
	sql, args, err := query.Compile(SQLite())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."id" FROM "users" WHERE ("users"."email" = ?)`,
		[]any{"x"},
	)
}

func TestMySQLBacktickEscape(t *testing.T) {
	col := Column[string]{Name: "na`me"}
	q := Select(col).From(TableRef{Name: "ta`ble"})
	sql, args, err := q.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, "SELECT `na``me` FROM `ta``ble`", nil)
}

func TestPostgresQuoteEscape(t *testing.T) {
	col := Column[string]{Name: `na"me`}
	q := Select(col).From(TableRef{Name: `ta"ble`})
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `SELECT "na""me" FROM "ta""ble"`, nil)
}
