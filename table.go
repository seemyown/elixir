package elixir

import (
	"reflect"

	"github.com/seemyown/elixir/internal/ast"
)

// Table identifies a relation that can appear in FROM / JOIN clauses.
type Table interface {
	tableRef() ast.RelationNode
}

// TableRef is an explicit named table, optionally aliased.
//
// Embed TableRef in a table definition struct so the value can be passed to
// From / Join. Prefer Bind so column Table qualifiers inherit the parent name:
//
//	type UserTable struct {
//		TableRef
//		ID    Column[int64]
//		Email Column[string]
//	}
//
//	var Users = Bind(UserTable{
//		TableRef: TableRef{Name: "users"},
//		ID:       Column[int64]{Name: "id"},
//		Email:    Column[string]{Name: "email"},
//	})
type TableRef struct {
	// Schema, when non-empty, qualifies the table as schema.table in FROM / DML.
	// Column qualifiers still use Alias or Name, never Schema. Bind / As leave Schema unchanged.
	Schema string
	Name   string
	Alias  string
}

func (t TableRef) tableRef() ast.RelationNode {
	return ast.RelationNode{Schema: t.Schema, Name: t.Name, Alias: t.Alias}
}

// TableName returns the underlying table name.
func (t TableRef) TableName() string { return t.Name }

// As returns a copy of the table reference with the given alias.
func (t TableRef) As(alias string) TableRef {
	return TableRef{Schema: t.Schema, Name: t.Name, Alias: alias}
}

type columnMarker interface {
	isColumn()
}

// Bind fills empty Column.Table fields from the embedded TableRef.
//
// Qualifier is Alias when set, otherwise Name. Reflection runs once at init;
// it is not used when compiling SQL.
func Bind[T any](t T) T {
	v := reflect.ValueOf(&t).Elem()
	name, alias := tableRefNameAlias(v)
	qualifier := name
	if alias != "" {
		qualifier = alias
	}
	if qualifier != "" {
		setColumnTables(v, qualifier, false)
	}
	return t
}

// As returns a copy of table with TableRef.Alias set and all Column.Table
// fields rebound to that alias (for JOINs). Prefer Bind for the base table,
// then As for aliased copies:
//
//	U := As(Users, "u")
func As[T any](table T, alias string) T {
	v := reflect.ValueOf(&table).Elem()
	setTableRefAlias(v, alias)
	setColumnTables(v, alias, true)
	return table
}

func tableRefNameAlias(v reflect.Value) (name, alias string) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Type == reflect.TypeOf(TableRef{}) {
			tr := v.Field(i).Interface().(TableRef)
			return tr.Name, tr.Alias
		}
	}
	return "", ""
}

func setTableRefAlias(v reflect.Value, alias string) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Type != reflect.TypeOf(TableRef{}) {
			continue
		}
		f := v.Field(i)
		if !f.CanSet() {
			return
		}
		tr := f.Interface().(TableRef)
		tr.Alias = alias
		f.Set(reflect.ValueOf(tr))
		return
	}
}

func setColumnTables(v reflect.Value, qualifier string, overwrite bool) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanInterface() || !f.CanSet() {
			continue
		}
		if _, ok := f.Interface().(columnMarker); !ok {
			continue
		}
		tableField := f.FieldByName("Table")
		if !tableField.IsValid() || !tableField.CanSet() || tableField.Kind() != reflect.String {
			continue
		}
		if !overwrite && tableField.String() != "" {
			continue
		}
		tableField.SetString(qualifier)
	}
}

// columnDefaults walks table struct fields and returns DEFAULT assignments for
// columns whose HasDefault metadata is true.
func columnDefaults(src any) []ast.AssignNode {
	if src == nil {
		return nil
	}
	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	var out []ast.AssignNode
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanInterface() {
			continue
		}
		if _, ok := f.Interface().(columnMarker); !ok {
			continue
		}
		hasDef := f.FieldByName("HasDefault")
		if !hasDef.IsValid() || hasDef.Kind() != reflect.Bool || !hasDef.Bool() {
			continue
		}
		nameField := f.FieldByName("Name")
		tableField := f.FieldByName("Table")
		if !nameField.IsValid() || nameField.Kind() != reflect.String {
			continue
		}
		table := ""
		if tableField.IsValid() && tableField.Kind() == reflect.String {
			table = tableField.String()
		}
		out = append(out, ast.AssignNode{
			Column: ast.ColumnNode{Table: table, Name: nameField.String()},
			Value:  ast.DefaultNode{},
		})
	}
	return out
}

// subqueryTable wraps a SELECT as a FROM / JOIN source.
type subqueryTable struct {
	query ast.SelectNode
	alias string
}

func (s subqueryTable) tableRef() ast.RelationNode {
	q := s.query
	// Subqueries used as tables should not carry an outer WITH clause.
	q.With = nil
	return ast.RelationNode{Subquery: &q, Alias: s.alias}
}
