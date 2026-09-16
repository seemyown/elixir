package elixir

import "testing"

func TestWindowPartitionOrder(t *testing.T) {
	q := Select(
		Users.Email,
		Count(Users.ID).Over(
			PartitionBy(Users.Email).OrderBy(Users.CreatedAt.Desc()),
		).As("cnt"),
	).From(Users)

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT "users"."email", COUNT("users"."id") OVER (PARTITION BY "users"."email" ORDER BY "users"."created_at" DESC) AS "cnt" FROM "users"`,
		nil,
	)
}

func TestWindowRowNumberFrame(t *testing.T) {
	q := Select(
		RowNumber().Over(
			WinOrderBy(Users.ID.Asc()).Rows(UnboundedPreceding(), CurrentRow()),
		).As("rn"),
	).From(Users)

	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT ROW_NUMBER() OVER (ORDER BY "users"."id" ASC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS "rn" FROM "users"`,
		nil,
	)
}

func TestWindowRankMySQL(t *testing.T) {
	q := Select(
		Rank().Over(PartitionBy(Users.Active).OrderBy(Users.ID.Desc())).As("r"),
	).From(Users)

	sql, args, err := q.Compile(MySQL())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		"SELECT RANK() OVER (PARTITION BY `users`.`active` ORDER BY `users`.`id` DESC) AS `r` FROM `users`",
		nil,
	)
}

func TestWindowRangeFrame(t *testing.T) {
	q := Select(
		Sum(Orders.Amount).Over(
			PartitionBy(Orders.UserID).
				OrderBy(Orders.ID.Asc()).
				Range(Preceding(1), Following(1)),
		),
	).From(Orders)

	sql, args, err := q.Compile(SQLite())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT SUM("orders"."amount") OVER (PARTITION BY "orders"."user_id" ORDER BY "orders"."id" ASC RANGE BETWEEN 1 PRECEDING AND 1 FOLLOWING) FROM "orders"`,
		nil,
	)
}

func TestDenseRank(t *testing.T) {
	q := Select(DenseRank().Over(WinOrderBy(Orders.Amount.Desc()))).From(Orders)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT DENSE_RANK() OVER (ORDER BY "orders"."amount" DESC) FROM "orders"`,
		nil,
	)
}

func TestUnboundedFollowing(t *testing.T) {
	q := Select(
		Sum(Orders.Amount).Over(
			WinOrderBy(Orders.ID.Asc()).Rows(CurrentRow(), UnboundedFollowing()),
		),
	).From(Orders)
	sql, args, err := q.Compile(Postgres())
	if err != nil {
		t.Fatal(err)
	}
	assertSQL(t, sql, args,
		`SELECT SUM("orders"."amount") OVER (ORDER BY "orders"."id" ASC ROWS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING) FROM "orders"`,
		nil,
	)
}
