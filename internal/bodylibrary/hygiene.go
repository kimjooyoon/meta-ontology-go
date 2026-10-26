package bodylibrary

import (
	"fmt"
	"go/ast"
	"go/types"
)

// A readable exported type or a private helper must not capture a source local.
// Allocate local names in AST order, outside all existing and introduced names.
func avoidLibraryShadowing(file *ast.File, info *types.Info, names map[types.Object]string) {
	globals, used := map[string]bool{}, map[string]bool{}
	for object, name := range names {
		used[name] = true
		if _, field := object.(*types.Var); !field {
			globals[name] = true
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		if name, ok := node.(*ast.Ident); ok {
			used[name.Name] = true
		}
		return true
	})
	ordinal := 0
	ast.Inspect(file, func(node ast.Node) bool {
		name, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		variable, ok := info.Defs[name].(*types.Var)
		if !ok || variable.IsField() || !globals[name.Name] {
			return true
		}
		for {
			candidate := fmt.Sprintf("goooLibraryLocal%d", ordinal)
			ordinal++
			if !used[candidate] {
				names[variable], used[candidate] = candidate, true
				return true
			}
		}
	})
}
