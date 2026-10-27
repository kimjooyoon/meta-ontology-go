package bodyexecution

import (
	"go/ast"
	"go/token"
)

func (p *compositionFieldProjection) collectLocalRecordWrite(n ast.Node, value ast.Node, binding int, objects map[*ast.Object]int, functions map[string]int, fset *token.FileSet, source string) {
	if loop, ok := n.(*ast.RangeStmt); ok {
		first := p.addSites(binding, "record_assignment", "shared_rhs", loop.Value, fset, source)
		p.addNode(loop, loop.Body, "range_record_write", binding, first, 0, fset)
		return
	}
	first := p.addSites(binding, "record_assignment", "shared_rhs", value, fset, source)
	i := p.addNode(n, value, "record_write", binding, first, 0, fset)
	p.nodes[i].parents = []int{-1}
	if literal, ok := value.(*ast.CompositeLit); ok {
		p.collectReturnParts(literal, binding, first, i, fset, source)
	}
	if id, ok := value.(*ast.Ident); ok {
		if parent, found := objects[id.Obj]; found {
			p.nodes[i].parents[0], p.nodes[i].mode = parent, 1
			for j := range p.inventory.bindings[binding].Record.Fields {
				p.sites[first+j].DependencyMode = "field_copy"
			}
		}
	}
	if call, ok := value.(*ast.CallExpr); ok {
		if target, ok := call.Fun.(*ast.Ident); ok {
			if called, found := functions[target.Name]; found && p.functions[called].recordOut {
				p.nodes[i].mode = 2
				for j := range p.inventory.bindings[binding].Record.Fields {
					p.sites[first+j].DependencyMode = "record_return"
				}
			}
		}
	}
}
