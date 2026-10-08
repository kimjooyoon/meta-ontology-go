package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

func recordValueFlow(p recordAssemblyPlan) *RecordValueFlow {
	view := "computes"
	if p.spec.Baseline != "" {
		view = "baseline"
	}
	r := &RecordValueFlow{Schema: recordFlowSchema, Status: "RESOLVED", ActivityID: p.body.activityID,
		BodyView: view, BodySHA256: digest([]byte(p.body.body)), Scope: "typed symbolic value definitions, copies, choice alternatives and conditional joins; body-relative byte offsets retain same-length let/var spelling; no input values, cases, predictions or candidate execution"}
	c := recordFlowBuilder{body: p.body, sites: p.sites, fset: token.NewFileSet(), view: view, bodyView: view,
		info: &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}}
	fn, err := c.parse()
	if err == nil {
		state := c.parameters()
		c.block(fn.Body, &state)
		err = c.err
	}
	for i := range p.sites {
		if err == nil && !c.choiceSeen[i] {
			err = fmt.Errorf("FLOW_CHOICE_UNREACHED")
		}
	}
	if err != nil {
		r.Status, r.Reason = "UNRESOLVED", err.Error()
		return r
	}
	r.Nodes = append([]RecordFlowNode(nil), c.nodes[:c.count]...)
	r.Choices = append([]RecordFlowChoice(nil), c.choices[:len(p.sites)]...)
	for _, name := range c.helperOrder {
		r.Helpers = append(r.Helpers, c.helpers[name].function.identity)
	}
	if len(r.Helpers) > 0 {
		r.Scope += "; fixed same-source helpers retain argument bindings, guarded returns, callee identities and helper-relative spans"
	}
	return r
}

func (c *recordFlowBuilder) parse() (*ast.FuncDecl, error) {
	prefix, root, err := c.sourcePrelude()
	if err != nil {
		return nil, err
	}
	prefix += "func " + root + "(" + parameterDeclaration(c.body.parameters) + ") " + c.body.outputType + "{\n"
	c.base = len(prefix)
	file, err := parser.ParseFile(c.fset, "record-flow", prefix+c.body.body+"\n}", parser.AllErrors)
	if err != nil {
		return nil, err
	}
	normalizeIntegerLocalInitializers("selection", file, c.fset)
	if _, err = checkBodyTypes("selection", file, c.fset, c.info); err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Name.Name != root {
			helper := c.helpers[fn.Name.Name]
			helper.body = fn.Body
			c.helpers[fn.Name.Name] = helper
		}
	}
	return file.Decls[len(file.Decls)-1].(*ast.FuncDecl), nil
}

func (c *recordFlowBuilder) add(kind string, parents [2]uint16, span RecordFlowSpan, field string, local uint16, op string) uint16 {
	if c.err != nil {
		return 0
	}
	if c.count == len(c.nodes) {
		c.err = fmt.Errorf("FLOW_NODE_BOUND")
		return 0
	}
	id := uint16(c.count + 1)
	c.nodes[c.count] = RecordFlowNode{ID: id, Kind: kind, Parents: parents, Span: span, FieldID: field, Local: local, Operator: op, Guard: c.guardRoot}
	c.count++
	return id
}

func (c *recordFlowBuilder) bind(state *flowState, name string, value flowValue, span RecordFlowSpan, kind string) {
	if state.count+c.outerBindings >= len(state.bindings) {
		c.err = fmt.Errorf("FLOW_LOCAL_BOUND")
		return
	}
	c.locals++
	value = c.wrap(value, span, kind, c.locals)
	state.bindings[state.count] = flowBinding{name: name, id: c.locals, value: value}
	state.count++
}

func (c *recordFlowBuilder) wrap(value flowValue, span RecordFlowSpan, kind string, local uint16) flowValue {
	if value.record < 0 {
		value.scalar = c.add(kind, [2]uint16{value.scalar}, span, "", local, "")
		return value
	}
	for i, field := range c.body.records[value.record].Fields {
		value.fields[i] = c.add(kind, [2]uint16{value.fields[i]}, span, field.ID, local, "")
	}
	return value
}

func (c *recordFlowBuilder) parameters() flowState {
	var state flowState
	for _, parameter := range c.body.parameters {
		value := flowScalar(0)
		for i, record := range c.body.records {
			if record.Name == parameter.Type || record.GoName == parameter.Type {
				value.record = i
			}
		}
		c.bind(&state, parameter.Name, value, RecordFlowSpan{View: "input-port"}, "input")
	}
	return state
}
