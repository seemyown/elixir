package elixir

import "testing"

func TestREADMESelectWhere(t *testing.T) {
	id := int64(1)
	email := "a@example.com"

	query := Select(
		Users.ID,
		Users.Email,
	).
		From(Users).
		Where(
			Users.ID.Eq(id),
			Users.Email.Eq(email),
		)

	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}

	assertSQL(t, sql, args,
		`SELECT "users"."id", "users"."email" FROM "users" WHERE (("users"."id" = $1) AND ("users"."email" = $2))`,
		[]any{int64(1), "a@example.com"},
	)
}

func TestNameOnlyColumn(t *testing.T) {
	type tdef struct {
		TableRef
		ID    Column[int64]
		Email Column[string]
	}
	u := tdef{
		TableRef: TableRef{Name: "users"},
		ID:       Column[int64]{Name: "id"},
		Email:    Column[string]{Name: "email"},
	}

	query := Select(u.ID, u.Email).From(u).Where(u.ID.Eq(1))
	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "id", "email" FROM "users" WHERE ("id" = $1)`,
		[]any{int64(1)},
	)
}

func TestOrderLimitOffset(t *testing.T) {
	query := Select(Users.ID).
		From(Users).
		OrderBy(Users.CreatedAt.Desc(), Users.ID.Asc()).
		Limit(10).
		Offset(20)

	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."id" FROM "users" ORDER BY "users"."created_at" DESC, "users"."id" ASC LIMIT $1 OFFSET $2`,
		[]any{10, 20},
	)
}

func TestJoin(t *testing.T) {
	query := Select(Users.Email, Orders.Amount).
		From(Users).
		Join(Orders, Orders.UserID.EqExpr(Users.ID)).
		Where(Users.Active.Eq(true))

	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email", "orders"."amount" FROM "users" JOIN "orders" ON ("orders"."user_id" = "users"."id") WHERE ("users"."active" = $1)`,
		[]any{true},
	)
}

func TestLeftRightFullJoin(t *testing.T) {
	q := Select(Users.ID).From(Users).
		LeftJoin(Orders, Orders.UserID.EqExpr(Users.ID)).
		RightJoin(Orders, Orders.UserID.EqExpr(Users.ID)).
		FullJoin(Orders, Orders.UserID.EqExpr(Users.ID))
	sql, _, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	if sql == "" {
		t.Fatal("empty sql")
	}
}

func TestSelectCompileErrors(t *testing.T) {
	if _, _, err := Select(Users.ID).Compile(Postgres()); err == nil {
		t.Fatal("expected FROM error")
	}
	if _, _, err := Select().From(Users).Compile(Postgres()); err == nil {
		t.Fatal("expected columns error")
	}
	if _, _, err := Select(Users.ID).From(Users).Compile(nil); err == nil {
		t.Fatal("expected nil dialect error")
	}
}

func TestSelectToSQL(t *testing.T) {
	sql, args, err := Select(Users.ID).From(Users).ToSQL(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `SELECT "users"."id" FROM "users"`, nil)
}
