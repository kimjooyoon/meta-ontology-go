package bodycodegen

import "go/ast"

type bodyReturnCapture struct {
	origins    map[*ast.ReturnStmt]BodyReturnOrigin
	activities map[string]string
}

func (p bodyProjection) captureReturnOrigins(activityID, body string) *bodyReturnCapture {
	bodies := map[string]string{p.function.Name.Name: body}
	result := &bodyReturnCapture{origins: map[*ast.ReturnStmt]BodyReturnOrigin{}, activities: map[string]string{p.function.Name.Name: activityID}}
	for _, call := range p.calls {
		bodies[call.identity.Name], result.activities[call.identity.Name] = call.body, call.identity.ActivityID
	}
	for _, declaration := range p.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		capture := localOriginCapture{projection: p, text: bodies[function.Name.Name], base: p.fset.Position(function.Body.Lbrace).Offset + 2}
		WalkReturnSyntax(function.Body, func(statement *ast.ReturnStmt) {
			if value, valid := capture.span(statement.Results[0]); valid {
				result.origins[statement] = BodyReturnOrigin{ActivityID: result.activities[function.Name.Name],
					BodySHA256: digest([]byte(capture.text)), Value: &value}
			}
		})
	}
	return result
}

func (c *bodyReturnCapture) emitted(file *ast.File) []BodyReturnOrigin {
	var result []BodyReturnOrigin
	if c == nil {
		return result
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ordinal, start := 0, len(result)
		WalkReturnSyntax(function.Body, func(statement *ast.ReturnStmt) {
			if origin, found := c.origins[statement]; found && origin.RecordID != "" {
				origin.NativeOrdinal = ordinal
				result = append(result, origin)
			}
			ordinal++
		})
		for i := start; i < len(result); i++ {
			result[i].NativePopulation = ordinal
		}
	}
	return result
}
