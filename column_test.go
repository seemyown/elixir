package elixir

import "testing"

func TestColumnComparisons(t *testing.T) {
	cases := []struct {
		name string
		p    Predicate
		sql  string
		args []any
	}{
		{"Ne", Users.Email.Ne("x"), `("users"."email" <> $1)`, []any{"x"}},
		{"NeExpr", Users.Email.NeExpr(Lower(Users.Email)), `("users"."email" <> LOWER("users"."email"))`, nil},
		{"Gte", Users.ID.Gte(10), `("users"."id" >= $1)`, []any{int64(10)}},
		{"GteExpr", Users.ID.GteExpr(Orders.UserID), `("users"."id" >= "orders"."user_id")`, nil},
		{"Lt", Users.ID.Lt(5), `("users"."id" < $1)`, []any{int64(5)}},
		{"LtExpr", Users.ID.LtExpr(Orders.UserID), `("users"."id" < "orders"."user_id")`, nil},
		{"Lte", Users.ID.Lte(5), `("users"."id" <= $1)`, []any{int64(5)}},
		{"LteExpr", Users.ID.LteExpr(Orders.UserID), `("users"."id" <= "orders"."user_id")`, nil},
		{"GtExpr", Users.ID.GtExpr(Orders.UserID), `("users"."id" > "orders"."user_id")`, nil},
		{"IsNotNull", Users.Bio.IsNotNull(), `("users"."bio" IS NOT NULL)`, nil},
		{"NotIn", Users.ID.NotIn(1, 2), `("users"."id" NOT IN ($1, $2))`, []any{int64(1), int64(2)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sql, args, err := CompilePredicate(Postgres(), tc.p)
			if err != nil {
				t.Fatal(err)
			}
			assertSQL(t, sql, args, tc.sql, tc.args)
		})
	}
}

func TestColumnAsAndOver(t *testing.T) {
	q := Select(Users.Email.As("e"), Users.ID.Over(WinOrderBy(Users.ID.Asc()))).From(Users)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email" AS "e", "users"."id" OVER (ORDER BY "users"."id" ASC) FROM "users"`,
		nil,
	)
}
