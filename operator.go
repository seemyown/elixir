package elixir

import "github.com/seemyown/elixir/internal/ast"

// BinaryOperator is a comparison operator between two expressions.
type BinaryOperator = ast.BinaryOperator

const (
	OpEq       = ast.OpEq
	OpNe       = ast.OpNe
	OpGt       = ast.OpGt
	OpGte      = ast.OpGte
	OpLt       = ast.OpLt
	OpLte      = ast.OpLte
	OpLike     = ast.OpLike
	OpNotLike  = ast.OpNotLike
	OpILike    = ast.OpILike
	OpNotILike = ast.OpNotILike
	OpAdd      = ast.OpAdd
	OpSub      = ast.OpSub
	OpMul      = ast.OpMul
	OpDiv      = ast.OpDiv
)

// UnaryOperator is a unary SQL operator.
type UnaryOperator = ast.UnaryOperator

const (
	OpNot       = ast.OpNot
	OpIsNull    = ast.OpIsNull
	OpIsNotNull = ast.OpIsNotNull
)

// LogicalOperator combines predicates.
type LogicalOperator = ast.LogicalOperator

const (
	OpAnd = ast.OpAnd
	OpOr  = ast.OpOr
)
