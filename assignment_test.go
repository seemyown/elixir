package elixir

import "testing"

func TestAssignmentHelpers(t *testing.T) {
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

	a := Set(tbio.Email, "a@example.com")
	b := SetExpr(tbio.Email, Lower(tbio.Email))
	c := SetNull(tbio.Bio)
	q := Update(tbio).Set(a, b, c)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "email" = $1, "email" = LOWER("users"."email"), "bio" = NULL`,
		[]any{"a@example.com"},
	)
}
