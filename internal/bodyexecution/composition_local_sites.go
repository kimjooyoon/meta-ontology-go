package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// A local site names an authored write; one loop site can execute many times.
// Source declarations identify lexical bindings, not equality of scalar values.
type CompositionLocalSite struct {
	ExpressionID     string                      `json:"expression_id"`
	RootActivityID   string                      `json:"root_activity_id"`
	ActivityID       string                      `json:"activity_id"`
	NativeOrdinal    int                         `json:"native_ordinal"`
	VariableOrdinal  int                         `json:"variable_ordinal"`
	Name             string                      `json:"name"`
	Kind             string                      `json:"kind"`
	Operation        string                      `json:"operation"`
	ScalarType       string                      `json:"scalar_type"`
	Expression       string                      `json:"expression"`
	ProjectionSHA256 string                      `json:"projection_sha256"`
	Start            int                         `json:"start_byte"`
	End              int                         `json:"end_byte"`
	Source           bodycodegen.BodyLocalSource `json:"source"`
}

type compositionLocalOrigin struct {
	definition bodycodegen.BodyLocalOrigin
	source     bodycodegen.BodyLocalSource
}

func compositionSourceLocals(prior Composition) (map[conditionOriginKey]compositionLocalOrigin, error) {
	var origins []bodycodegen.BodyLocalOrigin
	var keys []conditionOriginKey
	for _, step := range prior.Steps {
		for _, origin := range step.Generation.Report.BodyLocals {
			origins = append(origins, origin)
			keys = append(keys, conditionOriginKey{step.Generation.Report.ActivityID, origin.ActivityID, origin.NativeOrdinal})
		}
	}
	if len(origins) == 0 {
		return map[conditionOriginKey]compositionLocalOrigin{}, nil
	}
	bound, err := bodycodegen.BindBodyLocalSources([]byte(prior.GoooSource), origins)
	if err != nil {
		return nil, err
	}
	result := map[conditionOriginKey]compositionLocalOrigin{}
	for index, key := range keys {
		if _, duplicate := result[key]; duplicate {
			return nil, fmt.Errorf("local write origin has duplicate activity/ordinal identity")
		}
		result[key] = compositionLocalOrigin{origins[index], bound[index]}
	}
	return result, nil
}

func compositionLocalSites(prior Composition) ([]CompositionLocalSite, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "composition.go", prior.Source, 0)
	if err != nil {
		return nil, err
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	if _, err := new(types.Config).Check("gooo.observed.locals", fset, []*ast.File{file}, info); err != nil {
		return nil, err
	}
	origins, err := compositionSourceLocals(prior)
	if err != nil {
		return nil, err
	}
	bindings, sha := arithmeticFunctionIDs(prior), digest([]byte(prior.Source))
	matched := map[conditionOriginKey]bool{}
	objects := map[types.Object]conditionOriginKey{}
	variables := map[conditionOriginKey]types.Object{}
	var sites []CompositionLocalSite
	var mismatch error
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids, ordinal := bindings[function.Name.Name], 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			name, expression, kind, operation := bodycodegen.LocalWriteExpression(node)
			if name == nil {
				return true
			}
			key := conditionOriginKey{ids[0], ids[1], ordinal}
			origin, found := origins[key]
			current := ordinal
			ordinal++
			if !found {
				return true
			}
			matched[key] = true
			if origin.definition.Name != name.Name || origin.definition.Kind != kind || origin.definition.Operation != operation {
				mismatch = fmt.Errorf("local write origin differs from emitted projection")
				return false
			}
			object := info.ObjectOf(name)
			if object == nil {
				mismatch = fmt.Errorf("local write has no native lexical binding")
				return false
			}
			variable := conditionOriginKey{ids[0], ids[1], origin.definition.VariableOrdinal}
			if previous, found := objects[object]; (found && previous != variable) ||
				(variables[variable] != nil && variables[variable] != object) || origin.definition.VariableOrdinal < 0 {
				mismatch = fmt.Errorf("local variable identity differs from native lexical binding")
				return false
			}
			objects[object], variables[variable] = variable, object
			scalar := object.Type().String()
			if scalar != "int64" && scalar != "bool" && scalar != "string" {
				return true
			}
			start, end := fset.Position(expression.Pos()).Offset, fset.Position(expression.End()).Offset
			site := CompositionLocalSite{RootActivityID: ids[0], ActivityID: ids[1], NativeOrdinal: current,
				VariableOrdinal: origin.definition.VariableOrdinal, Name: name.Name, Kind: kind, Operation: operation,
				ScalarType: scalar, Expression: prior.Source[start:end], ProjectionSHA256: sha, Start: start, End: end,
				Source: origin.source}
			site.ExpressionID = compositionDigest(site)
			sites = append(sites, site)
			return true
		})
	}
	if mismatch != nil {
		return nil, mismatch
	}
	if len(matched) != len(origins) {
		return nil, fmt.Errorf("retained local write origin is missing from native projection")
	}
	if len(sites) > 4096 {
		return nil, fmt.Errorf("scalar local writes exceed 4096 sites")
	}
	return sites, nil
}
