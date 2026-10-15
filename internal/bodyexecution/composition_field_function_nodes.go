package bodyexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
	"go/token"
)

func (p *compositionFieldProjection) collectFunctionWrites(fn *ast.FuncDecl, current compositionFieldFunction, nestedReturns map[*ast.ReturnStmt]bool, objects map[*ast.Object]int, functions map[string]int, fset *token.FileSet, source string) map[*ast.Ident]bool {
	skipped := map[*ast.Ident]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		name, value, _, _ := bodycodegen.LocalWriteExpression(n)
		if name != nil {
			skipped[name] = true
			if binding, found := objects[name.Obj]; found {
				p.collectLocalRecordWrite(n, value, binding, objects, functions, fset, source)
			}
		}
		if assign, ok := n.(*ast.AssignStmt); ok {
			if id, ok := singleAssignmentTarget(assign).(*ast.Ident); ok && id.Name == "_" {
				return false
			}
			p.collectFieldAssignment(assign, skipped, objects, fset, source)
		}

		if ret, ok := n.(*ast.ReturnStmt); ok && current.recordOut && !nestedReturns[ret] && len(ret.Results) == 1 {
			if binding, found := current.returns[ret]; found {
				p.collectReturnWrite(ret, binding, objects, functions, fset, source)
				return true
			}
			binding := -1
			if id, ok := ret.Results[0].(*ast.Ident); ok {
				if found, ok := objects[id.Obj]; ok {
					binding = found
				}
			}
			p.addNode(ret, ret.Results[0], "return", binding, 0, 0, fset)
		}
		return true
	})
	return skipped
}

func (p *compositionFieldProjection) collectFunctionReads(fn *ast.FuncDecl, skipped map[*ast.Ident]bool, objects map[*ast.Object]int, functions map[string]int, fset *token.FileSet) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 {
			if id, ok := assign.Lhs[0].(*ast.Ident); ok && id.Name == "_" {
				return false
			}
		}
		if selector, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := ast.Unparen(selector.X).(*ast.Ident); ok {
				if binding, found := objects[id.Obj]; found {
					if !skipped[id] {
						for field, descriptor := range p.inventory.bindings[binding].Record.Fields {
							if descriptor.GoName == selector.Sel.Name {
								first := len(p.uses)
								p.uses = append(p.uses, CompositionFieldSite{Binding: binding, Field: field, Kind: "field_read", Start: fset.Position(n.Pos()).Offset, End: fset.Position(n.End()).Offset})
								p.addNode(n, n, "use", binding, first, field, fset)
							}
						}
					}
					return false
				}
			}
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok {
				if target, found := functions[id.Name]; found && len(p.functions[target].parameters) > 0 {
					i := p.addNode(n, n, "call", 0, 0, 0, fset)
					for _, port := range p.functions[target].paramSlots {
						binding := -1
						if port < len(call.Args) {
							if arg, ok := call.Args[port].(*ast.Ident); ok {
								if found, ok := objects[arg.Obj]; ok {
									binding = found
								}
							}
						}
						p.nodes[i].parents = append(p.nodes[i].parents, binding)
					}
				}
			}
		}
		if id, ok := n.(*ast.Ident); ok && !skipped[id] {
			if binding, found := objects[id.Obj]; found {
				first := len(p.uses)
				for field := range p.inventory.bindings[binding].Record.Fields {
					p.uses = append(p.uses, CompositionFieldSite{Binding: binding, Field: field, Kind: "record_read", Start: fset.Position(id.Pos()).Offset, End: fset.Position(id.End()).Offset})
				}
				p.addNode(n, n, "record_use", binding, first, 0, fset)
			}
		}
		return true
	})
}
