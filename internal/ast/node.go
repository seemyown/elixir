package ast

// Node is the untyped AST building block used by the SQL compiler.
type Node interface {
	node()
}

type ColumnNode struct {
	Table string
	Name  string
}

func (ColumnNode) node() {}

type LiteralNode struct {
	Value any
}

func (LiteralNode) node() {}

// NullNode renders SQL NULL (not a bound parameter).
type NullNode struct{}

func (NullNode) node() {}

// DefaultNode renders SQL DEFAULT (INSERT WithDefaults).
type DefaultNode struct{}

func (DefaultNode) node() {}

type BinaryNode struct {
	Left  Node
	Op    BinaryOperator
	Right Node
}

func (BinaryNode) node() {}

type UnaryNode struct {
	Op   UnaryOperator
	Expr Node
}

func (UnaryNode) node() {}

type LogicalNode struct {
	Op    LogicalOperator
	Exprs []Node
}

func (LogicalNode) node() {}

type FunctionNode struct {
	Name     string
	Args     []Node
	Distinct bool
}

func (FunctionNode) node() {}

type AliasNode struct {
	Expr  Node
	Alias string
}

func (AliasNode) node() {}

type DistinctNode struct {
	Expr Node
}

func (DistinctNode) node() {}

type StarNode struct{}

func (StarNode) node() {}

// RelationNode is a table, CTE name, or subquery usable in FROM / JOIN.
type RelationNode struct {
	Name     string
	Alias    string
	Subquery *SelectNode
}

func (RelationNode) node() {}

type OrderNode struct {
	Expr Node
	Desc bool
}

func (OrderNode) node() {}

type JoinKind uint8

const (
	JoinInner JoinKind = iota
	JoinLeft
	JoinRight
	JoinFull
)

type JoinNode struct {
	Kind     JoinKind
	Relation RelationNode
	On       Node
}

func (JoinNode) node() {}

type CTENode struct {
	Name  string
	Query SelectNode
}

func (CTENode) node() {}

type SelectNode struct {
	With    []CTENode
	Columns []Node
	From    *RelationNode
	Joins   []JoinNode
	Where   []Node
	GroupBy []Node
	Having  []Node
	OrderBy []OrderNode
	Limit   *int
	Offset  *int
}

func (SelectNode) node() {}

type AssignNode struct {
	Column Node
	Value  Node
}

func (AssignNode) node() {}

type InsertNode struct {
	With      []CTENode
	Table     RelationNode
	Columns   []Node
	Rows      [][]Node
	Returning []Node
}

func (InsertNode) node() {}

type UpdateNode struct {
	With      []CTENode
	Table     RelationNode
	Sets      []AssignNode
	Where     []Node
	Returning []Node
}

func (UpdateNode) node() {}

type DeleteNode struct {
	With      []CTENode
	Table     RelationNode
	Where     []Node
	Returning []Node
}

func (DeleteNode) node() {}

type SubqueryExprNode struct {
	Query SelectNode
}

func (SubqueryExprNode) node() {}

type InNode struct {
	Expr  Node
	Not   bool
	List  []Node
	Query *SelectNode
}

func (InNode) node() {}

type ExistsNode struct {
	Not   bool
	Query SelectNode
}

func (ExistsNode) node() {}

type FrameUnit uint8

const (
	FrameRows FrameUnit = iota
	FrameRange
)

type FrameBoundKind uint8

const (
	FrameUnboundedPreceding FrameBoundKind = iota
	FrameCurrentRow
	FrameUnboundedFollowing
	FramePreceding
	FrameFollowing
)

type FrameBoundNode struct {
	Kind   FrameBoundKind
	Offset int
}

type FrameNode struct {
	Unit  FrameUnit
	Start FrameBoundNode
	End   FrameBoundNode
}

type WindowNode struct {
	Expr        Node
	PartitionBy []Node
	OrderBy     []OrderNode
	Frame       *FrameNode
}

func (WindowNode) node() {}

type CaseWhenNode struct {
	When Node
	Then Node
}

type CaseNode struct {
	Value Node // nil for searched CASE; set for simple CASE
	Whens []CaseWhenNode
	Else  Node // optional
}

func (CaseNode) node() {}
