package bodycodegen

import (
	"fmt"
	"go/ast"
)

func (c *recordFlowBuilder) selectorBase(expression ast.Expr, state *flowState) (flowValue, uint16) {
	switch e := expression.(type) {
	case *ast.Ident:
		if binding := state.find(e.Name); binding != nil {
			return binding.value, binding.id
		}
	case *ast.ParenExpr:
		return c.selectorBase(e.X, state)
	case *ast.SelectorExpr:
		value, local, _ := c.selectorValue(e, state)
		return value, local
	}
	return c.expression(expression, state, true), 0
}

func (c *recordFlowBuilder) selectorValue(e *ast.SelectorExpr, state *flowState) (flowValue, uint16, string) {
	value, local := c.selectorBase(e.X, state)
	if value.record >= 0 {
		for i, field := range c.body.records[value.record].Fields {
			if field.Name != e.Sel.Name && field.GoName != e.Sel.Name {
				continue
			}
			if field.Cardinality != "" {
				root := c.add("read", [2]uint16{value.fields[i]}, c.span(e), field.ID, local, "sequence")
				c.linkRecordVersion(root, value.scalar)
				return c.sequenceValue(root, field.TypeID), local, field.ID
			}
			if c.compactRecords {
				root := c.add("read", [2]uint16{value.fields[i]}, c.span(e), field.ID, local, "")
				c.linkRecordVersion(root, value.scalar)
				if nested, ok := c.nominalValue(value.fields[i]); ok {
					nested.scalar = root
					c.retainNominalValue(nested)
					return nested, local, field.ID
				}
				return flowScalar(root), local, field.ID
			}
			if nested, ok := c.nominalValue(value.fields[i]); ok {
				return nested, local, field.ID
			}
			return flowScalar(value.fields[i]), local, field.ID
		}
	}
	c.err = fmt.Errorf("FLOW_FIELD_UNRESOLVED")
	return flowScalar(0), 0, ""
}
