package elixir

import "github.com/seemyown/elixir/internal/ast"

// starExpr is the SQL * (or table.*) projection expression.
type starExpr struct {
	table string
}

func (s starExpr) exprNode() ast.Node {
	return ast.StarNode{Table: s.table}
}

// All is the SQL * expression for SELECT lists:
//
//	Select(All()).From(Users)  // SELECT * FROM "users"
func All() Expression {
	return starExpr{}
}

// AllOf is a table-qualified star (table.*):
//
//	Select(AllOf(Users)).From(Users)  // SELECT "users".* FROM "users"
//
// Uses the table alias when set, otherwise the table name.
func AllOf(table Table) Expression {
	ref := table.tableRef()
	name := ref.Alias
	if name == "" {
		name = ref.Name
	}
	return starExpr{table: name}
}
