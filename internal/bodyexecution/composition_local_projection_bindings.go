package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionLocalUseSite struct {
	RootActivityID   string                            `json:"root_activity_id"`
	ActivityID       string                            `json:"activity_id"`
	VariableOrdinal  int                               `json:"variable_ordinal"`
	NativeOrdinal    int                               `json:"native_ordinal"`
	Name             string                            `json:"name"`
	ScalarType       string                            `json:"scalar_type"`
	ProjectionSHA256 string                            `json:"projection_sha256"`
	Start            int                               `json:"start_byte"`
	End              int                               `json:"end_byte"`
	Declaration      *bodycodegen.BodyExpressionSource `json:"declaration"`
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
		ids, ordinal := functions[function.Name.Name], 0
		objects := map[*ast.Object]int{}
		skipped := map[*ast.Ident]bool{}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			name, _, _, _ := bodycodegen.LocalWriteExpression(node)
			if name == nil {
				return true
			}
			skipped[name] = true
			index, found := lookup[conditionOriginKey{ids[0], ids[1], ordinal}]
			ordinal++
			if found && name.Obj != nil {
				objects[name.Obj] = index
				result.writes[node] = index
			}
			return true
		})
		variables := []int{}
		for _, index := range objects {
			variables = append(variables, sites[index].VariableOrdinal)
		}
		if len(variables) > 0 {
			slices.Sort(variables)
			result.scopes[function.Body] = slices.Compact(variables)
		}
		useOrdinal := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if assignment, ok := node.(*ast.AssignStmt); ok && len(assignment.Lhs) == 1 {
				if name, ok := assignment.Lhs[0].(*ast.Ident); ok && name.Name == "_" {
					return false // The emitter's unused-local marker is not a source use.
				}
			}
			if address, ok := node.(*ast.UnaryExpr); ok && address.Op == token.AND {
				if name, ok := address.X.(*ast.Ident); ok {
					skipped[name] = true // Taking an address does not read the scalar value.
				}
			}
			name, ok := node.(*ast.Ident)
			if !ok || name.Obj == nil || skipped[name] {
				return true
			}
			index, found := objects[name.Obj]
			if !found {
				return true
			}
			write := sites[index]
			result.uses[name] = len(result.sites)
			result.sites = append(result.sites, CompositionLocalUseSite{RootActivityID: ids[0], ActivityID: ids[1],
				VariableOrdinal: write.VariableOrdinal, NativeOrdinal: useOrdinal, Name: name.Name, ScalarType: write.ScalarType,
				ProjectionSHA256: digest([]byte(source)), Start: fset.Position(name.Pos()).Offset, End: fset.Position(name.End()).Offset,
				Declaration: write.Source.Declaration})
			useOrdinal++
			return true
		})
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
