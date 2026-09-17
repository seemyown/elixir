package ast

// BinaryOperator is a comparison operator between two expressions.
type BinaryOperator uint8

const (
	OpEq BinaryOperator = iota
	OpNe
	OpGt
	OpGte
	OpLt
	OpLte
	OpLike
	OpNotLike
	OpILike
	OpNotILike
	OpAdd
	OpSub
	OpMul
	OpDiv
)

// SQL returns the SQL token for the operator.
func (op BinaryOperator) SQL() string {
	switch op {
	case OpEq:
		return "="
	case OpNe:
		return "<>"
	case OpGt:
		return ">"
	case OpGte:
		return ">="
	case OpLt:
		return "<"
	case OpLte:
		return "<="
	case OpLike:
		return "LIKE"
	case OpNotLike:
		return "NOT LIKE"
	case OpILike:
		return "ILIKE"
	case OpNotILike:
		return "NOT ILIKE"
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	default:
		return "?"
	}
}

// UnaryOperator is a unary SQL operator.
type UnaryOperator uint8

const (
	OpNot UnaryOperator = iota
	OpIsNull
	OpIsNotNull
)

// SQL returns the SQL token for the unary operator.
func (op UnaryOperator) SQL() string {
	switch op {
	case OpNot:
		return "NOT"
	case OpIsNull:
		return "IS NULL"
	case OpIsNotNull:
		return "IS NOT NULL"
	default:
		return "?"
	}
}

// LogicalOperator combines predicates.
type LogicalOperator uint8

const (
	OpAnd LogicalOperator = iota
	OpOr
)

// SQL returns the SQL token for the logical operator.
func (op LogicalOperator) SQL() string {
	switch op {
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	default:
		return "?"
	}
}
