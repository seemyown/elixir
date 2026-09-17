package elixir

import (
	"database/sql"
	"testing"
)

func TestNullHelpers(t *testing.T) {
	s := Some("hi")
	if !s.Valid || s.V != "hi" {
		t.Fatalf("Some: %+v", s)
	}
	if None[string]().Valid {
		t.Fatal("None should be invalid")
	}
	p := s.Ptr()
	if p == nil || *p != "hi" {
		t.Fatalf("Ptr: %v", p)
	}
	if None[string]().Ptr() != nil {
		t.Fatal("None.Ptr should be nil")
	}
	if FromPtr[string](nil).Valid {
		t.Fatal("FromPtr(nil) should be NULL")
	}
	v := "x"
	if n := FromPtr(&v); !n.Valid || n.V != "x" {
		t.Fatalf("FromPtr: %+v", n)
	}
	if Some("a").Or("b") != "a" || None[string]().Or("b") != "b" {
		t.Fatal("Or failed")
	}
	sqlN := sql.Null[string]{V: "z", Valid: true}
	if n := FromSQL(sqlN); !n.Valid || n.V != "z" || n.SQL() != sqlN {
		t.Fatalf("SQL roundtrip: %+v", n)
	}
}

func TestNullColumnSetPtrAndOpt(t *testing.T) {
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
	if tbio.Bio.Table != "users" {
		t.Fatalf("Bind should set NullColumn.Table, got %q", tbio.Bio.Table)
	}

	bio := "about"
	q := Insert(tbio).Set(
		tbio.Email.Set("a@example.com"),
		tbio.Bio.SetPtr(&bio),
	)
	gotSQL, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, gotSQL, args,
		`INSERT INTO "users" ("email", "bio") VALUES ($1, $2)`,
		[]any{"a@example.com", "about"},
	)

	q2 := Insert(tbio).Set(
		tbio.Email.Set("a@example.com"),
		tbio.Bio.SetPtr(nil),
	)
	gotSQL, args, err = q2.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, gotSQL, args,
		`INSERT INTO "users" ("email", "bio") VALUES ($1, NULL)`,
		[]any{"a@example.com"},
	)

	q3 := Insert(tbio).Set(
		tbio.Email.Set("a@example.com"),
		tbio.Bio.SetOpt(None[string]()),
	)
	gotSQL, args, err = q3.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, gotSQL, args,
		`INSERT INTO "users" ("email", "bio") VALUES ($1, NULL)`,
		[]any{"a@example.com"},
	)

	q4 := Insert(tbio).Set(
		tbio.Email.Set("a@example.com"),
		tbio.Bio.SetSQLNull(sql.Null[string]{V: "from-sql", Valid: true}),
	)
	gotSQL, args, err = q4.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, gotSQL, args,
		`INSERT INTO "users" ("email", "bio") VALUES ($1, $2)`,
		[]any{"a@example.com", "from-sql"},
	)
}

func TestNullColumnEqOpt(t *testing.T) {
	type bioTable struct {
		TableRef
		Bio NullColumn[string]
	}
	tbio := Bind(bioTable{
		TableRef: TableRef{Name: "users"},
		Bio:      NullColumn[string]{Name: "bio"},
	})

	sql, args, err := CompilePredicate(Postgres(), tbio.Bio.EqOpt(None[string]()))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `("users"."bio" IS NULL)`, nil)

	sql, args, err = CompilePredicate(Postgres(), tbio.Bio.EqOpt(Some("x")))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `("users"."bio" = $1)`, []any{"x"})

	sql, args, err = CompilePredicate(Postgres(), tbio.Bio.EqPtr(nil))
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args, `("users"."bio" IS NULL)`, nil)
}

func TestValueOpt(t *testing.T) {
	type bioTable struct {
		TableRef
		Bio NullColumn[string]
	}
	tbio := Bind(bioTable{
		TableRef: TableRef{Name: "users"},
		Bio:      NullColumn[string]{Name: "bio"},
	})
	q := Insert(tbio).
		Columns(tbio.Bio).
		Values(ValueOpt(None[string]())).
		Values(ValuePtr[string](nil))
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`INSERT INTO "users" ("bio") VALUES (NULL), (NULL)`,
		nil,
	)
}
