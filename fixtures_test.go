package elixir

import (
	"testing"
	"time"
)

type userTable struct {
	TableRef

	ID        Column[int64]
	Email     Column[string]
	Bio       NullColumn[string]
	Active    Column[bool]
	CreatedAt Column[time.Time]
}

type orderTable struct {
	TableRef

	ID     Column[int64]
	UserID Column[int64]
	Amount Column[float64]
}

var (
	Users = Bind(userTable{
		TableRef:  TableRef{Name: "users"},
		ID:        Column[int64]{Name: "id"},
		Email:     Column[string]{Name: "email"},
		Bio:       NullColumn[string]{Name: "bio"},
		Active:    Column[bool]{Name: "active"}.WithDefault(true),
		CreatedAt: Column[time.Time]{Name: "created_at"},
	})

	Orders = Bind(orderTable{
		TableRef: TableRef{Name: "orders"},
		ID:       Column[int64]{Name: "id"},
		UserID:   Column[int64]{Name: "user_id"},
		Amount:   Column[float64]{Name: "amount"},
	})
)

func assertSQL(t *testing.T, gotSQL string, gotArgs []any, wantSQL string, wantArgs []any) {
	t.Helper()
	if gotSQL != wantSQL {
		t.Fatalf("SQL\n got: %s\nwant: %s", gotSQL, wantSQL)
	}
	if len(gotArgs) != len(wantArgs) {
		t.Fatalf("args len got %d want %d\n got: %#v\nwant: %#v", len(gotArgs), len(wantArgs), gotArgs, wantArgs)
	}
	for i := range wantArgs {
		if gotArgs[i] != wantArgs[i] {
			t.Fatalf("args[%d] got %#v want %#v\nfull got: %#v\nfull want: %#v", i, gotArgs[i], wantArgs[i], gotArgs, wantArgs)
		}
	}
}
