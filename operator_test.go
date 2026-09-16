package elixir

import "testing"

func TestOperatorSQL(t *testing.T) {
	if OpEq.SQL() != "=" || OpNe.SQL() != "<>" || OpGt.SQL() != ">" {
		t.Fatal("binary SQL tokens")
	}
	if OpGte.SQL() != ">=" || OpLt.SQL() != "<" || OpLte.SQL() != "<=" {
		t.Fatal("binary SQL tokens")
	}
	if OpNot.SQL() != "NOT" || OpIsNull.SQL() != "IS NULL" || OpIsNotNull.SQL() != "IS NOT NULL" {
		t.Fatal("unary SQL tokens")
	}
	if OpAnd.SQL() != "AND" || OpOr.SQL() != "OR" {
		t.Fatal("logical SQL tokens")
	}
	if BinaryOperator(99).SQL() != "?" || UnaryOperator(99).SQL() != "?" || LogicalOperator(99).SQL() != "?" {
		t.Fatal("unknown op fallback")
	}
}
