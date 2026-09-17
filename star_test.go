package elixir

import "testing"

func TestSelectAll(t *testing.T) {
	q := Select(All()).From(Users).Where(Users.ID.Eq(1))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT * FROM "users" WHERE ("users"."id" = $1)`,
		[]any{int64(1)},
	)
}

func TestSelectAllOf(t *testing.T) {
	q := Select(AllOf(Users), Orders.Amount).
		From(Users).
		Join(Orders, Users.ID.EqExpr(Orders.UserID))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users".*, "orders"."amount" FROM "users" JOIN "orders" ON ("users"."id" = "orders"."user_id")`,
		nil,
	)
}

func TestSelectAllOfAlias(t *testing.T) {
	u := As(Users, "u")
	q := Select(AllOf(u)).From(u)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "u".* FROM "users" AS "u"`,
		nil,
	)
}

func TestSelectAllMySQL(t *testing.T) {
	sql, args, err := Select(All()).From(Users).Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, "SELECT * FROM `users`", nil)
}
