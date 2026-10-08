package bodycodegen

import (
	"fmt"
	"go/ast"
	"strings"
)

type recordFlowHelper struct {
	function pureCallFunction
	body     *ast.BlockStmt
	base     int
}

type recordFlowReturns struct {
	value flowValue
	guard uint16
	seen  bool
}

// Resolve the same fixed source closure used by native projection. Alternative
// calls belong to the analysis even when they are absent from the baseline.
func (c *recordFlowBuilder) sourcePrelude() (string, string, error) {
	closureSource := c.body.body
	for _, site := range c.sites {
		closureSource += "\n_ = (" + site.choice.Second + ")"
	}
	functions, _, _, err := c.body.resolvePureCalls(closureSource)
	if err != nil {
		return "", "", err
	}
	if len(functions) > 0 {
		c.body.records = c.body.allRecords
	}
	var source strings.Builder
	source.WriteString("package selection\n" + RecordDeclarations(c.body.records, false))
	c.helpers = make(map[string]recordFlowHelper, len(functions))
	for _, function := range functions {
		prefix := fmt.Sprintf("func %s(%s) %s {\n", function.identity.Name,
			parameterDeclaration(function.parameters), function.output)
		c.helpers[function.identity.Name] = recordFlowHelper{function: function, base: source.Len() + len(prefix)}
		c.helperOrder = append(c.helperOrder, function.identity.Name)
		source.WriteString(prefix + function.body + "\n}\n")
	}
	root := "selected"
	for c.helpers[root].function.identity.ActivityID != "" {
		root = "_" + root
	}
	return source.String(), root, nil
}

func (c *recordFlowBuilder) call(call *ast.CallExpr, state *flowState, selected bool) flowValue {
	helper := c.helpers[call.Fun.(*ast.Ident).Name]
	if len(call.Args) != len(helper.function.parameters) || c.callDepth >= 16 {
		c.err = fmt.Errorf("FLOW_CALL_BOUND_OR_ARITY")
		return flowScalar(0)
	}
	var arguments [16]flowValue
	for i, argument := range call.Args {
		arguments[i] = c.expression(argument, state, selected)
	}
	value := c.expandHelper(helper, arguments, state.count)
	if c.err != nil {
		return flowScalar(0)
	}
	value = c.wrap(value, c.span(call), "call", 0)
	if value.record < 0 {
		c.markCallee(value.scalar, helper.function.identity.ActivityID)
	} else {
		for i := range c.body.records[value.record].Fields {
			c.markCallee(value.fields[i], helper.function.identity.ActivityID)
		}
	}
	return value
}

func (c *recordFlowBuilder) markCallee(root uint16, activity string) {
	if root != 0 {
		c.nodes[root-1].CalleeID = activity
	}
}

func (c *recordFlowBuilder) expandHelper(helper recordFlowHelper, arguments [16]flowValue, callerBindings int) flowValue {
	base, view, activity, alt, guard := c.base, c.view, c.activityID, c.altSet, c.guardRoot
	prior, suppressed, outer := c.returns, c.returnSuppressed, c.outerBindings
	defer func() {
		c.base, c.view, c.activityID, c.altSet, c.guardRoot = base, view, activity, alt, guard
		c.returns, c.returnSuppressed, c.outerBindings = prior, suppressed, outer
		c.callDepth--
	}()
	c.callDepth++
	c.base, c.view, c.activityID, c.altSet = helper.base, "computes", helper.function.identity.ActivityID, nil
	c.outerBindings += callerBindings
	result := recordFlowReturns{guard: guard}
	c.returns, c.returnSuppressed = &result, false
	state := flowState{guard: guard}
	for i, parameter := range helper.function.parameters {
		span := RecordFlowSpan{ActivityID: c.activityID, View: "input-port"}
		c.bind(&state, parameter.Name, arguments[i], span, "call_parameter")
	}
	if c.block(helper.body, &state) || !result.seen {
		if c.err == nil {
			c.err = fmt.Errorf("FLOW_CALL_RETURN_UNRESOLVED")
		}
	}
	return result.value
}

// A callee's returns are alternatives selected by their execution guards.
// Constant-unreachable returns cannot contribute to the returned value.
func (c *recordFlowBuilder) retainReturn(value flowValue, span RecordFlowSpan) {
	if c.returns == nil || c.returnSuppressed || c.err != nil {
		return
	}
	if !c.returns.seen {
		c.returns.value, c.returns.seen = value, true
		return
	}
	guard := c.guardRoot
	c.guardRoot = c.returns.guard
	defer func() { c.guardRoot = guard }()
	previous := c.returns.value
	if value.record < 0 {
		value.scalar = c.join(value.scalar, previous.scalar, span, "", 0, guard)
	} else {
		for i, field := range c.body.records[value.record].Fields {
			value.fields[i] = c.join(value.fields[i], previous.fields[i], span, field.ID, 0, guard)
		}
	}
	c.returns.value = value
}
