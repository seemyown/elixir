package elixir_test

import (
	"fmt"
	"time"

	"github.com/seemyown/elixir"
)

func ExampleEngine() {
	type userTable struct {
		elixir.TableRef
		ID elixir.Column[int64]
	}
	Users := elixir.Bind(userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Name: "id"},
	})

	e := elixir.New(elixir.Postgres())
	q := e.Select(Users.ID).From(Users).Where(Users.ID.Eq(1))
	sql, args, err := q.Compile()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "users"."id" FROM "users" WHERE ("users"."id" = $1)
	// [1]
}

func ExampleAll() {
	type userTable struct {
		elixir.TableRef
		ID elixir.Column[int64]
	}
	Users := elixir.Bind(userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Name: "id"},
	})

	sql, args, err := elixir.Select(elixir.All()).From(Users).Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT * FROM "users"
	// []
}

func ExampleSelect() {
	type userTable struct {
		elixir.TableRef

		ID     elixir.Column[int64]
		Email  elixir.Column[string]
		Active elixir.Column[bool]
	}

	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
		Email:    elixir.Column[string]{Table: "users", Name: "email"},
		Active:   elixir.Column[bool]{Table: "users", Name: "active"},
	}

	query := elixir.Select(
		Users.Email,
		elixir.Count(Users.ID).As("total"),
	).
		From(Users).
		Where(
			Users.Active.Eq(true),
		).
		GroupBy(
			Users.Email,
		).
		Having(
			elixir.Count(Users.ID).Gt(5),
		)

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "users"."email", COUNT("users"."id") AS "total" FROM "users" WHERE ("users"."active" = $1) GROUP BY "users"."email" HAVING (COUNT("users"."id") > $2)
	// [true 5]
}

func ExampleColumn_nameOnly() {
	type userTable struct {
		elixir.TableRef

		ID        elixir.Column[int64]
		Email     elixir.Column[string]
		Password  elixir.Column[string]
		CreatedAt elixir.Column[time.Time]
	}

	User := userTable{
		TableRef:  elixir.TableRef{Name: "users"},
		ID:        elixir.Column[int64]{Name: "id"},
		Email:     elixir.Column[string]{Name: "email"},
		Password:  elixir.Column[string]{Name: "password"},
		CreatedAt: elixir.Column[time.Time]{Name: "created_at"},
	}

	query := elixir.Select(User.ID, User.Email).
		From(User).
		Where(User.ID.Eq(1))

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "id", "email" FROM "users" WHERE ("id" = $1)
	// [1]
}

func ExampleInsert() {
	type userTable struct {
		elixir.TableRef
		Email  elixir.Column[string]
		Active elixir.Column[bool]
	}
	Users := elixir.Bind(userTable{
		TableRef: elixir.TableRef{Name: "users"},
		Email:    elixir.Column[string]{Name: "email"},
		Active:   elixir.Column[bool]{Name: "active"},
	})

	query := elixir.Insert(Users).Set(
		Users.Email.Set("a@example.com"),
		Users.Active.Set(true),
	)

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// INSERT INTO "users" ("email", "active") VALUES ($1, $2)
	// [a@example.com true]
}

func ExampleUpdate() {
	type userTable struct {
		elixir.TableRef
		ID     elixir.Column[int64]
		Email  elixir.Column[string]
		Active elixir.Column[bool]
	}
	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
		Email:    elixir.Column[string]{Table: "users", Name: "email"},
		Active:   elixir.Column[bool]{Table: "users", Name: "active"},
	}

	query := elixir.Update(Users).
		Set(
			elixir.Set(Users.Email, "new@example.com"),
			elixir.Set(Users.Active, false),
		).
		Where(Users.ID.Eq(1))

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// UPDATE "users" SET "email" = $1, "active" = $2 WHERE ("users"."id" = $3)
	// [new@example.com false 1]
}

func ExampleDelete() {
	type userTable struct {
		elixir.TableRef
		ID elixir.Column[int64]
	}
	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
	}

	query := elixir.Delete(Users).Where(Users.ID.Eq(1))
	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// DELETE FROM "users" WHERE ("users"."id" = $1)
	// [1]
}

func ExampleSubquery() {
	type userTable struct {
		elixir.TableRef
		ID    elixir.Column[int64]
		Email elixir.Column[string]
	}
	type orderTable struct {
		elixir.TableRef
		UserID elixir.Column[int64]
		Amount elixir.Column[float64]
	}
	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
		Email:    elixir.Column[string]{Table: "users", Name: "email"},
	}
	Orders := orderTable{
		TableRef: elixir.TableRef{Name: "orders"},
		UserID:   elixir.Column[int64]{Table: "orders", Name: "user_id"},
		Amount:   elixir.Column[float64]{Table: "orders", Name: "amount"},
	}

	sub := elixir.Select(Orders.UserID).From(Orders).Where(Orders.Amount.Gt(100))
	query := elixir.Select(Users.Email).From(Users).Where(Users.ID.InQuery(sub))

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "users"."email" FROM "users" WHERE ("users"."id" IN (SELECT "orders"."user_id" FROM "orders" WHERE ("orders"."amount" > $1)))
	// [100]
}

func ExampleWith() {
	type userTable struct {
		elixir.TableRef
		ID     elixir.Column[int64]
		Active elixir.Column[bool]
	}
	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
		Active:   elixir.Column[bool]{Table: "users", Name: "active"},
	}

	active := elixir.TableRef{Name: "active_users"}
	activeID := elixir.Column[int64]{Table: "active_users", Name: "id"}

	query := elixir.With(
		elixir.CTE("active_users",
			elixir.Select(Users.ID).From(Users).Where(Users.Active.Eq(true)),
		),
	).
		Select(activeID).
		From(active)

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// WITH "active_users" AS (SELECT "users"."id" FROM "users" WHERE ("users"."active" = $1)) SELECT "active_users"."id" FROM "active_users"
	// [true]
}

func ExampleExpr_Over() {
	type userTable struct {
		elixir.TableRef
		ID    elixir.Column[int64]
		Email elixir.Column[string]
	}
	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
		Email:    elixir.Column[string]{Table: "users", Name: "email"},
	}

	query := elixir.Select(
		Users.Email,
		elixir.Count(Users.ID).Over(
			elixir.PartitionBy(Users.Email).OrderBy(Users.ID.Asc()),
		).As("cnt"),
	).From(Users)

	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "users"."email", COUNT("users"."id") OVER (PARTITION BY "users"."email" ORDER BY "users"."id" ASC) AS "cnt" FROM "users"
	// []
}

func ExampleCase() {
	type userTable struct {
		elixir.TableRef
		ID     elixir.Column[int64]
		Active elixir.Column[bool]
	}
	Users := userTable{
		TableRef: elixir.TableRef{Name: "users"},
		ID:       elixir.Column[int64]{Table: "users", Name: "id"},
		Active:   elixir.Column[bool]{Table: "users", Name: "active"},
	}

	label := elixir.Case[string]().
		When(Users.Active.Eq(true), elixir.Value("yes")).
		Else(elixir.Value("no"))

	query := elixir.Select(Users.ID, label.As("label")).From(Users)
	sql, args, err := query.Compile(elixir.Postgres())
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// SELECT "users"."id", CASE WHEN ("users"."active" = $1) THEN $2 ELSE $3 END AS "label" FROM "users"
	// [true yes no]
}
