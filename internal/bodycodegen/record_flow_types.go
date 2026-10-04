package bodycodegen

import (
	"go/ast"
	"go/token"
	"go/types"
)

const recordFlowSchema = "gooo/record-value-flow/v1"
const recordFlowNodeLimit = 512
const recordFlowLocalLimit = 64

// RecordValueFlow records symbolic definitions in source order. Input values
// and finite expectations are absent; choices retain both permitted expressions.
type RecordValueFlow struct {
	Schema     string             `json:"schema"`
	Status     string             `json:"status"`
	Reason     string             `json:"reason,omitempty"`
	ActivityID string             `json:"activity_id"`
	BodySHA256 string             `json:"body_sha256"`
	BodyView   string             `json:"body_view"`
	Nodes      []RecordFlowNode   `json:"nodes,omitempty"`
	Choices    []RecordFlowChoice `json:"choices,omitempty"`
	Scope      string             `json:"scope"`
}

type RecordFlowSpan struct {
	View  string `json:"view"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type RecordFlowNode struct {
	ID         uint16         `json:"id"`
	Kind       string         `json:"kind"`
	FieldID    string         `json:"field_id,omitempty"`
	Local      uint16         `json:"local,omitempty"`
	Operator   string         `json:"operator,omitempty"`
	Parents    [2]uint16      `json:"parents"`
	Guard      uint16         `json:"guard,omitempty"`
	Condition  uint16         `json:"condition,omitempty"`
	KnownTruth string         `json:"known_truth,omitempty"`
	Span       RecordFlowSpan `json:"span"`
}

type RecordFlowChoice struct {
	ID      string         `json:"id"`
	FieldID string         `json:"field_id"`
	First   uint16         `json:"first"`
	Second  uint16         `json:"second"`
	Result  uint16         `json:"result"`
	Guard   uint16         `json:"guard,omitempty"`
	Span    RecordFlowSpan `json:"span"`
}

type flowValue struct {
	record int
	scalar uint16
	fields [16]uint16
}

type flowBinding struct {
	name  string
	id    uint16
	value flowValue
}

type flowState struct {
	bindings [recordFlowLocalLimit]flowBinding
	count    int
	guard    uint16
}

type recordFlowBuilder struct {
	body       preparedBody
	sites      []recordValueSite
	fset       *token.FileSet
	info       *types.Info
	base       int
	nodes      [recordFlowNodeLimit]RecordFlowNode
	count      int
	locals     uint16
	choices    [6]RecordFlowChoice
	choiceSeen [6]bool
	depth      int
	guardRoot  uint16
	err        error
	view       string
	bodyView   string
	altSet     *token.FileSet
}

func flowScalar(root uint16) flowValue { return flowValue{record: -1, scalar: root} }

func (s *flowState) find(name string) *flowBinding {
	for i := s.count - 1; i >= 0; i-- {
		if s.bindings[i].name == name {
			return &s.bindings[i]
		}
	}
	return nil
}

func (c *recordFlowBuilder) span(node ast.Node) RecordFlowSpan {
	fset, base := c.fset, c.base
	if c.altSet != nil {
		fset, base = c.altSet, 0
	}
	return RecordFlowSpan{View: c.view, Start: fset.Position(node.Pos()).Offset - base,
		End: fset.Position(node.End()).Offset - base}
}
