package elixir

import "testing"

func TestREADMEAggregateHaving(t *testing.T) {
	query := Select(
		Users.Email,
		Count(Users.ID).As("total"),
	).
		From(Users).
		Where(
			Users.Active.Eq(true),
		).
		GroupBy(
			Users.Email,
		).
		Having(
			Count(Users.ID).Gt(5),
		)

	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}

	assertSQL(t, sql, args,
		`SELECT "users"."email", COUNT("users"."id") AS "total" FROM "users" WHERE ("users"."active" = $1) GROUP BY "users"."email" HAVING (COUNT("users"."id") > $2)`,
		[]any{true, int64(5)},
	)
}

func TestCountDistinct(t *testing.T) {
	query := Select(Count(Distinct(Users.Email))).From(Users)
	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT COUNT(DISTINCT "users"."email") FROM "users"`,
		nil,
	)
}

func TestAggregates(t *testing.T) {
	query := Select(
		Sum(Orders.Amount),
		Avg(Orders.Amount),
		Min(Orders.Amount),
		Max(Orders.Amount),
		CountAll(),
	).From(Orders)

	sql, args, err := query.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT SUM("orders"."amount"), AVG("orders"."amount"), MIN("orders"."amount"), MAX("orders"."amount"), COUNT(*) FROM "orders"`,
		nil,
	)
}
