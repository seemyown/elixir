package elixir

import "testing"

func TestEngineCompile(t *testing.T) {
	e := New(Postgres())
	q := Select(Users.ID).From(Users).Where(Users.ID.Eq(1))
	sql, args, err := e.Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."id" FROM "users" WHERE ("users"."id" = $1)`,
		[]any{int64(1)},
	)
}

func TestEngineSelectCompileNoArg(t *testing.T) {
	e := New(Postgres())
	q := e.Select(Users.Email).From(Users).Where(Users.Active.Eq(true))
	sql, args, err := q.Compile()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email" FROM "users" WHERE ("users"."active" = $1)`,
		[]any{true},
	)
}

func TestEngineSelectDialectOverride(t *testing.T) {
	e := New(Postgres())
	q := e.Select(Users.ID).From(Users).Where(Users.ID.Eq(7))
	sql, args, err := q.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"SELECT `users`.`id` FROM `users` WHERE (`users`.`id` = ?)",
		[]any{int64(7)},
	)
}

func TestEngineInsertUpdateDelete(t *testing.T) {
	e := WithDialect(SQLite())

	ins, args, err := e.Insert(Users).Set(Users.Email.Set("a@example.com")).Compile()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, ins, args,
		`INSERT INTO "users" ("email") VALUES (?)`,
		[]any{"a@example.com"},
	)

	upd, args, err := e.Update(Users).Set(Set(Users.Active, false)).Where(Users.ID.Eq(1)).Compile()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, upd, args,
		`UPDATE "users" SET "active" = ? WHERE ("users"."id" = ?)`,
		[]any{false, int64(1)},
	)

	del, args, err := e.Delete(Users).Where(Users.ID.Eq(2)).Compile()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, del, args,
		`DELETE FROM "users" WHERE ("users"."id" = ?)`,
		[]any{int64(2)},
	)
}

func TestEngineWith(t *testing.T) {
	e := New(Postgres())
	active := TableRef{Name: "active_users"}
	activeID := Column[int64]{Table: "active_users", Name: "id"}

	q := e.With(
		CTE("active_users", Select(Users.ID).From(Users).Where(Users.Active.Eq(true))),
	).Select(activeID).From(active)

	sql, args, err := q.Compile()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`WITH "active_users" AS (SELECT "users"."id" FROM "users" WHERE ("users"."active" = $1)) SELECT "active_users"."id" FROM "active_users"`,
		[]any{true},
	)
}

func TestEngineNilDialect(t *testing.T) {
	e := New(nil)
	if _, _, err := e.Compile(Select(Users.ID).From(Users)); err == nil {
		t.Fatal("expected nil dialect error")
	}
	if e.Dialect() != nil {
		t.Fatal("expected nil Dialect()")
	}
	var nilEng *Engine
	if nilEng.Dialect() != nil {
		t.Fatal("expected nil Engine.Dialect()")
	}
}

func TestCompileWithoutDialect(t *testing.T) {
	q := Select(Users.ID).From(Users)
	if _, _, err := q.Compile(); err == nil {
		t.Fatal("expected missing dialect error")
	}
	if _, _, err := q.Compile(nil); err == nil {
		t.Fatal("expected nil dialect error")
	}
	if _, _, err := q.Compile(Postgres(), MySQL()); err == nil {
		t.Fatal("expected too many dialects error")
	}
}

func TestWithDialectOnBuilder(t *testing.T) {
	q := Select(Users.ID).From(Users).Where(Users.ID.Eq(1)).WithDialect(Postgres())
	sql, args, err := q.ToSQL()
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."id" FROM "users" WHERE ("users"."id" = $1)`,
		[]any{int64(1)},
	)
}
