package elixir

import "testing"

func TestCaseSearched(t *testing.T) {
	expr := Case[string]().
		When(Users.Active.Eq(true), Value("yes")).
		When(Users.ID.Gt(100), Value("vip")).
		Else(Value("no"))

	q := Select(expr.As("label")).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT CASE WHEN ("users"."active" = $1) THEN $2 WHEN ("users"."id" > $3) THEN $4 ELSE $5 END AS "label" FROM "users"`,
		[]any{true, "yes", int64(100), "vip", "no"},
	)
}

func TestCaseSimple(t *testing.T) {
	expr := CaseOn[bool, string](Users.Active).
		When(Value(true), Value("active")).
		When(Value(false), Value("inactive")).
		End()

	q := Select(expr).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT CASE "users"."active" WHEN $1 THEN $2 WHEN $3 THEN $4 END FROM "users"`,
		[]any{true, "active", false, "inactive"},
	)
}

func TestCaseSimpleElse(t *testing.T) {
	expr := CaseOn[bool, string](Users.Active).
		When(Value(true), Value("active")).
		Else(Value("other"))
	q := Select(expr).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT CASE "users"."active" WHEN $1 THEN $2 ELSE $3 END FROM "users"`,
		[]any{true, "active", "other"},
	)
}

func TestCaseInWhere(t *testing.T) {
	label := Case[string]().
		When(Users.Active.Eq(true), Value("a")).
		Else(Value("b"))

	q := Select(Users.ID).From(Users).Where(label.Eq("a"))
	sql, args, err := q.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"SELECT `users`.`id` FROM `users` WHERE (CASE WHEN (`users`.`active` = ?) THEN ? ELSE ? END = ?)",
		[]any{true, "a", "b", "a"},
	)
}

func TestCaseEndWithoutElse(t *testing.T) {
	expr := Case[string]().When(Users.Active.Eq(true), Value("yes")).End()
	q := Select(expr).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT CASE WHEN ("users"."active" = $1) THEN $2 END FROM "users"`,
		[]any{true, "yes"},
	)
}
