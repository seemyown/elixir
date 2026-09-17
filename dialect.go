package elixir

import (
	"fmt"
	"strings"
)

// Dialect customizes SQL rendering for a specific database.
type Dialect interface {
	// Placeholder returns the nth parameter placeholder (1-based).
	Placeholder(n int) string
	// QuoteIdent quotes an identifier (table/column/alias).
	QuoteIdent(name string) string
}

type postgresDialect struct{}

// Postgres returns the PostgreSQL dialect ($1, $2, …).
func Postgres() Dialect { return postgresDialect{} }

func (postgresDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (postgresDialect) QuoteIdent(name string) string {
	return quoteDouble(name)
}

func (postgresDialect) ilikeOK() bool { return true }

type questionDialect struct{}

func (questionDialect) Placeholder(n int) string {
	_ = n
	return "?"
}

func (questionDialect) QuoteIdent(name string) string {
	return quoteDouble(name)
}

func (questionDialect) unionBare() bool { return true }

// MySQL returns a MySQL-oriented dialect (? placeholders, backtick quoting).
func MySQL() Dialect { return mysqlDialect{} }

type mysqlDialect struct{}

func (mysqlDialect) Placeholder(n int) string {
	_ = n
	return "?"
}

func (mysqlDialect) QuoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// SQLite returns the SQLite dialect (? placeholders).
func SQLite() Dialect { return questionDialect{} }

func quoteDouble(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
