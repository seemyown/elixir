package elixir

import "testing"

func TestBindFillsColumnTable(t *testing.T) {
	if Users.ID.Table != "users" || Users.Email.Table != "users" {
		t.Fatalf("Bind did not set Table: id=%q email=%q", Users.ID.Table, Users.Email.Table)
	}
}

func TestAsRebindsAlias(t *testing.T) {
	u := As(Users, "u")
	if u.Alias != "u" {
		t.Fatalf("alias: got %q", u.Alias)
	}
	if u.ID.Table != "u" || u.Email.Table != "u" {
		t.Fatalf("As did not rebind columns: id=%q email=%q", u.ID.Table, u.Email.Table)
	}
	sql, args, err := Select(u.ID).From(u).Where(u.ID.Eq(1)).Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "u"."id" FROM "users" AS "u" WHERE ("u"."id" = $1)`,
		[]any{int64(1)},
	)
}

func TestTableAlias(t *testing.T) {
	u := TableRef{Name: "users", Alias: "u"}
	id := Column[int64]{Table: "u", Name: "id"}
	query := Select(id).From(u).Where(id.Eq(1))
	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "u"."id" FROM "users" AS "u" WHERE ("u"."id" = $1)`,
		[]any{int64(1)},
	)
}

func TestTableRefAsAndTableName(t *testing.T) {
	u := TableRef{Name: "users"}.As("u")
	if u.TableName() != "users" || u.Alias != "u" {
		t.Fatalf("got %+v", u)
	}
}

func TestBindPreservesManualTable(t *testing.T) {
	type tdef struct {
		TableRef
		ID Column[int64]
	}
	u := Bind(tdef{
		TableRef: TableRef{Name: "users"},
		ID:       Column[int64]{Table: "custom", Name: "id"},
	})
	if u.ID.Table != "custom" {
		t.Fatalf("Bind overwrote manual Table: %q", u.ID.Table)
	}
}

func TestBindWithAliasQualifier(t *testing.T) {
	type tdef struct {
		TableRef
		ID Column[int64]
	}
	u := Bind(tdef{
		TableRef: TableRef{Name: "users", Alias: "u"},
		ID:       Column[int64]{Name: "id"},
	})
	if u.ID.Table != "u" {
		t.Fatalf("expected alias qualifier, got %q", u.ID.Table)
	}
}
