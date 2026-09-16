package elixir

import "testing"

func TestAssignmentHelpers(t *testing.T) {
	a := Set(Users.Email, "a@example.com")
	b := SetExpr(Users.Email, Lower(Users.Email))
	c := SetNull(Users.Email)
	q := Update(Users).Set(a, b, c)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "email" = $1, "email" = LOWER("users"."email"), "email" = NULL`,
		[]any{"a@example.com"},
	)
}
