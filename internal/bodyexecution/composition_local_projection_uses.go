package bodyexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
	"go/token"
)

func localFunctionSourceUses(ids [2]string, sites []CompositionLocalSite) map[int]map[int]*bodycodegen.BodyExpressionSource {
	originalUses := map[int]map[int]*bodycodegen.BodyExpressionSource{}
	for _, site := range sites {
		if site.RootActivityID != ids[0] || site.ActivityID != ids[1] {
			continue
		}
		if originalUses[site.VariableOrdinal] == nil {
			originalUses[site.VariableOrdinal] = map[int]*bodycodegen.BodyExpressionSource{}
		}
		for _, use := range site.Source.Uses {
			originalUses[site.VariableOrdinal][use.NativeOrdinal] = use.Source
		}
	}

	return originalUses
}
func bindLocalFunctionUses(function *ast.FuncDecl, fset *token.FileSet, source string, ids [2]string, objects map[*ast.Object]int, sites []CompositionLocalSite, originalUses map[int]map[int]*bodycodegen.BodyExpressionSource, result *localProjectionBindings) {
	useOrdinal := 0
	variableOrdinals := map[int]int{}
	bodycodegen.WalkLocalValueUses(function.Body, func(name *ast.Ident) {
		index, found := objects[name.Obj]
		if !found {
			return
		}
		write := sites[index]
		ordinal := variableOrdinals[write.VariableOrdinal]
		variableOrdinals[write.VariableOrdinal]++
		result.uses[name] = len(result.sites)
		result.sites = append(result.sites, CompositionLocalUseSite{RootActivityID: ids[0], ActivityID: ids[1],
			VariableOrdinal: write.VariableOrdinal, NativeOrdinal: useOrdinal, Name: name.Name, ScalarType: write.ScalarType,
			ProjectionSHA256: digest([]byte(source)), Start: fset.Position(name.Pos()).Offset, End: fset.Position(name.End()).Offset,
			Declaration: write.Source.Declaration, VariableUseOrdinal: ordinal, Source: originalUses[write.VariableOrdinal][ordinal]})
		useOrdinal++
	})
}
