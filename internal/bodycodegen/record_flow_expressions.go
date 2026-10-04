package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func (c *recordFlowBuilder) expression(node ast.Expr, state *flowState, selected bool) flowValue {
	if selected && c.altSet == nil {
		for i, site := range c.sites {
			span := c.span(node)
			if span.Start == site.start && span.End == site.end {
				return c.choice(i, node, state)
			}
		}
	}
	switch e := node.(type) {
	case *ast.ParenExpr:
		return c.expression(e.X, state, selected)
	case *ast.Ident:
		if binding := state.find(e.Name); binding != nil {
			return c.wrap(binding.value, c.span(e), "read", binding.id)
		}
		if e.Name == "true" || e.Name == "false" {
			return flowScalar(c.add("literal", [2]uint16{}, c.span(e), "", 0, e.Name))
		}
	case *ast.BasicLit:
		return flowScalar(c.add("literal", [2]uint16{}, c.span(e), "", 0, e.Kind.String()))
	case *ast.SelectorExpr:
		return c.selector(e, state)
	case *ast.BinaryExpr:
		return c.binary(e, state, selected)
	case *ast.UnaryExpr:
		a := c.expression(e.X, state, selected)
		return flowScalar(c.add("expression", [2]uint16{a.scalar}, c.span(e), "", 0, e.Op.String()))
	case *ast.CompositeLit:
		return c.constructor(e, state)
	}
	c.err = fmt.Errorf("FLOW_EXPRESSION_UNSUPPORTED")
	return flowScalar(0)
}

func (c *recordFlowBuilder) selector(e *ast.SelectorExpr, state *flowState) flowValue {
	value, local := flowScalar(0), uint16(0)
	if name, ok := assignmentLocalName(e.X); ok && state.find(name) != nil {
		binding := state.find(name)
		value, local = binding.value, binding.id
	} else {
		value = c.expression(e.X, state, true)
	}
	if value.record >= 0 {
		for i, field := range c.body.records[value.record].Fields {
			if field.Name == e.Sel.Name || field.GoName == e.Sel.Name {
				return flowScalar(c.add("read", [2]uint16{value.fields[i]}, c.span(e), field.ID, local, ""))
			}
		}
	}
	c.err = fmt.Errorf("FLOW_FIELD_UNRESOLVED")
	return flowScalar(0)
}

func (c *recordFlowBuilder) binary(e *ast.BinaryExpr, state *flowState, selected bool) flowValue {
	a, b := c.expression(e.X, state, selected), c.expression(e.Y, state, selected)
	if a.record < 0 && b.record < 0 {
		return flowScalar(c.add("expression", [2]uint16{a.scalar, b.scalar}, c.span(e), "", 0, e.Op.String()))
	}
	if a.record < 0 || a.record != b.record || (e.Op != token.EQL && e.Op != token.NEQ) {
		c.err = fmt.Errorf("FLOW_RECORD_OPERATOR_UNSUPPORTED")
		return flowScalar(0)
	}
	root, combine := uint16(0), "&&"
	if e.Op == token.NEQ {
		combine = "||"
	}
	for i, field := range c.body.records[a.record].Fields {
		part := c.add("expression", [2]uint16{a.fields[i], b.fields[i]}, c.span(e), field.ID, 0, e.Op.String())
		if root == 0 {
			root = part
		} else {
			root = c.add("expression", [2]uint16{root, part}, c.span(e), "", 0, combine)
		}
	}
	return flowScalar(root)
}

func (c *recordFlowBuilder) constructor(e *ast.CompositeLit, state *flowState) flowValue {
	value := flowScalar(0)
	name, ok := e.Type.(*ast.Ident)
	if ok {
		for i, record := range c.body.records {
			if record.Name == name.Name || record.GoName == name.Name {
				value.record = i
			}
		}
	}
	if value.record < 0 {
		c.err = fmt.Errorf("FLOW_RECORD_UNRESOLVED")
		return value
	}
	for _, element := range e.Elts {
		pair := element.(*ast.KeyValueExpr)
		for i, field := range c.body.records[value.record].Fields {
			if field.Name == pair.Key.(*ast.Ident).Name || field.GoName == pair.Key.(*ast.Ident).Name {
				value.fields[i] = c.expression(pair.Value, state, true).scalar
			}
		}
	}
	return value
}

func (c *recordFlowBuilder) choice(i int, node ast.Expr, state *flowState) flowValue {
	site := c.sites[i]
	first := c.expression(node, state, false).scalar
	fset := token.NewFileSet()
	other, err := parser.ParseExprFrom(fset, "alternative", site.choice.Second, parser.AllErrors)
	if err != nil {
		c.err = err
		return flowScalar(0)
	}
	c.altSet, c.view = fset, "alternative:"+site.choice.ID
	second := c.expression(other, state, false).scalar
	c.altSet, c.view = nil, c.bodyView
	span := c.span(node)
	result := c.add("choice", [2]uint16{first, second}, span, site.choice.FieldID, 0, site.choice.ID)
	c.choices[i] = RecordFlowChoice{ID: site.choice.ID, FieldID: site.choice.FieldID,
		First: first, Second: second, Result: result, Span: span, Guard: c.guardRoot}
	c.choiceSeen[i] = true
	return flowScalar(result)
}
