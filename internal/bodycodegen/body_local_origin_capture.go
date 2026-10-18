package bodycodegen

import "go/ast"

type localOriginCapture struct {
	projection       bodyProjection
	text, activityID string
	base             int
	variables        map[*ast.Object]int
	declarations     map[*ast.Object]BodyElementOrigin
	origins          map[ast.Node]BodyLocalOrigin
}

func (p bodyProjection) captureLocalOrigins(activityID, body string) map[ast.Node]BodyLocalOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	ids := map[string]string{p.function.Name.Name: activityID}
	for _, call := range p.calls {
		bodies[call.identity.Name], ids[call.identity.Name] = call.body, call.identity.ActivityID
	}
	origins := map[ast.Node]BodyLocalOrigin{}
	for _, declaration := range p.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		capture := localOriginCapture{projection: p, text: bodies[function.Name.Name], activityID: ids[function.Name.Name], base: p.fset.Position(function.Body.Lbrace).Offset + 2, variables: map[*ast.Object]int{}, declarations: map[*ast.Object]BodyElementOrigin{}, origins: origins}
		InspectBodyProjection(function.Body, func(node ast.Node) bool { capture.add(node); return true })
	}
	return origins
}

func (capture *localOriginCapture) span(node ast.Node) (BodyElementOrigin, bool) {
	start := capture.projection.fset.Position(node.Pos()).Offset - capture.base
	end := capture.projection.fset.Position(node.End()).Offset - capture.base
	if start < 0 || end <= start || end > len(capture.text) {
		return BodyElementOrigin{}, false
	}
	return BodyElementOrigin{start, end, capture.text[start:end]}, true
}

func (capture *localOriginCapture) add(node ast.Node) {
	write := retainedLocalWrite(node)
	if write.name == nil {
		return
	}
	value, valid := capture.span(write.value)
	if !valid {
		return
	}
	ordinal, found := capture.variables[write.name.Obj]
	if !found {
		// A first assignment to an existing parameter must not invent a local.
		if write.kind != "declaration" && write.kind != "iteration" && write.operation != ":=" {
			return
		}
		declaration, valid := capture.span(write.name)
		if !valid {
			return
		}
		ordinal = len(capture.variables)
		capture.variables[write.name.Obj] = ordinal
		capture.declarations[write.name.Obj] = declaration
	}
	capture.origins[node] = BodyLocalOrigin{ActivityID: capture.activityID, VariableOrdinal: ordinal, Name: write.name.Name, Kind: write.kind, Operation: write.operation, BodySHA256: digest([]byte(capture.text)), Start: value.Start, End: value.End, Expression: value.Expression, Declaration: capture.declarations[write.name.Obj]}
}

func (p bodyProjection) emittedLocalOrigins() []BodyLocalOrigin {
	var result []BodyLocalOrigin
	for _, declaration := range p.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ordinal := 0
		InspectBodyProjection(function.Body, func(node ast.Node) bool {
			write := localWrite(node)
			// Lowering creates identifiers without ast.Object links. Count their
			// emitted slots too: parsing the projection restores those links.
			if write.name == nil || write.name.Name == "_" {
				return true
			}
			if origin, found := p.locals[node]; found {
				origin.NativeOrdinal = ordinal
				result = append(result, origin)
			}
			ordinal++
			return true
		})
	}
	p.attachLocalUseOrigins(result)
	return result
}
