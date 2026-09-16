package elixir

import "testing"

func TestExprComparisons(t *testing.T) {
	e := Lower(Users.Email)
	sql, args, err := CompilePredicate(Postgres(), e.Ne("x"))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `(LOWER("users"."email") <> $1)`, []any{"x"})

	sql, args, err = CompilePredicate(Postgres(), e.Gte("a"))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `(LOWER("users"."email") >= $1)`, []any{"a"})

	sql, args, err = CompilePredicate(Postgres(), e.Lt("z"))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `(LOWER("users"."email") < $1)`, []any{"z"})

	sql, args, err = CompilePredicate(Postgres(), e.Lte("z"))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `(LOWER("users"."email") <= $1)`, []any{"z"})

	sql, args, err = CompilePredicate(Postgres(), e.IsNotNull())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `(LOWER("users"."email") IS NOT NULL)`, nil)
}
