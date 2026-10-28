package bodycodegen

import (
	"fmt"
	"go/ast"
)

type RecordFlowSequenceType struct {
	ElementTypeID string `json:"element_type_id"`
}

type RecordFlowLoop struct {
	ID           uint16         `json:"id"`
	Parent       uint16         `json:"parent,omitempty"`
	Collection   uint16         `json:"collection"`
	SequenceType uint16         `json:"sequence_type"`
	Index        uint16         `json:"index"`
	Element      uint16         `json:"element"`
	Entry        uint16         `json:"entry"`
	Done         uint16         `json:"done"`
	Continues    bool           `json:"continues"`
	IndexKind    string         `json:"index_kind"`
	Span         RecordFlowSpan `json:"span"`
}

// Yield is a separate recurrence edge, never a backward ordinary parent.
// A zero yield means every entered-body path returns before another iteration.
type RecordFlowLoopCarry struct {
	Loop   uint16 `json:"loop"`
	Header uint16 `json:"header"`
	Seed   uint16 `json:"seed"`
	Yield  uint16 `json:"yield"`
	Exit   uint16 `json:"exit"`
}

type recordFlowLoopData struct {
	active  uint16
	count   uint16
	loops   [recordFlowNodeLimit]RecordFlowLoop
	carries [recordFlowNodeLimit]RecordFlowLoopCarry
}

func (c *recordFlowBuilder) rangeLoop(loop *ast.RangeStmt, state *flowState) bool {
	name, err := collectionLoopShape(loop)
	if err != nil || c.depth == 16 {
		c.err = fmt.Errorf("FLOW_LOOP_SHAPE_OR_DEPTH")
		return false
	}
	sequence := c.expression(loop.X, state, true)
	if c.err != nil {
		return false
	}
	if sequence.sequence == 0 {
		c.err = fmt.Errorf("FLOW_SEQUENCE_UNRESOLVED")
		return false
	}
	if c.loopData == nil {
		c.loopData = new(recordFlowLoopData)
	}
	entry := c.add("loop_entry", [2]uint16{sequence.scalar}, c.span(loop.X), "", 0, "has_element")
	if entry == 0 {
		return false
	}
	priorLoop, parentGuard := c.loopData.active, state.guard
	c.loopData.count++
	id := c.loopData.count
	c.loopData.active = id
	c.depth++
	defer func() { c.loopData.active = priorLoop; c.depth-- }()
	next := *state
	next.guard = c.add("guard", [2]uint16{entry, parentGuard}, c.span(loop.X), "", 0, "true")
	c.guardRoot = next.guard
	seeds, headers := *state, *state
	written := loopWrittenBindings(loop.Body, state)
	for i := range state.count {
		if written[i] {
			next.bindings[i].value = c.openLoopCarry(state.bindings[i].value, c.span(loop), state.bindings[i].id, "")
			headers.bindings[i] = next.bindings[i]
		}
	}
	index := c.add("loop_index", [2]uint16{}, c.span(loop.Value), "", 0, "symbolic_any")
	element := c.sequenceElement(sequence, index, c.span(loop.Value))
	c.loopData.loops[id-1] = RecordFlowLoop{ID: id, Parent: priorLoop, Collection: sequence.scalar,
		SequenceType: sequence.sequence, Index: index, Element: element.scalar, Entry: entry,
		IndexKind: "symbolic_any", Span: c.span(loop)}
	c.bind(&next, name, element, c.span(loop.Value), "copy")
	live := c.block(loop.Body, &next)
	c.guardRoot = parentGuard
	for i := range state.count {
		if written[i] {
			state.bindings[i].value = c.closeLoopCarry(seeds.bindings[i].value, headers.bindings[i].value,
				next.bindings[i].value, c.span(loop), state.bindings[i].id, "", live)
		}
	}
	if !live {
		state.guard = c.add("guard", [2]uint16{entry, parentGuard}, c.span(loop.X), "", 0, "false")
	}
	item := &c.loopData.loops[id-1]
	item.Continues, item.Done = live, c.add("loop_done", [2]uint16{entry}, c.span(loop), "", 0, "zero_or_more")
	return c.err == nil
}

func loopWrittenBindings(body *ast.BlockStmt, state *flowState) [recordFlowLocalLimit]bool {
	var written [recordFlowLocalLimit]bool
	ast.Inspect(body, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) != 1 {
			return true
		}
		name, ok := assignmentLocalName(assignment.Lhs[0])
		if ok {
			for i := range state.count {
				written[i] = written[i] || state.bindings[i].name == name
			}
		}
		return true
	})
	return written
}

func (c *recordFlowBuilder) exportLoops(flow *RecordValueFlow) {
	if c.err != nil || len(c.sequenceTypes) == 0 && c.loopData == nil {
		return
	}
	flow.Schema = recordLoopFlowSchema
	flow.SequenceTypes = append([]RecordFlowSequenceType(nil), c.sequenceTypes...)
	if c.loopData != nil {
		flow.Loops = append([]RecordFlowLoop(nil), c.loopData.loops[:c.loopData.count]...)
		for _, carry := range c.loopData.carries[:c.count] {
			if carry.Header != 0 {
				flow.LoopCarries = append(flow.LoopCarries, carry)
			}
		}
	}
}
