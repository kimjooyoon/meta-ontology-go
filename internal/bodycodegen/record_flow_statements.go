package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/constant"
	"strconv"
)

func (c *recordFlowBuilder) block(block *ast.BlockStmt, state *flowState) bool {
	inherited := state.count
	defer func() { state.count = inherited }()
	for _, statement := range block.List {
		c.guardRoot = state.guard
		if c.err != nil || !c.statement(statement, state) {
			return false
		}
	}
	return true
}

func (c *recordFlowBuilder) statement(node ast.Stmt, state *flowState) bool {
	switch s := node.(type) {
	case *ast.DeclStmt:
		decl := s.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
		value := c.expression(decl.Values[0], state, true)
		c.bind(state, decl.Names[0].Name, value, c.span(decl), "copy")
	case *ast.AssignStmt:
		c.assignment(s, state)
	case *ast.IfStmt:
		return c.branch(s, state)
	case *ast.ReturnStmt:
		value := c.wrap(c.expression(s.Results[0], state, true), c.span(s), "return", 0)
		c.retainReturn(value, c.span(s))
		return false
	default:
		c.err = fmt.Errorf("FLOW_STATEMENT_UNSUPPORTED")
	}
	return true
}

func (c *recordFlowBuilder) assignment(s *ast.AssignStmt, state *flowState) {
	name, ok := assignmentLocalName(s.Lhs[0])
	binding := state.find(name)
	if !ok || binding == nil {
		c.err = fmt.Errorf("FLOW_ASSIGNMENT_UNRESOLVED")
		return
	}
	value := c.expression(s.Rhs[0], state, true)
	if selector, ok := s.Lhs[0].(*ast.SelectorExpr); ok && binding.value.record >= 0 {
		for i, field := range c.body.records[binding.value.record].Fields {
			if field.Name == selector.Sel.Name || field.GoName == selector.Sel.Name {
				binding.value.fields[i] = c.add("write", [2]uint16{value.scalar}, c.span(s), field.ID, binding.id, "")
				return
			}
		}
		c.err = fmt.Errorf("FLOW_ASSIGNMENT_FIELD_UNRESOLVED")
		return
	}
	binding.value = c.wrap(value, c.span(s), "write", binding.id)
}

func (c *recordFlowBuilder) branch(s *ast.IfStmt, state *flowState) bool {
	if c.depth == 16 {
		c.err = fmt.Errorf("FLOW_BRANCH_BOUND")
		return false
	}
	condition := c.expression(s.Cond, state, true).scalar
	parentGuard := state.guard
	thenGuard := c.add("guard", [2]uint16{condition, parentGuard}, c.span(s.Cond), "", 0, "true")
	elseGuard := c.add("guard", [2]uint16{condition, parentGuard}, c.span(s.Cond), "", 0, "false")
	then, otherwise, thenLive, elseLive := c.branchArms(s, *state, thenGuard, elseGuard)
	c.guardRoot = parentGuard
	if value := c.info.Types[s.Cond].Value; value != nil && value.Kind() == constant.Bool {
		if condition != 0 {
			c.nodes[condition-1].KnownTruth = strconv.FormatBool(constant.BoolVal(value))
		}
		if constant.BoolVal(value) {
			*state = then
			return thenLive
		}
		*state = otherwise
		return elseLive
	}
	live := c.joinStates(state, then, otherwise, thenLive, elseLive, c.span(s), condition)
	if thenLive && elseLive && (then.guard != thenGuard || otherwise.guard != elseGuard) {
		state.guard = c.add("guard_join", [2]uint16{then.guard, otherwise.guard}, c.span(s), "", 0, "or")
	}
	return live
}

func (c *recordFlowBuilder) branchArms(s *ast.IfStmt, parent flowState, thenGuard, elseGuard uint16) (flowState, flowState, bool, bool) {
	then, otherwise := parent, parent
	then.guard, otherwise.guard = thenGuard, elseGuard
	c.depth++
	defer func() { c.depth-- }()
	suppressed := c.returnSuppressed
	defer func() { c.returnSuppressed = suppressed }()
	value := c.info.Types[s.Cond].Value
	known := value != nil && value.Kind() == constant.Bool
	c.returnSuppressed = suppressed || known && !constant.BoolVal(value)
	thenLive, elseLive := c.block(s.Body, &then), true
	c.returnSuppressed = suppressed || known && constant.BoolVal(value)
	if s.Else != nil {
		c.guardRoot = otherwise.guard
		if block, ok := s.Else.(*ast.BlockStmt); ok {
			elseLive = c.block(block, &otherwise)
		} else {
			elseLive = c.statement(s.Else.(ast.Stmt), &otherwise)
		}
	}
	return then, otherwise, thenLive, elseLive
}

func (c *recordFlowBuilder) joinStates(state *flowState, then, otherwise flowState, a, b bool, span RecordFlowSpan, guard uint16) bool {
	if !a {
		*state = otherwise
		return b
	}
	if !b {
		*state = then
		return a
	}
	for i := range state.count {
		x, y := then.bindings[i].value, otherwise.bindings[i].value
		if x.record < 0 {
			x.scalar = c.join(x.scalar, y.scalar, span, "", state.bindings[i].id, guard)
		} else {
			for j, field := range c.body.records[x.record].Fields {
				x.fields[j] = c.join(x.fields[j], y.fields[j], span, field.ID, state.bindings[i].id, guard)
			}
		}
		state.bindings[i].value = x
	}
	return true
}

func (c *recordFlowBuilder) join(a, b uint16, span RecordFlowSpan, field string, local uint16, guard uint16) uint16 {
	if a == b {
		return a
	}
	root := c.add("join", [2]uint16{a, b}, span, field, local, "")
	if root != 0 {
		c.nodes[root-1].Condition = guard
	}
	return root
}
