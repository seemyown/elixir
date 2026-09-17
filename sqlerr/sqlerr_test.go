package sqlerr_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/seemyown/elixir/sqlerr"
)

type sqlStateErr struct {
	state string
	msg   string
}

func (e sqlStateErr) Error() string    { return e.msg }
func (e sqlStateErr) SQLState() string { return e.state }

type mysqlErr struct {
	num uint16
	msg string
}

func (e mysqlErr) Error() string  { return e.msg }
func (e mysqlErr) Number() uint16 { return e.num }

type sqliteErr struct {
	code int
	msg  string
}

func (e sqliteErr) Error() string { return e.msg }
func (e sqliteErr) Code() int     { return e.code }

type pqCodeErr struct {
	code string
	msg  string
}

func (e pqCodeErr) Error() string { return e.msg }
func (e pqCodeErr) Code() string  { return e.code }

func TestIsNoRows(t *testing.T) {
	if !sqlerr.IsNoRows(sql.ErrNoRows) {
		t.Fatal("expected sql.ErrNoRows")
	}
	if !sqlerr.IsNoRows(fmt.Errorf("wrap: %w", sql.ErrNoRows)) {
		t.Fatal("expected wrapped sql.ErrNoRows")
	}
	if sqlerr.IsNoRows(errors.New("other")) {
		t.Fatal("unexpected match")
	}
}

func TestIsUniqueViolation_SQLState(t *testing.T) {
	err := sqlStateErr{state: sqlerr.PGUniqueViolation, msg: "duplicate key value violates unique constraint"}
	if !sqlerr.IsUniqueViolation(err) {
		t.Fatal("expected unique violation from SQLSTATE 23505")
	}
	if sqlerr.Code(err) != sqlerr.PGUniqueViolation {
		t.Fatalf("Code = %q", sqlerr.Code(err))
	}
}

func TestIsUniqueViolation_MySQL(t *testing.T) {
	err := mysqlErr{num: sqlerr.MySQLDupEntry, msg: "Error 1062: Duplicate entry"}
	if !sqlerr.IsUniqueViolation(err) {
		t.Fatal("expected MySQL 1062")
	}
	if sqlerr.Code(err) != "1062" {
		t.Fatalf("Code = %q", sqlerr.Code(err))
	}
}

func TestIsUniqueViolation_SQLite(t *testing.T) {
	err := sqliteErr{code: sqlerr.SQLiteConstraintUnique, msg: "UNIQUE constraint failed: users.email"}
	if !sqlerr.IsUniqueViolation(err) {
		t.Fatal("expected SQLite unique constraint code")
	}
}

func TestIsUniqueViolation_Message(t *testing.T) {
	err := errors.New("pq: duplicate key value violates unique constraint \"users_email_key\"")
	if !sqlerr.IsUniqueViolation(err) {
		t.Fatal("expected message fallback")
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	cases := []error{
		sqlStateErr{state: sqlerr.PGForeignKeyViolation, msg: "foreign_key_violation"},
		mysqlErr{num: sqlerr.MySQLNoReferencedRow, msg: "Cannot add or update a child row"},
		sqliteErr{code: sqlerr.SQLiteConstraintForeignKey, msg: "FOREIGN KEY constraint failed"},
		errors.New("FOREIGN KEY constraint failed"),
	}
	for _, err := range cases {
		if !sqlerr.IsForeignKeyViolation(err) {
			t.Fatalf("expected FK violation for %v", err)
		}
	}
}

func TestIsNotNullViolation(t *testing.T) {
	cases := []error{
		sqlStateErr{state: sqlerr.PGNotNullViolation, msg: "not_null_violation"},
		mysqlErr{num: sqlerr.MySQLBadNull, msg: "Column 'email' cannot be null"},
		sqliteErr{code: sqlerr.SQLiteConstraintNotNull, msg: "NOT NULL constraint failed"},
		errors.New("NOT NULL constraint failed: users.email"),
	}
	for _, err := range cases {
		if !sqlerr.IsNotNullViolation(err) {
			t.Fatalf("expected NOT NULL violation for %v", err)
		}
	}
}

func TestIsCheckViolation(t *testing.T) {
	cases := []error{
		sqlStateErr{state: sqlerr.PGCheckViolation, msg: "check_violation"},
		mysqlErr{num: sqlerr.MySQLCheckConstraint, msg: "Check constraint violated"},
		sqliteErr{code: sqlerr.SQLiteConstraintCheck, msg: "CHECK constraint failed"},
		errors.New("CHECK constraint failed: users"),
	}
	for _, err := range cases {
		if !sqlerr.IsCheckViolation(err) {
			t.Fatalf("expected CHECK violation for %v", err)
		}
	}
}

func TestIsSerializationFailure(t *testing.T) {
	cases := []error{
		sqlStateErr{state: sqlerr.PGSerializationFailure, msg: "could not serialize access"},
		sqlStateErr{state: sqlerr.PGDeadlockDetected, msg: "deadlock detected"},
		mysqlErr{num: sqlerr.MySQLLockDeadlock, msg: "Deadlock found when trying to get lock"},
		mysqlErr{num: sqlerr.MySQLLockWaitTimeout, msg: "Lock wait timeout exceeded"},
		sqliteErr{code: sqlerr.SQLiteBusy, msg: "database is locked"},
		errors.New("ERROR: deadlock detected"),
	}
	for _, err := range cases {
		if !sqlerr.IsSerializationFailure(err) {
			t.Fatalf("expected serialization/deadlock for %v", err)
		}
	}
}

func TestCode_Unwrap(t *testing.T) {
	inner := sqlStateErr{state: sqlerr.PGUniqueViolation, msg: "unique"}
	wrapped := fmt.Errorf("insert failed: %w", inner)
	if sqlerr.Code(wrapped) != sqlerr.PGUniqueViolation {
		t.Fatalf("Code = %q", sqlerr.Code(wrapped))
	}
	if !sqlerr.IsUniqueViolation(wrapped) {
		t.Fatal("expected unique on wrapped error")
	}
}

func TestCode_PQStyle(t *testing.T) {
	err := pqCodeErr{code: sqlerr.PGCheckViolation, msg: "new row violates check constraint"}
	if sqlerr.Code(err) != sqlerr.PGCheckViolation {
		t.Fatalf("Code = %q", sqlerr.Code(err))
	}
	if !sqlerr.IsCheckViolation(err) {
		t.Fatal("expected check via Code()")
	}
}

func TestNegative(t *testing.T) {
	err := errors.New("connection refused")
	if sqlerr.IsUniqueViolation(err) ||
		sqlerr.IsForeignKeyViolation(err) ||
		sqlerr.IsNotNullViolation(err) ||
		sqlerr.IsCheckViolation(err) ||
		sqlerr.IsSerializationFailure(err) {
		t.Fatal("unexpected classification")
	}
	if sqlerr.Code(err) != "" {
		t.Fatalf("Code = %q, want empty", sqlerr.Code(err))
	}
	if sqlerr.IsUniqueViolation(nil) {
		t.Fatal("nil should be false")
	}
}
