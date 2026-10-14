package bodyexecution

import (
	"go/ast"
	"go/token"
)

func (p *compositionFieldProjection) collectReturnWrite(statement *ast.ReturnStmt, binding int,
	objects map[*ast.Object]int, functions map[string]int, fset *token.FileSet, source string) {
	value := statement.Results[0]
	first := p.addSites(binding, "record_assignment", "shared_rhs", value, fset, source)
	index := p.addNode(statement, value, "return_write", binding, first, 0, fset)
	p.nodes[index].parents = []int{-1}
	switch expression := unparenthesizedFieldValue(value).(type) {
	case *ast.CompositeLit:
		p.collectReturnParts(expression, binding, first, index, fset, source)
	case *ast.Ident:
		if parent, found := objects[expression.Obj]; found {
			p.nodes[index].parents[0], p.nodes[index].mode = parent, 1
			for field := range p.inventory.bindings[binding].Record.Fields {
				p.sites[first+field].DependencyMode = "field_copy"
			}
		}
	case *ast.CallExpr:
		if name, ok := expression.Fun.(*ast.Ident); ok {
			if called, found := functions[name.Name]; found && p.functions[called].recordOut {
				p.nodes[index].mode = 2
				for field := range p.inventory.bindings[binding].Record.Fields {
					p.sites[first+field].DependencyMode = "record_return"
				}
			}
		}
	}
}

func (p *compositionFieldProjection) collectReturnParts(literal *ast.CompositeLit, binding, first, node int,
	fset *token.FileSet, source string) {
	matched := 0
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok {
			continue
		}
		for field, descriptor := range p.inventory.bindings[binding].Record.Fields {
			if descriptor.GoName == key.Name {
				p.addNode(pair, pair.Value, "part", binding, first+field, field, fset)
				start, end := fset.Position(pair.Value.Pos()).Offset, fset.Position(pair.Value.End()).Offset
				p.sites[first+field] = CompositionFieldSite{Binding: binding, Field: field, Kind: "record_constructor",
					DependencyMode: "scalar_rhs", Start: start, End: end, Expression: source[start:end]}
				matched++
			}
		}
	}
	if matched == len(p.inventory.bindings[binding].Record.Fields) {
		p.nodes[node].mode = 3
	}
}
