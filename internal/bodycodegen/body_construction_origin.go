package bodycodegen

import (
	"go/ast"
	"reflect"
)

// BodyElementOrigin names an expression in the decoded, normalized Gooo body.
type BodyElementOrigin struct {
	Start      int    `json:"decoded_start"`
	End        int    `json:"decoded_end"`
	Expression string `json:"expression"`
}

type BodyConstructionOrigin struct {
	ActivityID       string              `json:"activity_id"`
	NativeOrdinal    int                 `json:"native_ordinal"`
	BodySHA256       string              `json:"body_sha256"`
	Start            int                 `json:"decoded_start"`
	End              int                 `json:"decoded_end"`
	Expression       string              `json:"expression"`
	Elements         []BodyElementOrigin `json:"elements"`
	Kind             string              `json:"kind,omitempty"`
	MapperActivityID string              `json:"mapper_activity_id,omitempty"`
}

func sequenceConstruction(node ast.Node) *ast.CompositeLit {
	literal, ok := node.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	array, ok := literal.Type.(*ast.ArrayType)
	if !ok || array.Len != nil {
		return nil
	}
	return literal
}

func (p bodyProjection) captureConstructionOrigins(activityID, body string) map[ast.Node]BodyConstructionOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	ids := map[string]string{p.function.Name.Name: activityID}
	for _, call := range p.calls {
		bodies[call.identity.Name], ids[call.identity.Name] = call.body, call.identity.ActivityID
	}
	origins := map[ast.Node]BodyConstructionOrigin{}
	_, mapShadowed := findFunction(p.file, "map_values")
	for _, declaration := range p.file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		text, id := bodies[f.Name.Name], ids[f.Name.Name]
		base := p.fset.Position(f.Body.Lbrace).Offset + 2
		span := func(n ast.Node) (BodyElementOrigin, bool) {
			a, b := p.fset.Position(n.Pos()).Offset-base, p.fset.Position(n.End()).Offset-base
			if a < 0 || b <= a || b > len(text) {
				return BodyElementOrigin{}, false
			}
			return BodyElementOrigin{a, b, text[a:b]}, true
		}
		ast.Inspect(f.Body, func(n ast.Node) bool {
			literal := sequenceConstruction(n)
			var elements []ast.Expr
			kind, mapperID := "", ""
			if literal != nil {
				elements = literal.Elts
			} else if call, ok := n.(*ast.CallExpr); ok && !mapShadowed && mapName(call.Fun, "map_values") && len(call.Args) == 2 {
				mapper, valid := call.Args[1].(*ast.Ident)
				if !valid {
					return true
				}
				kind, mapperID, elements = "map_values", ids[mapper.Name], call.Args[1:]
			} else {
				return true
			}
			location, ok := span(n)
			if !ok {
				return true
			}
			origin := BodyConstructionOrigin{ActivityID: id, BodySHA256: digest([]byte(text)), Start: location.Start, End: location.End, Expression: location.Expression, Elements: []BodyElementOrigin{}, Kind: kind, MapperActivityID: mapperID}
			for _, element := range elements {
				e, valid := span(element)
				if !valid {
					return true
				}
				origin.Elements = append(origin.Elements, e)
			}
			origins[n] = origin
			return true
		})
	}
	return origins
}

func (p bodyProjection) emittedConstructionOrigins() []BodyConstructionOrigin {
	var result []BodyConstructionOrigin
	for _, declaration := range p.file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ordinal := 0
		InspectBodyProjection(f.Body, func(n ast.Node) bool {
			literal := sequenceConstruction(n)
			mapped := false
			if call, ok := n.(*ast.CallExpr); ok {
				_, _, mapped = CollectionMapProjection(call)
			}
			if literal == nil && !mapped {
				return true
			}
			if origin, found := p.constructions[n]; found {
				origin.NativeOrdinal = ordinal
				result = append(result, origin)
			}
			ordinal++
			return true
		})
	}
	return result
}

func EqualRetainedConstructionOrigins(prior, expected []BodyConstructionOrigin) bool {
	return prior == nil || reflect.DeepEqual(prior, expected)
}
