package elixir

import "testing"

func TestLowerEq(t *testing.T) {
	sql, args, err := CompilePredicate(Postgres(), Lower(Users.Email).Eq("foo@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`(LOWER("users"."email") = $1)`,
		[]any{"foo@example.com"},
	)
}

func TestCoalesceUpper(t *testing.T) {
	query := Select(
		Upper(Coalesce(Users.Email, Value("unknown"))),
	).From(Users)

	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT UPPER(COALESCE("users"."email", $1)) FROM "users"`,
		[]any{"unknown"},
	)
}

func TestFunc(t *testing.T) {
	q := Select(Func[string]("TRIM", Users.Email)).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `SELECT TRIM("users"."email") FROM "users"`, nil)
}
