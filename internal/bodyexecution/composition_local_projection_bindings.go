package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionLocalUseSite struct {
	RootActivityID     string                            `json:"root_activity_id"`
	ActivityID         string                            `json:"activity_id"`
	VariableOrdinal    int                               `json:"variable_ordinal"`
	NativeOrdinal      int                               `json:"native_ordinal"`
	Name               string                            `json:"name"`
	ScalarType         string                            `json:"scalar_type"`
	ProjectionSHA256   string                            `json:"projection_sha256"`
	Start              int                               `json:"start_byte"`
	End                int                               `json:"end_byte"`
	Declaration        *bodycodegen.BodyExpressionSource `json:"declaration"`
	VariableUseOrdinal int                               `json:"variable_use_ordinal"`
	Source             *bodycodegen.BodyExpressionSource `json:"source,omitempty"`
}

type localProjectionBindings struct {
	writes map[ast.Node]int
	uses   map[*ast.Ident]int
	scopes map[*ast.BlockStmt][]int
	sites  []CompositionLocalUseSite
}

func bindLocalProjection(file *ast.File, fset *token.FileSet, source string, prior Composition, sites []CompositionLocalSite) (localProjectionBindings, error) {
	result := localProjectionBindings{writes: map[ast.Node]int{}, uses: map[*ast.Ident]int{}, scopes: map[*ast.BlockStmt][]int{}}
	lookup := map[conditionOriginKey]int{}
	for index, site := range sites {
		lookup[conditionOriginKey{site.RootActivityID, site.ActivityID, site.NativeOrdinal}] = index
	}
	functions := arithmeticFunctionIDs(prior)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids := functions[function.Name.Name]
		objects := bindLocalFunctionWrites(function, ids, lookup, &result)
		bindLocalFunctionScope(function, objects, sites, &result)
		originalUses := localFunctionSourceUses(ids, sites)
		bindLocalFunctionUses(function, fset, source, ids, objects, sites, originalUses, &result)
	}
	if len(result.writes) != len(sites) || len(result.sites) > 65536 {
		return result, fmt.Errorf("local projection inventory or use-site bound differs")
	}
	return result, nil
}

func compositionLocalUseSites(prior Composition, sites []CompositionLocalSite) ([]CompositionLocalUseSite, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "composition.go", prior.Source, 0)
	if err != nil {
		return nil, err
	}
	bindings, err := bindLocalProjection(file, fset, prior.Source, prior, sites)
	return bindings.sites, err
}
