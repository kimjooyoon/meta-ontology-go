package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/types"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Literal item values and shapes have separate roots. Reading one item field
// must not acquire the other fields through a container-wide data edge.
type flowSequenceOrigin struct {
	shape, left, right, item, condition uint16
	length                              int16
	joined                              bool
}

func (c *recordFlowBuilder) sequenceOrigin(root uint16) (flowSequenceOrigin, bool) {
	if root == 0 || c.sequenceOrigins == nil {
		return flowSequenceOrigin{}, false
	}
	origin := c.sequenceOrigins[root-1]
	return origin, origin.shape != 0
}

func (c *recordFlowBuilder) retainSequenceOrigin(root uint16, origin flowSequenceOrigin) {
	if root == 0 || c.err != nil {
		return
	}
	if c.sequenceOrigins == nil {
		c.sequenceOrigins = new([recordFlowNodeLimit]flowSequenceOrigin)
	}
	c.sequenceOrigins[root-1] = origin
}

func (c *recordFlowBuilder) sequenceShape(root uint16) uint16 {
	if origin, ok := c.sequenceOrigin(root); ok {
		return origin.shape
	}
	return root
}

func (c *recordFlowBuilder) sequenceElementRoot(root uint16) uint16 {
	if root != 0 && c.nodes[root-1].Kind == "expression" && c.nodes[root-1].Operator == "source_item" {
		return c.nodes[root-1].Parents[1]
	}
	return root
}

func (c *recordFlowBuilder) rejectConstructedSequenceCarry(roots ...uint16) bool {
	if c.err != nil {
		return true
	}
	for _, root := range roots {
		if _, ok := c.sequenceOrigin(root); ok {
			c.err = fmt.Errorf("FLOW_CONSTRUCTED_SEQUENCE_RECURRENCE_UNRESOLVED")
			return true
		}
	}
	return false
}

func (c *recordFlowBuilder) forwardSequenceOrigin(root uint16, kind string, parents [2]uint16,
	span RecordFlowSpan, field string, local uint16) {
	switch kind {
	case "read", "copy", "write", "return", "call", "call_parameter", "element_field":
		if origin, ok := c.sequenceOrigin(parents[0]); ok && c.err == nil {
			origin.shape = c.add("expression", [2]uint16{origin.shape, parents[1]}, span, field, local, "sequence_shape")
			c.retainSequenceOrigin(root, origin)
		}
	}
}

func (c *recordFlowBuilder) joinSequenceOrigins(root, left, right uint16, span RecordFlowSpan,
	field string, local, condition uint16) {
	a, aok := c.sequenceOrigin(left)
	b, bok := c.sequenceOrigin(right)
	if !aok && !bok || c.err != nil {
		return
	}
	shape := c.add("join", [2]uint16{c.sequenceShape(left), c.sequenceShape(right)}, span, field, local, "sequence_shape")
	if shape == 0 {
		return
	}
	c.nodes[shape-1].Condition = condition
	length := int16(-1)
	if aok && bok && a.length == b.length {
		length = a.length
	}
	c.retainSequenceOrigin(root, flowSequenceOrigin{shape: shape, left: left, right: right,
		condition: condition, length: length, joined: true})
}

func (c *recordFlowBuilder) sequenceElementType(literal *ast.CompositeLit) string {
	sequence, ok := c.info.Types[literal].Type.(*types.Slice)
	if !ok {
		return ""
	}
	if named, ok := sequence.Elem().(*types.Named); ok {
		for _, record := range c.body.records {
			if named.Obj().Name() == record.Name || named.Obj().Name() == record.GoName {
				return record.ID
			}
		}
	}
	if scalar, ok := sequence.Elem().Underlying().(*types.Basic); ok {
		switch scalar.Kind() {
		case types.Int64:
			return string(semantic.BuiltinIntegerTypeID)
		case types.String:
			return string(semantic.BuiltinStringTypeID)
		case types.Bool:
			return string(semantic.BuiltinBooleanTypeID)
		}
	}
	return ""
}

func (c *recordFlowBuilder) sequenceLiteral(literal *ast.CompositeLit, state *flowState) flowValue {
	element := c.sequenceElementType(literal)
	if element == "" || len(literal.Elts) > RepeatedFieldLimit {
		c.err = fmt.Errorf("FLOW_SEQUENCE_LITERAL_TYPE_OR_BOUND")
		return flowScalar(0)
	}
	root := c.add("expression", [2]uint16{}, c.span(literal), "", 0, "sequence_literal")
	c.retainSequenceOrigin(root, flowSequenceOrigin{shape: root})
	for position, expression := range literal.Elts {
		item := c.expression(expression, state, true)
		if item.scalar == 0 || c.err != nil {
			return flowScalar(0)
		}
		previous, _ := c.sequenceOrigin(root)
		data := c.nominalRoot(item.scalar, c.span(expression))
		next := c.add("expression", [2]uint16{root, data}, c.span(expression), "", 0,
			"sequence_item:"+strconv.Itoa(position))
		shape := c.add("expression", [2]uint16{previous.shape}, c.span(expression), "", 0,
			"sequence_shape:"+strconv.Itoa(position))
		c.retainSequenceOrigin(next, flowSequenceOrigin{shape: shape, left: root, item: item.scalar,
			length: int16(position + 1)})
		root = next
	}
	return c.sequenceValue(root, element)
}

// A single bounded memo table closes shared join paths without exponential
// expansion. It is local to this source read and never changes flowValue layout.
type sequenceProjection struct {
	builder   *recordFlowBuilder
	index     uint16
	known     int64
	span      RecordFlowSpan
	remaining int
	seen      [recordFlowNodeLimit]bool
	values    [recordFlowNodeLimit]flowValue
}

func (p *sequenceProjection) element(root uint16, element string) flowValue {
	c := p.builder
	if root == 0 || c.err != nil {
		return flowScalar(0)
	}
	if p.seen[root-1] {
		return p.values[root-1]
	}
	if p.remaining == 0 {
		c.err = fmt.Errorf("FLOW_SEQUENCE_PROJECTION_BOUND")
		return flowScalar(0)
	}
	p.remaining--
	origin, ok := c.sequenceOrigin(root)
	value := flowScalar(0)
	if !ok {
		value = c.sequenceElement(c.sequenceValue(root, element), p.index, p.span)
	} else if origin.joined {
		left := p.element(origin.left, element)
		right := p.element(origin.right, element)
		value = p.merge(left, right, origin.condition)
	} else {
		value = p.literal(origin, element)
	}
	p.seen[root-1], p.values[root-1] = true, value
	return value
}

func (p *sequenceProjection) literal(origin flowSequenceOrigin, element string) flowValue {
	c := p.builder
	if origin.length == 0 || p.known >= int64(origin.length) {
		c.err = fmt.Errorf("FLOW_SEQUENCE_LITERAL_INDEX_BOUND")
		return flowScalar(0)
	}
	if p.known >= 0 && p.known != int64(origin.length)-1 {
		return p.element(origin.left, element)
	}
	value := flowScalar(origin.item)
	if record, ok := c.nominalValue(origin.item); ok {
		value = record
	}
	if p.known < 0 && origin.length > 1 {
		return p.merge(p.element(origin.left, element), value, 0)
	}
	return value
}

func (p *sequenceProjection) merge(left, right flowValue, condition uint16) flowValue {
	c := p.builder
	root := c.joinField(left.scalar, right.scalar, p.span, "", 0, condition)
	if value, ok := c.nominalValue(root); ok {
		return value
	}
	return flowScalar(root)
}

func (c *recordFlowBuilder) projectSequenceFields(value flowValue, root uint16, span RecordFlowSpan) flowValue {
	if c.err != nil {
		return flowScalar(0)
	}
	if value.record < 0 {
		return flowScalar(c.add("expression", [2]uint16{value.scalar, root}, span, "", 0, "source_item"))
	}
	value.scalar = root
	for i, field := range c.body.records[value.record].Fields {
		previous := value.fields[i]
		value.fields[i] = c.add("element_field", [2]uint16{previous, root}, span, field.ID, 0, "source_item_field")
		if nested, ok := c.nominalValue(previous); ok {
			c.projectSequenceFields(nested, value.fields[i], span)
		}
	}
	c.retainNominalValue(value)
	return value
}
