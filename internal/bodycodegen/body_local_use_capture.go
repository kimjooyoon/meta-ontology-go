package bodycodegen

import (
	"go/ast"
	"go/token"
)

// WalkLocalValueUses shares the source and native traversal. Assignment targets,
// direct address-taking and compiler unused-local markers are not value reads.
func WalkLocalValueUses(body *ast.BlockStmt, visit func(*ast.Ident)) {
	skipped := map[*ast.Ident]bool{}
	ast.Inspect(body, func(node ast.Node) bool {
		name, _, _, _ := LocalWriteExpression(node)
		_, update := node.(*ast.IncDecStmt)
		if name != nil && !update {
			skipped[name] = true
		}
		return true
	})
	ast.Inspect(body, func(node ast.Node) bool {
		if a, ok := node.(*ast.AssignStmt); ok && len(a.Lhs) == 1 {
			if name, ok := a.Lhs[0].(*ast.Ident); ok && name.Name == "_" {
				return false
			}
		}
		if address, ok := node.(*ast.UnaryExpr); ok && address.Op == token.AND {
			if name, ok := address.X.(*ast.Ident); ok {
				skipped[name] = true
			}
		}
		if name, ok := node.(*ast.Ident); ok && name.Obj != nil && !skipped[name] {
			visit(name)
		}
		return true
	})
}

func (p bodyProjection) captureLocalUses(body string) map[*ast.Ident]BodyElementOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	for _, call := range p.calls {
		bodies[call.identity.Name] = call.body
	}
	result := map[*ast.Ident]BodyElementOrigin{}
	for _, declaration := range p.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		capture := localOriginCapture{projection: p, text: bodies[function.Name.Name],
			base: p.fset.Position(function.Body.Lbrace).Offset + 2}
		WalkLocalValueUses(function.Body, func(name *ast.Ident) {
			if span, ok := capture.span(name); ok {
				result[name] = span
			}
		})
	}
	return result
}

func (p bodyProjection) attachLocalUseOrigins(origins []BodyLocalOrigin) {
	type key struct {
		activity string
		variable int
	}
	first := map[key]int{}
	for i, origin := range origins {
		k := key{origin.ActivityID, origin.VariableOrdinal}
		if _, found := first[k]; !found {
			first[k] = i
		}
	}
	for _, declaration := range p.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		variables := map[*ast.Object]int{}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if origin, found := p.locals[node]; found {
				name, _, _, _ := LocalWriteExpression(node)
				if name != nil && name.Obj != nil {
					variables[name.Obj] = first[key{origin.ActivityID, origin.VariableOrdinal}]
				}
			}
			return true
		})
		ordinals := map[*ast.Object]int{}
		WalkLocalValueUses(function.Body, func(name *ast.Ident) {
			index, found := variables[name.Obj]
			if !found {
				return
			}
			ordinal := ordinals[name.Obj]
			ordinals[name.Obj]++
			if span, captured := p.localUses[name]; captured {
				origins[index].Uses = append(origins[index].Uses, BodyLocalUseOrigin{NativeOrdinal: ordinal, Start: span.Start, End: span.End, Expression: span.Expression})
			}
		})
	}
}
