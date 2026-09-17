// Package sqlerr classifies common database errors across Postgres, MySQL, and SQLite.
//
// Detection is driver-agnostic: prefer SQLSTATE via interface{ SQLState() string },
// then soft type assertions for common driver error shapes, then message/code
// string fallbacks. Only the standard library is required; drivers are never imported.
//
// Documented codes:
//
//	Postgres SQLSTATE:
//	  23505 unique_violation
//	  23503 foreign_key_violation
//	  23502 not_null_violation
//	  23514 check_violation
//	  40001 serialization_failure
//	  40P01 deadlock_detected
//
//	MySQL errno:
//	  1062 ER_DUP_ENTRY
//	  1451 ER_ROW_IS_REFERENCED_2 / 1452 ER_NO_REFERENCED_ROW_2
//	  1048 ER_BAD_NULL_ERROR
//	  3819 ER_CHECK_CONSTRAINT_VIOLATED
//	  1213 ER_LOCK_DEADLOCK
//	  1205 ER_LOCK_WAIT_TIMEOUT (treated as serialization/contention)
//
//	SQLite:
//	  SQLITE_CONSTRAINT (19) and extended codes:
//	    SQLITE_CONSTRAINT_UNIQUE (2067), SQLITE_CONSTRAINT_PRIMARYKEY (1555)
//	    SQLITE_CONSTRAINT_FOREIGNKEY (787), SQLITE_CONSTRAINT_NOTNULL (1299)
//	    SQLITE_CONSTRAINT_CHECK (275)
//	  SQLITE_BUSY (5), SQLITE_LOCKED (6) for contention / deadlock-like failures
package sqlerr

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Postgres SQLSTATE values used by this package.
const (
	PGUniqueViolation      = "23505"
	PGForeignKeyViolation  = "23503"
	PGNotNullViolation     = "23502"
	PGCheckViolation       = "23514"
	PGSerializationFailure = "40001"
	PGDeadlockDetected     = "40P01"
)

// MySQL errno values used by this package.
const (
	MySQLDupEntry        = 1062
	MySQLRowIsReferenced = 1451
	MySQLNoReferencedRow = 1452
	MySQLBadNull         = 1048
	MySQLCheckConstraint = 3819
	MySQLLockDeadlock    = 1213
	MySQLLockWaitTimeout = 1205
)

// SQLite result / extended constraint codes used by this package.
const (
	SQLiteBusy                 = 5
	SQLiteLocked               = 6
	SQLiteConstraint           = 19
	SQLiteConstraintCheck      = 275
	SQLiteConstraintForeignKey = 787
	SQLiteConstraintNotNull    = 1299
	SQLiteConstraintPrimaryKey = 1555
	SQLiteConstraintUnique     = 2067
)

// IsNoRows reports whether err is or wraps database/sql.ErrNoRows.
func IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// IsUniqueViolation reports duplicate-key / unique-constraint failures.
func IsUniqueViolation(err error) bool {
	return match(err, classUnique)
}

// IsForeignKeyViolation reports foreign-key constraint failures.
func IsForeignKeyViolation(err error) bool {
	return match(err, classFK)
}

// IsNotNullViolation reports NOT NULL constraint failures.
func IsNotNullViolation(err error) bool {
	return match(err, classNotNull)
}

// IsCheckViolation reports CHECK constraint failures.
func IsCheckViolation(err error) bool {
	return match(err, classCheck)
}

// IsSerializationFailure reports serialization failures and common deadlocks /
// lock contention that applications typically retry.
func IsSerializationFailure(err error) bool {
	return match(err, classSerialization)
}

// Code returns a SQLSTATE (or best-effort code string) when available.
// For MySQL errno it returns the decimal number as a string (e.g. "1062").
// For SQLite it returns the numeric code as a string when present.
// An empty string means no code could be extracted.
func Code(err error) string {
	for e := err; e != nil; e = errors.Unwrap(e) {
		if s := codeOf(e); s != "" {
			return s
		}
	}
	return ""
}

type errorClass int

const (
	classUnique errorClass = iota
	classFK
	classNotNull
	classCheck
	classSerialization
)

func match(err error, class errorClass) bool {
	if err == nil {
		return false
	}
	for e := err; e != nil; e = errors.Unwrap(e) {
		if classify(e) == class {
			return true
		}
	}
	// Fallback: whole-chain message scan (covers wrapped fmt.Errorf without Unwrap of inner codes).
	return classifyMessage(err.Error(), class)
}

func classify(err error) errorClass {
	if c, ok := classifyCode(codeOf(err)); ok {
		return c
	}
	if c, ok := classifyMessageOK(err.Error()); ok {
		return c
	}
	return -1
}

func codeOf(err error) string {
	type sqlStater interface{ SQLState() string }
	if e, ok := err.(sqlStater); ok {
		if s := strings.TrimSpace(e.SQLState()); s != "" {
			return s
		}
	}

	// lib/pq: Error.Code is a string-like type with String() / SQLSTATE.
	type pqCoder interface{ Code() string }
	if e, ok := err.(pqCoder); ok {
		if s := strings.TrimSpace(e.Code()); s != "" {
			return s
		}
	}

	// Some drivers expose Code as fmt.Stringer (pq.ErrorCode).
	type stringCoder interface{ Code() fmt.Stringer }
	if e, ok := err.(stringCoder); ok {
		if c := e.Code(); c != nil {
			if s := strings.TrimSpace(c.String()); s != "" {
				return s
			}
		}
	}

	// go-sql-driver/mysql: Number() uint16
	type mysqlNumberer interface{ Number() uint16 }
	if e, ok := err.(mysqlNumberer); ok {
		return strconv.FormatUint(uint64(e.Number()), 10)
	}

	// SQLite drivers: Code() / ExtendedCode() as int or int64
	type intCoder interface{ Code() int }
	if e, ok := err.(intCoder); ok {
		return strconv.Itoa(e.Code())
	}
	type int64Coder interface{ Code() int64 }
	if e, ok := err.(int64Coder); ok {
		return strconv.FormatInt(e.Code(), 10)
	}
	type extCoder interface{ ExtendedCode() int }
	if e, ok := err.(extCoder); ok {
		return strconv.Itoa(e.ExtendedCode())
	}
	type ext64Coder interface{ ExtendedCode() int64 }
	if e, ok := err.(ext64Coder); ok {
		return strconv.FormatInt(e.ExtendedCode(), 10)
	}

	// pgx / jackc: field named Code (exported) via interface with SQLSTATE string.
	type codeFielder interface{ GetCode() string }
	if e, ok := err.(codeFielder); ok {
		if s := strings.TrimSpace(e.GetCode()); s != "" {
			return s
		}
	}

	return ""
}

func classifyCode(code string) (errorClass, bool) {
	if code == "" {
		return -1, false
	}
	switch strings.ToUpper(code) {
	case PGUniqueViolation:
		return classUnique, true
	case PGForeignKeyViolation:
		return classFK, true
	case PGNotNullViolation:
		return classNotNull, true
	case PGCheckViolation:
		return classCheck, true
	case PGSerializationFailure, PGDeadlockDetected:
		return classSerialization, true
	}

	n, err := strconv.Atoi(code)
	if err != nil {
		return -1, false
	}
	switch n {
	case MySQLDupEntry, SQLiteConstraintUnique, SQLiteConstraintPrimaryKey:
		return classUnique, true
	case MySQLRowIsReferenced, MySQLNoReferencedRow, SQLiteConstraintForeignKey:
		return classFK, true
	case MySQLBadNull, SQLiteConstraintNotNull:
		return classNotNull, true
	case MySQLCheckConstraint, SQLiteConstraintCheck:
		return classCheck, true
	case MySQLLockDeadlock, MySQLLockWaitTimeout, SQLiteBusy, SQLiteLocked:
		return classSerialization, true
	case SQLiteConstraint:
		// Generic constraint; message fallback may refine, but alone is ambiguous.
		return -1, false
	}
	return -1, false
}

func classifyMessage(msg string, want errorClass) bool {
	c, ok := classifyMessageOK(msg)
	return ok && c == want
}

func classifyMessageOK(msg string) (errorClass, bool) {
	m := strings.ToLower(msg)

	// Serialization / deadlock first (often overlap with "constraint" wording less).
	if strings.Contains(m, "serialization failure") ||
		strings.Contains(m, "could not serialize") ||
		strings.Contains(m, "deadlock") ||
		strings.Contains(m, "lock wait timeout") ||
		strings.Contains(m, "database is locked") ||
		strings.Contains(m, "sqlite_busy") ||
		strings.Contains(m, "sqlite_locked") {
		return classSerialization, true
	}

	if strings.Contains(m, "unique constraint") ||
		strings.Contains(m, "duplicate key") ||
		strings.Contains(m, "duplicate entry") ||
		strings.Contains(m, "unique_violation") ||
		strings.Contains(m, "sqlite_constraint_unique") ||
		strings.Contains(m, "sqlite_constraint_primarykey") ||
		strings.Contains(m, "primary key constraint") {
		return classUnique, true
	}

	if strings.Contains(m, "foreign key constraint") ||
		strings.Contains(m, "foreign_key_violation") ||
		strings.Contains(m, "sqlite_constraint_foreignkey") ||
		strings.Contains(m, "references constraint") {
		return classFK, true
	}

	if strings.Contains(m, "not null constraint") ||
		strings.Contains(m, "not_null_violation") ||
		strings.Contains(m, "sqlite_constraint_notnull") ||
		strings.Contains(m, "cannot be null") ||
		strings.Contains(m, "column cannot be null") {
		return classNotNull, true
	}

	if strings.Contains(m, "check constraint") ||
		strings.Contains(m, "check_violation") ||
		strings.Contains(m, "sqlite_constraint_check") ||
		strings.Contains(m, "check constraint violated") {
		return classCheck, true
	}

	return -1, false
}
