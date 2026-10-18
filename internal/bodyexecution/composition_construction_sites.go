package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func constructionLiteral(n ast.Node) *ast.CompositeLit {
	literal, ok := n.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	array, ok := literal.Type.(*ast.ArrayType)
	if !ok || array.Len != nil {
		return nil
	}
	return literal
}

func compositionSourceConstructions(prior Composition) (map[conditionOriginKey]bodycodegen.BodyConstructionSource, error) {
	var origins []bodycodegen.BodyConstructionOrigin
	var keys []conditionOriginKey
	for _, step := range prior.Steps {
		for _, o := range step.Generation.Report.BodyConstructions {
			origins = append(origins, o)
			keys = append(keys, conditionOriginKey{step.Generation.Report.ActivityID, o.ActivityID, o.NativeOrdinal})
		}
	}
	result := map[conditionOriginKey]bodycodegen.BodyConstructionSource{}
	if len(origins) == 0 {
		return result, nil
	}
	bound, err := bodycodegen.BindBodyConstructionSources([]byte(prior.GoooSource), origins)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		result[key] = bound[i]
	}
	return result, nil
}

func compositionConstructionSites(prior Composition) ([]CompositionConstructionSite, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "composition.go", prior.Source, 0)
	if err != nil {
		return nil, err
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	if _, err := new(types.Config).Check("gooo.observed.constructions", fset, []*ast.File{file}, info); err != nil {
		return nil, err
	}
	origins, err := compositionSourceConstructions(prior)
	if err != nil {
		return nil, err
	}
	bindings, sha := arithmeticFunctionIDs(prior), digest([]byte(prior.Source))
	var sites []CompositionConstructionSite
	var bindingError error
	for _, declaration := range file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids, ordinal := bindings[f.Name.Name], 0
		bodycodegen.InspectBodyProjection(f.Body, func(n ast.Node) bool {
			literal := constructionLiteral(n)
			var expression ast.Expr
			var elements []ast.Expr
			kind, mapperID := "", ""
			if literal != nil {
				expression, elements = literal, literal.Elts
			} else if call, ok := n.(*ast.CallExpr); ok {
				_, mapper, mapped := bodycodegen.CollectionMapProjection(call)
				if !mapped {
					return true
				}
				expression, elements, kind, mapperID = call, []ast.Expr{mapper}, "map_values", bindings[mapper.Name][1]
			} else {
				return true
			}
			current := ordinal
			ordinal++
			sequence, ok := info.TypeOf(expression).Underlying().(*types.Slice)
			if !ok {
				return true
			}
			start, end := fset.Position(expression.Pos()).Offset, fset.Position(expression.End()).Offset
			site := CompositionConstructionSite{RootActivityID: ids[0], ActivityID: ids[1], NativeOrdinal: current, ElementType: sequence.Elem().String(), Expression: prior.Source[start:end], ProjectionSHA256: sha, Start: start, End: end, Elements: []CompositionConstructionElementSite{}, Kind: kind, MapperActivityID: mapperID}
			origin := origins[conditionOriginKey{ids[0], ids[1], current}]
			if origin.Kind != kind || kind == "map_values" && origin.MapperActivityID != mapperID {
				bindingError = fmt.Errorf("construction source kind or mapper identity differs from native projection")
				return false
			}
			site.Source = origin.Literal
			if site.Source == nil {
				site.Source = bodycodegen.UnmappedReadSource(prior.SelectedSourceSHA256, "NO_RETAINED_ORIGIN")
			}
			for i, e := range elements {
				a, b := fset.Position(e.Pos()).Offset, fset.Position(e.End()).Offset
				part := CompositionConstructionElementSite{Expression: prior.Source[a:b], Start: a, End: b, Source: bodycodegen.UnmappedReadSource(prior.SelectedSourceSHA256, "NO_RETAINED_ORIGIN")}
				if len(origin.Elements) == len(elements) {
					part.Source = origin.Elements[i]
				}
				site.Elements = append(site.Elements, part)
			}
			site.ExpressionID = compositionDigest(site)
			sites = append(sites, site)
			return true
		})
	}
	if bindingError != nil {
		return nil, bindingError
	}
	if len(sites) > 4096 {
		return nil, fmt.Errorf("collection constructions exceed 4096 sites")
	}
	for _, s := range sites {
		if (s.Kind == "map_values" && (s.MapperActivityID == "" || len(s.Elements) != 1 || s.Source.Status != "MAPPED")) || s.ActivityID == "" || s.RootActivityID == "" || len(s.Elements) > 256 || (s.ElementType != "int64" && s.ElementType != "string" && s.ElementType != "bool") {
			return nil, fmt.Errorf("collection construction lacks supported identity, type or element bound")
		}
	}
	return sites, nil
}
