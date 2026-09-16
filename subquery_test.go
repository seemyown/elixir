package elixir

import "testing"

func TestSubqueryIn(t *testing.T) {
	sub := Select(Orders.UserID).From(Orders).Where(Orders.Amount.Gt(100))
	q := Select(Users.Email).From(Users).Where(Users.ID.InQuery(sub))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email" FROM "users" WHERE ("users"."id" IN (SELECT "orders"."user_id" FROM "orders" WHERE ("orders"."amount" > $1)))`,
		[]any{float64(100)},
	)
}

func TestInValues(t *testing.T) {
	sql, args, err := CompilePredicate(Postgres(), Users.ID.In(1, 2, 3))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`("users"."id" IN ($1, $2, $3))`,
		[]any{int64(1), int64(2), int64(3)},
	)
}

func TestNotInQuery(t *testing.T) {
	sub := Select(Orders.UserID).From(Orders)
	sql, args, err := CompilePredicate(MySQL(), Users.ID.NotInQuery(sub))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"(`users`.`id` NOT IN (SELECT `orders`.`user_id` FROM `orders`))",
		nil,
	)
}

func TestScalarSubquery(t *testing.T) {
	maxID := Subquery[int64](Select(Max(Users.ID)).From(Users))
	q := Select(Users.Email).From(Users).Where(Users.ID.EqExpr(maxID))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email" FROM "users" WHERE ("users"."id" = (SELECT MAX("users"."id") FROM "users"))`,
		nil,
	)
}

func TestSubqueryFrom(t *testing.T) {
	sub := Select(Users.ID, Users.Email).
		From(Users).
		Where(Users.Active.Eq(true)).
		AsTable("u")

	id := Column[int64]{Table: "u", Name: "id"}
	email := Column[string]{Table: "u", Name: "email"}

	q := Select(id, email).From(sub).Where(id.Gt(10))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "u"."id", "u"."email" FROM (SELECT "users"."id", "users"."email" FROM "users" WHERE ("users"."active" = $1)) AS "u" WHERE ("u"."id" > $2)`,
		[]any{true, int64(10)},
	)
}

func TestExists(t *testing.T) {
	sub := Select(Orders.ID).From(Orders).Where(Orders.UserID.EqExpr(Users.ID))
	q := Select(Users.Email).From(Users).Where(Exists(sub))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email" FROM "users" WHERE (EXISTS (SELECT "orders"."id" FROM "orders" WHERE ("orders"."user_id" = "users"."id")))`,
		nil,
	)
}

func TestNotExists(t *testing.T) {
	sub := Select(Orders.ID).From(Orders).Where(Orders.UserID.Eq(1))
	sql, args, err := CompilePredicate(SQLite(), NotExists(sub))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`(NOT EXISTS (SELECT "orders"."id" FROM "orders" WHERE ("orders"."user_id" = ?)))`,
		[]any{int64(1)},
	)
}

func TestEmptyInError(t *testing.T) {
	p := Users.ID.In()
	if _, _, err := CompilePredicate(Postgres(), p); err == nil {
		t.Fatal("expected empty IN error")
	}
}
