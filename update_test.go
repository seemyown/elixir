package elixir

import "testing"

func TestUpdate(t *testing.T) {
	q := Update(Users).
		Set(
			Set(Users.Email, "new@example.com"),
			Set(Users.Active, false),
		).
		Where(Users.ID.Eq(1))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "email" = $1, "active" = $2 WHERE ("users"."id" = $3)`,
		[]any{"new@example.com", false, int64(1)},
	)
}

func TestUpdateSetExpr(t *testing.T) {
	q := Update(Users).
		Set(SetExpr(Users.Email, Lower(Users.Email))).
		Where(Users.ID.Eq(2))

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "email" = LOWER("users"."email") WHERE ("users"."id" = $1)`,
		[]any{int64(2)},
	)
}

func TestUpdateColumnSetExpr(t *testing.T) {
	q := Update(Users).
		Set(Users.Email.SetExpr(Upper(Users.Email))).
		Where(Users.ID.Eq(2))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "email" = UPPER("users"."email") WHERE ("users"."id" = $1)`,
		[]any{int64(2)},
	)
}

func TestUpdateSQLite(t *testing.T) {
	q := Update(Users).Set(Set(Users.Active, true)).Where(Users.ID.Eq(3))
	sql, args, err := q.Compile(SQLite())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "active" = ? WHERE ("users"."id" = ?)`,
		[]any{true, int64(3)},
	)
}

func TestUpdateErrors(t *testing.T) {
	if _, _, err := Update(Users).Where(Users.ID.Eq(1)).Compile(Postgres()); err == nil {
		t.Fatal("expected SET error")
	}
}

func TestUpdateReturning(t *testing.T) {
	q := Update(Users).
		Set(Set(Users.Active, false)).
		Where(Users.ID.Eq(1)).
		Returning(Users.ID)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "active" = $1 WHERE ("users"."id" = $2) RETURNING "users"."id"`,
		[]any{false, int64(1)},
	)
}

func TestUpdateSetNull(t *testing.T) {
	type bioTable struct {
		TableRef
		ID  Column[int64]
		Bio Column[string]
	}
	tbio := Bind(bioTable{
		TableRef: TableRef{Name: "users"},
		ID:       Column[int64]{Name: "id"},
		Bio:      Column[string]{Name: "bio", Nullable: true},
	})
	q := Update(tbio).Set(SetNull(tbio.Bio)).Where(tbio.ID.Eq(1))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`UPDATE "users" SET "bio" = NULL WHERE ("users"."id" = $1)`,
		[]any{int64(1)},
	)
}

func TestUpdateWithDialect(t *testing.T) {
	sql, args, err := Update(Users).
		Set(Set(Users.Active, true)).
		WithDialect(Postgres()).
		ToSQL()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `UPDATE "users" SET "active" = $1`, []any{true})
}
