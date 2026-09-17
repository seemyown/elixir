package compiler

import (
	"fmt"
	"strings"

	"github.com/seemyown/elixir/internal/ast"
)

// Dialect customizes SQL rendering for a specific database.
type Dialect interface {
	Placeholder(n int) string
	QuoteIdent(name string) string
}

type compiler struct {
	d    Dialect
	args []any
	err  error
}

// CompileSelect renders a SELECT statement.
func CompileSelect(d Dialect, q ast.SelectNode) (string, []any, error) {
	if d == nil {
		return "", nil, fmt.Errorf("elixir: dialect is nil")
	}
	c := &compiler{d: d}
	sql := c.selectSQL(q, true)
	if c.err != nil {
		return "", nil, c.err
	}
	return sql, c.args, nil
}

// CompileInsert renders an INSERT statement.
func CompileInsert(d Dialect, q ast.InsertNode) (string, []any, error) {
	if d == nil {
		return "", nil, fmt.Errorf("elixir: dialect is nil")
	}
	c := &compiler{d: d}
	sql := c.insertSQL(q)
	if c.err != nil {
		return "", nil, c.err
	}
	return sql, c.args, nil
}

// CompileUpdate renders an UPDATE statement.
func CompileUpdate(d Dialect, q ast.UpdateNode) (string, []any, error) {
	if d == nil {
		return "", nil, fmt.Errorf("elixir: dialect is nil")
	}
	c := &compiler{d: d}
	sql := c.updateSQL(q)
	if c.err != nil {
		return "", nil, c.err
	}
	return sql, c.args, nil
}

// CompileDelete renders a DELETE statement.
func CompileDelete(d Dialect, q ast.DeleteNode) (string, []any, error) {
	if d == nil {
		return "", nil, fmt.Errorf("elixir: dialect is nil")
	}
	c := &compiler{d: d}
	sql := c.deleteSQL(q)
	if c.err != nil {
		return "", nil, c.err
	}
	return sql, c.args, nil
}

// CompileExpr renders a standalone expression or predicate.
func CompileExpr(d Dialect, n ast.Node) (string, []any, error) {
	if d == nil {
		return "", nil, fmt.Errorf("elixir: dialect is nil")
	}
	c := &compiler{d: d}
	sql := c.expr(n)
	if c.err != nil {
		return "", nil, c.err
	}
	return sql, c.args, nil
}

func (c *compiler) selectSQL(q ast.SelectNode, requireFrom bool) string {
	if len(q.Columns) == 0 {
		c.err = fmt.Errorf("elixir: SELECT requires at least one column")
		return ""
	}
	if requireFrom && q.From == nil {
		c.err = fmt.Errorf("elixir: SELECT requires FROM")
		return ""
	}

	var b strings.Builder
	c.writeWith(&b, q.With)

	b.WriteString("SELECT ")
	for i, col := range q.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.expr(col))
	}

	if q.From != nil {
		b.WriteString(" FROM ")
		b.WriteString(c.relation(*q.From))
	}

	for _, j := range q.Joins {
		b.WriteByte(' ')
		b.WriteString(joinSQL(j.Kind))
		b.WriteByte(' ')
		b.WriteString(c.relation(j.Relation))
		b.WriteString(" ON ")
		b.WriteString(c.expr(j.On))
	}

	if len(q.Where) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(c.andGroup(q.Where))
	}

	if len(q.GroupBy) > 0 {
		b.WriteString(" GROUP BY ")
		for i, g := range q.GroupBy {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.expr(g))
		}
	}

	if len(q.Having) > 0 {
		b.WriteString(" HAVING ")
		b.WriteString(c.andGroup(q.Having))
	}

	if len(q.OrderBy) > 0 {
		b.WriteString(" ORDER BY ")
		c.writeOrderBy(&b, q.OrderBy)
	}

	if q.Limit != nil {
		b.WriteString(" LIMIT ")
		b.WriteString(c.bind(*q.Limit))
	}
	if q.Offset != nil {
		b.WriteString(" OFFSET ")
		b.WriteString(c.bind(*q.Offset))
	}

	return b.String()
}

func (c *compiler) insertSQL(q ast.InsertNode) string {
	if q.Table.Name == "" || q.Table.Subquery != nil {
		c.err = fmt.Errorf("elixir: INSERT requires a named table")
		return ""
	}
	hasRows := len(q.Rows) > 0
	hasSelect := q.Select != nil
	if hasRows && hasSelect {
		c.err = fmt.Errorf("elixir: INSERT cannot mix VALUES and SELECT")
		return ""
	}
	if !hasRows && !hasSelect {
		c.err = fmt.Errorf("elixir: INSERT requires VALUES or SELECT")
		return ""
	}

	var b strings.Builder
	c.writeWith(&b, q.With)

	b.WriteString("INSERT INTO ")
	b.WriteString(c.relation(q.Table))

	if len(q.Columns) > 0 {
		b.WriteString(" (")
		for i, col := range q.Columns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.columnNameOnly(col))
		}
		b.WriteByte(')')
	}

	if hasSelect {
		if len(q.Select.With) > 0 {
			c.err = fmt.Errorf("elixir: INSERT SELECT cannot carry WITH; use With(...).Insert(...)")
			return ""
		}
		b.WriteByte(' ')
		b.WriteString(c.selectSQL(*q.Select, true))
	} else {
		b.WriteString(" VALUES ")
		for i, row := range q.Rows {
			if i > 0 {
				b.WriteString(", ")
			}
			if len(q.Columns) > 0 && len(row) != len(q.Columns) {
				c.err = fmt.Errorf("elixir: INSERT VALUES row %d has %d values, expected %d columns", i, len(row), len(q.Columns))
				return ""
			}
			if len(row) == 0 {
				c.err = fmt.Errorf("elixir: INSERT VALUES row %d is empty", i)
				return ""
			}
			b.WriteByte('(')
			for j, v := range row {
				if j > 0 {
					b.WriteString(", ")
				}
				b.WriteString(c.expr(v))
			}
			b.WriteByte(')')
		}
	}

	c.writeConflict(&b, q.Conflict)
	c.writeReturning(&b, q.Returning)
	return b.String()
}

func (c *compiler) writeConflict(b *strings.Builder, conflict *ast.ConflictNode) {
	if conflict == nil || c.err != nil {
		return
	}

	hasCols := len(conflict.Columns) > 0
	hasConstraint := conflict.Constraint != ""
	if hasCols && hasConstraint {
		c.err = fmt.Errorf("elixir: ON CONFLICT cannot mix columns and constraint")
		return
	}

	b.WriteString(" ON CONFLICT")
	switch {
	case hasConstraint:
		b.WriteString(" ON CONSTRAINT ")
		b.WriteString(c.d.QuoteIdent(conflict.Constraint))
	case hasCols:
		b.WriteString(" (")
		for i, col := range conflict.Columns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.columnNameOnly(col))
		}
		b.WriteByte(')')
	}

	if conflict.DoNothing {
		b.WriteString(" DO NOTHING")
		return
	}

	if !hasCols && !hasConstraint {
		c.err = fmt.Errorf("elixir: ON CONFLICT DO UPDATE requires a conflict target")
		return
	}
	if len(conflict.Updates) == 0 {
		c.err = fmt.Errorf("elixir: ON CONFLICT DO UPDATE requires at least one SET")
		return
	}

	b.WriteString(" DO UPDATE SET ")
	for i, s := range conflict.Updates {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.columnNameOnly(s.Column))
		b.WriteString(" = ")
		b.WriteString(c.expr(s.Value))
	}
}

func (c *compiler) updateSQL(q ast.UpdateNode) string {
	if q.Table.Name == "" || q.Table.Subquery != nil {
		c.err = fmt.Errorf("elixir: UPDATE requires a named table")
		return ""
	}
	if len(q.Sets) == 0 {
		c.err = fmt.Errorf("elixir: UPDATE requires at least one SET")
		return ""
	}

	var b strings.Builder
	c.writeWith(&b, q.With)

	b.WriteString("UPDATE ")
	b.WriteString(c.relation(q.Table))
	b.WriteString(" SET ")
	for i, s := range q.Sets {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.columnNameOnly(s.Column))
		b.WriteString(" = ")
		b.WriteString(c.expr(s.Value))
	}

	if len(q.Where) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(c.andGroup(q.Where))
	}

	c.writeReturning(&b, q.Returning)
	return b.String()
}

func (c *compiler) deleteSQL(q ast.DeleteNode) string {
	if q.Table.Name == "" || q.Table.Subquery != nil {
		c.err = fmt.Errorf("elixir: DELETE requires a named table")
		return ""
	}

	var b strings.Builder
	c.writeWith(&b, q.With)

	b.WriteString("DELETE FROM ")
	b.WriteString(c.relation(q.Table))

	if len(q.Where) > 0 {
		b.WriteString(" WHERE ")
		b.WriteString(c.andGroup(q.Where))
	}

	c.writeReturning(&b, q.Returning)
	return b.String()
}

func (c *compiler) writeReturning(b *strings.Builder, cols []ast.Node) {
	if len(cols) == 0 {
		return
	}
	b.WriteString(" RETURNING ")
	for i, col := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.expr(col))
	}
}

func (c *compiler) writeWith(b *strings.Builder, ctes []ast.CTENode) {
	if len(ctes) == 0 {
		return
	}
	b.WriteString("WITH ")
	for i, cte := range ctes {
		if i > 0 {
			b.WriteString(", ")
		}
		if cte.Name == "" {
			c.err = fmt.Errorf("elixir: CTE requires a name")
			return
		}
		b.WriteString(c.d.QuoteIdent(cte.Name))
		b.WriteString(" AS (")
		inner := cte.Query
		inner.With = nil
		b.WriteString(c.selectSQL(inner, true))
		b.WriteByte(')')
	}
	b.WriteByte(' ')
}

func (c *compiler) writeOrderBy(b *strings.Builder, orders []ast.OrderNode) {
	for i, o := range orders {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.expr(o.Expr))
		if o.Desc {
			b.WriteString(" DESC")
		} else {
			b.WriteString(" ASC")
		}
	}
}

func joinSQL(k ast.JoinKind) string {
	switch k {
	case ast.JoinLeft:
		return "LEFT JOIN"
	case ast.JoinRight:
		return "RIGHT JOIN"
	case ast.JoinFull:
		return "FULL JOIN"
	default:
		return "JOIN"
	}
}

func (c *compiler) andGroup(nodes []ast.Node) string {
	if len(nodes) == 1 {
		return c.expr(nodes[0])
	}
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = c.expr(n)
	}
	return "(" + strings.Join(parts, " AND ") + ")"
}

func (c *compiler) relation(r ast.RelationNode) string {
	if r.Subquery != nil {
		if r.Alias == "" {
			c.err = fmt.Errorf("elixir: subquery in FROM/JOIN requires an alias")
			return ""
		}
		inner := *r.Subquery
		inner.With = nil
		return "(" + c.selectSQL(inner, true) + ") AS " + c.d.QuoteIdent(r.Alias)
	}
	if r.Name == "" {
		c.err = fmt.Errorf("elixir: relation requires a table name")
		return ""
	}
	s := c.d.QuoteIdent(r.Name)
	if r.Alias != "" {
		s += " AS " + c.d.QuoteIdent(r.Alias)
	}
	return s
}

func (c *compiler) columnNameOnly(n ast.Node) string {
	switch v := n.(type) {
	case ast.ColumnNode:
		return c.d.QuoteIdent(v.Name)
	case ast.AliasNode:
		return c.columnNameOnly(v.Expr)
	default:
		c.err = fmt.Errorf("elixir: expected column, got %T", n)
		return ""
	}
}

func (c *compiler) bind(v any) string {
	c.args = append(c.args, v)
	return c.d.Placeholder(len(c.args))
}

func (c *compiler) expr(n ast.Node) string {
	if c.err != nil {
		return ""
	}
	switch v := n.(type) {
	case ast.ColumnNode:
		if v.Table == "" {
			return c.d.QuoteIdent(v.Name)
		}
		return c.d.QuoteIdent(v.Table) + "." + c.d.QuoteIdent(v.Name)
	case ast.LiteralNode:
		return c.bind(v.Value)
	case ast.NullNode:
		return "NULL"
	case ast.DefaultNode:
		return "DEFAULT"
	case ast.BinaryNode:
		return "(" + c.expr(v.Left) + " " + v.Op.SQL() + " " + c.expr(v.Right) + ")"
	case ast.UnaryNode:
		switch v.Op {
		case ast.OpNot:
			return "(NOT " + c.expr(v.Expr) + ")"
		case ast.OpIsNull:
			return "(" + c.expr(v.Expr) + " IS NULL)"
		case ast.OpIsNotNull:
			return "(" + c.expr(v.Expr) + " IS NOT NULL)"
		default:
			c.err = fmt.Errorf("elixir: unknown unary operator %d", v.Op)
			return ""
		}
	case ast.LogicalNode:
		parts := make([]string, len(v.Exprs))
		for i, e := range v.Exprs {
			parts[i] = c.expr(e)
		}
		return "(" + strings.Join(parts, " "+v.Op.SQL()+" ") + ")"
	case ast.FunctionNode:
		var b strings.Builder
		b.WriteString(v.Name)
		b.WriteByte('(')
		if v.Distinct {
			b.WriteString("DISTINCT ")
		}
		for i, a := range v.Args {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(c.expr(a))
		}
		b.WriteByte(')')
		return b.String()
	case ast.AliasNode:
		return c.expr(v.Expr) + " AS " + c.d.QuoteIdent(v.Alias)
	case ast.DistinctNode:
		return "DISTINCT " + c.expr(v.Expr)
	case ast.StarNode:
		return "*"
	case ast.RelationNode:
		return c.relation(v)
	case ast.SubqueryExprNode:
		inner := v.Query
		inner.With = nil
		return "(" + c.selectSQL(inner, true) + ")"
	case ast.InNode:
		var b strings.Builder
		b.WriteByte('(')
		b.WriteString(c.expr(v.Expr))
		if v.Not {
			b.WriteString(" NOT IN ")
		} else {
			b.WriteString(" IN ")
		}
		if v.Query != nil {
			inner := *v.Query
			inner.With = nil
			b.WriteByte('(')
			b.WriteString(c.selectSQL(inner, true))
			b.WriteByte(')')
		} else {
			if len(v.List) == 0 {
				c.err = fmt.Errorf("elixir: IN requires at least one value")
				return ""
			}
			b.WriteByte('(')
			for i, item := range v.List {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(c.expr(item))
			}
			b.WriteByte(')')
		}
		b.WriteByte(')')
		return b.String()
	case ast.ExistsNode:
		inner := v.Query
		inner.With = nil
		if v.Not {
			return "(NOT EXISTS (" + c.selectSQL(inner, true) + "))"
		}
		return "(EXISTS (" + c.selectSQL(inner, true) + "))"
	case ast.WindowNode:
		var b strings.Builder
		b.WriteString(c.expr(v.Expr))
		b.WriteString(" OVER (")
		needSpace := false
		if len(v.PartitionBy) > 0 {
			b.WriteString("PARTITION BY ")
			for i, p := range v.PartitionBy {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(c.expr(p))
			}
			needSpace = true
		}
		if len(v.OrderBy) > 0 {
			if needSpace {
				b.WriteByte(' ')
			}
			b.WriteString("ORDER BY ")
			c.writeOrderBy(&b, v.OrderBy)
			needSpace = true
		}
		if v.Frame != nil {
			if needSpace {
				b.WriteByte(' ')
			}
			b.WriteString(c.frameSQL(*v.Frame))
		}
		b.WriteByte(')')
		return b.String()
	case ast.CaseNode:
		var b strings.Builder
		b.WriteString("CASE")
		if v.Value != nil {
			b.WriteByte(' ')
			b.WriteString(c.expr(v.Value))
		}
		for _, w := range v.Whens {
			b.WriteString(" WHEN ")
			b.WriteString(c.expr(w.When))
			b.WriteString(" THEN ")
			b.WriteString(c.expr(w.Then))
		}
		if v.Else != nil {
			b.WriteString(" ELSE ")
			b.WriteString(c.expr(v.Else))
		}
		b.WriteString(" END")
		return b.String()
	default:
		c.err = fmt.Errorf("elixir: unsupported AST node %T", n)
		return ""
	}
}

func (c *compiler) frameSQL(f ast.FrameNode) string {
	unit := "ROWS"
	if f.Unit == ast.FrameRange {
		unit = "RANGE"
	}
	return unit + " BETWEEN " + c.frameBoundSQL(f.Start) + " AND " + c.frameBoundSQL(f.End)
}

func (c *compiler) frameBoundSQL(b ast.FrameBoundNode) string {
	switch b.Kind {
	case ast.FrameUnboundedPreceding:
		return "UNBOUNDED PRECEDING"
	case ast.FrameCurrentRow:
		return "CURRENT ROW"
	case ast.FrameUnboundedFollowing:
		return "UNBOUNDED FOLLOWING"
	case ast.FramePreceding:
		return fmt.Sprintf("%d PRECEDING", b.Offset)
	case ast.FrameFollowing:
		return fmt.Sprintf("%d FOLLOWING", b.Offset)
	default:
		c.err = fmt.Errorf("elixir: unknown frame bound %d", b.Kind)
		return ""
	}
}
