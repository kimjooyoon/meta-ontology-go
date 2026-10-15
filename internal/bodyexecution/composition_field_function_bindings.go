package bodyexecution

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
	"go/token"
	"slices"
	"strings"
)

func (p *compositionFieldProjection) bindFunction(fn *ast.FuncDecl, ids [2]string, objects map[*ast.Object]int, fset *token.FileSet, source, ignorePrefix string) (compositionFieldFunction, error) {
	inventory := p.inventory
	current := compositionFieldFunction{name: fn.Name.Name, returns: map[*ast.ReturnStmt]int{}}
	p.bindParameters(fn, ids, &current, objects, fset, source)
	if err := p.bindLocals(fn, ids, &current, objects, ignorePrefix); err != nil {
		return current, err
	}
	for index, statement := range fieldFunctionReturns(fn) {
		if binding, found := inventory.returns[conditionOriginKey{ids[0], ids[1], index}]; found {
			current.bindings = append(current.bindings, binding)
			current.returns[statement] = binding
		}
	}
	slices.Sort(current.bindings)
	current.bindings = slices.Compact(current.bindings)
	if fn.Type.Results != nil && len(fn.Type.Results.List) == 1 {
		if name, ok := fn.Type.Results.List[0].Type.(*ast.Ident); ok {
			for _, binding := range inventory.bindings {
				current.recordOut = current.recordOut || binding.Record.GoName == name.Name
			}
		}
	}
	return current, nil
}

func (p *compositionFieldProjection) bindParameters(fn *ast.FuncDecl, ids [2]string, current *compositionFieldFunction, objects map[*ast.Object]int, fset *token.FileSet, source string) {
	inventory := p.inventory
	port := 0
	for _, group := range fn.Type.Params.List {
		for _, name := range group.Names {
			if index, found := inventory.inputs[conditionOriginKey{ids[0], ids[1], -port - 1}]; found {
				objects[name.Obj] = index
				current.bindings = append(current.bindings, index)
				current.parameters = append(current.parameters, index)
				current.paramSlots = append(current.paramSlots, port)
				current.paramSites = append(current.paramSites, p.addSites(index, "parameter", "call_parameter_or_input", name, fset, source))
			}
			port++
		}
	}
}

func (p *compositionFieldProjection) bindLocals(fn *ast.FuncDecl, ids [2]string, current *compositionFieldFunction, objects map[*ast.Object]int, ignorePrefix string) error {
	inventory := p.inventory
	var mismatch error
	ordinal := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		name, _, _, _ := bodycodegen.LocalWriteExpression(n)
		if name == nil || ignorePrefix != "" && strings.HasPrefix(name.Name, ignorePrefix) {
			return true
		}
		index, found := inventory.locals[conditionOriginKey{ids[0], ids[1], ordinal}]
		ordinal++
		if found && inventory.bindings[index].Name != name.Name {
			mismatch = fmt.Errorf("field lexical projection identity differs")
			return false
		}
		if found && name.Obj != nil {
			objects[name.Obj] = index
			current.bindings = append(current.bindings, index)
		}
		return true
	})
	return mismatch
}
