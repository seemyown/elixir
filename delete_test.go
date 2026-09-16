package elixir

import "testing"

func TestDelete(t *testing.T) {
	q := Delete(Users).Where(Users.ID.Eq(1), Users.Active.Eq(false))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`DELETE FROM "users" WHERE (("users"."id" = $1) AND ("users"."active" = $2))`,
		[]any{int64(1), false},
	)
}

func TestDeleteAll(t *testing.T) {
	q := Delete(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `DELETE FROM "users"`, nil)
}

func TestDeleteReturning(t *testing.T) {
	q := Delete(Users).Where(Users.ID.Eq(1)).Returning(Users.ID)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`DELETE FROM "users" WHERE ("users"."id" = $1) RETURNING "users"."id"`,
		[]any{int64(1)},
	)
}

func TestDeleteWithDialect(t *testing.T) {
	sql, args, err := Delete(Users).Where(Users.ID.Eq(9)).WithDialect(MySQL()).ToSQL()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, "DELETE FROM `users` WHERE (`users`.`id` = ?)", []any{int64(9)})
}

func TestCTEWithDelete(t *testing.T) {
	q := With(
		CTE("victims", Select(Users.ID).From(Users).Where(Users.Active.Eq(false))),
	).Delete(Users).Where(Users.ID.InQuery(Select(Column[int64]{Name: "id"}).From(TableRef{Name: "victims"})))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "victims" AS (SELECT "users"."id" FROM "users" WHERE ("users"."active" = $1)) DELETE FROM "users" WHERE ("users"."id" IN (SELECT "id" FROM "victims"))`,
		[]any{false},
	)
}
