package elixir

import "testing"

func TestCTE(t *testing.T) {
	active := TableRef{Name: "active_users"}
	activeID := Column[int64]{Table: "active_users", Name: "id"}

	q := With(
		CTE("active_users",
			Select(Users.ID).From(Users).Where(Users.Active.Eq(true)),
		),
	).
		Select(activeID).
		From(active).
		Where(activeID.Gt(5))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "active_users" AS (SELECT "users"."id" FROM "users" WHERE ("users"."active" = $1)) SELECT "active_users"."id" FROM "active_users" WHERE ("active_users"."id" > $2)`,
		[]any{true, int64(5)},
	)
}

func TestCTEWithUpdate(t *testing.T) {
	q := With(
		CTE("victims",
			Select(Users.ID).From(Users).Where(Users.Active.Eq(false)),
		),
	).
		Update(Users).
		Set(Set(Users.Email, "gone@example.com")).
		Where(Users.ID.InQuery(Select(Column[int64]{Name: "id"}).From(TableRef{Name: "victims"})))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "victims" AS (SELECT "users"."id" FROM "users" WHERE ("users"."active" = $1)) UPDATE "users" SET "email" = $2 WHERE ("users"."id" IN (SELECT "id" FROM "victims"))`,
		[]any{false, "gone@example.com"},
	)
}

func TestWithBuilderWithDialect(t *testing.T) {
	q := With(CTE("t", Select(Users.ID).From(Users))).
		WithDialect(Postgres()).
		Select(Column[int64]{Name: "id"}).
		From(TableRef{Name: "t"})
	sql, args, err := q.Compile()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "t" AS (SELECT "users"."id" FROM "users") SELECT "id" FROM "t"`,
		nil,
	)
}

func TestEmptyCTENameError(t *testing.T) {
	q := With(CTE("", Select(Users.ID).From(Users))).Select(Users.ID).From(Users)
	if _, _, err := q.Compile(Postgres()); err == nil {
		t.Fatal("expected CTE name error")
	}
}
