package elixir

import "testing"

func TestValue(t *testing.T) {
	q := Select(Value(42)).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `SELECT $1 FROM "users"`, []any{42})
}
