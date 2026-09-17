package elixir

import "testing"

func TestAndOrPredicates(t *testing.T) {
	p := And(
		Users.Active.Eq(true),
		Users.ID.Gt(100),
	)
	sql, args, err := CompilePredicate(Postgres(), p)
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`(("users"."active" = $1) AND ("users"."id" > $2))`,
		[]any{true, int64(100)},
	)

	p = Or(
		Users.Email.Eq("a@example.com"),
		Users.Email.Eq("b@example.com"),
	)
	sql, args, err = CompilePredicate(Postgres(), p)
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`(("users"."email" = $1) OR ("users"."email" = $2))`,
		[]any{"a@example.com", "b@example.com"},
	)
}

func TestNotPredicate(t *testing.T) {
	sql, args, err := CompilePredicate(Postgres(), Not(Users.Active.Eq(true)))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`(NOT ("users"."active" = $1))`,
		[]any{true},
	)
}

func TestAndOrEmptySingle(t *testing.T) {
	sql, args, err := CompilePredicate(Postgres(), And())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `$1`, []any{true})

	sql, args, err = CompilePredicate(Postgres(), Or())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `$1`, []any{false})

	p := Users.Active.Eq(true)
	sql, args, err = CompilePredicate(Postgres(), And(p))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `("users"."active" = $1)`, []any{true})

	sql, args, err = CompilePredicate(Postgres(), Or(p))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `("users"."active" = $1)`, []any{true})
}

func TestIsNull(t *testing.T) {
	sql, args, err := CompilePredicate(Postgres(), Users.Bio.IsNull())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `("users"."bio" IS NULL)`, nil)
}

func TestCompilePredicateNilDialect(t *testing.T) {
	if _, _, err := CompilePredicate(nil, Users.ID.Eq(1)); err == nil {
		t.Fatal("expected nil dialect error")
	}
}
