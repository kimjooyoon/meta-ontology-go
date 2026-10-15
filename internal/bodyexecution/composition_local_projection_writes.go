package bodyexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
	"slices"
)

func bindLocalFunctionWrites(function *ast.FuncDecl, ids [2]string, lookup map[conditionOriginKey]int, result *localProjectionBindings) map[*ast.Object]int {
	ordinal := 0
	objects := map[*ast.Object]int{}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		name, _, _, _ := bodycodegen.LocalWriteExpression(node)
		if name == nil {
			return true
		}
		index, found := lookup[conditionOriginKey{ids[0], ids[1], ordinal}]
		ordinal++
		if found && name.Obj != nil {
			objects[name.Obj] = index
			result.writes[node] = index
		}
		return true
	})

	return objects
}
func bindLocalFunctionScope(function *ast.FuncDecl, objects map[*ast.Object]int, sites []CompositionLocalSite, result *localProjectionBindings) {
	variables := []int{}
	for _, index := range objects {
		variables = append(variables, sites[index].VariableOrdinal)
	}
	if len(variables) > 0 {
		slices.Sort(variables)
		result.scopes[function.Body] = slices.Compact(variables)
	}
}
