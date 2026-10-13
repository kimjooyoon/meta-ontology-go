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
	ActivityID    string `json:"activity_id"`
	NativeOrdinal int    `json:"native_ordinal"`
	BodySHA256    string `json:"body_sha256"`
	BodyElementOrigin
	Elements []BodyElementOrigin `json:"elements"`
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

func (p bodyProjection) captureConstructionOrigins(activityID, body string) map[*ast.CompositeLit]BodyConstructionOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	ids := map[string]string{p.function.Name.Name: activityID}
	for _, call := range p.calls {
		bodies[call.identity.Name], ids[call.identity.Name] = call.body, call.identity.ActivityID
	}
	origins := map[*ast.CompositeLit]BodyConstructionOrigin{}
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
			if literal == nil {
				return true
			}
			location, ok := span(literal)
			if !ok {
				return true
			}
			origin := BodyConstructionOrigin{ActivityID: id, BodySHA256: digest([]byte(text)), BodyElementOrigin: location, Elements: []BodyElementOrigin{}}
			for _, element := range literal.Elts {
				e, valid := span(element)
				if !valid {
					return true
				}
				origin.Elements = append(origin.Elements, e)
			}
			origins[literal] = origin
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
		ast.Inspect(f.Body, func(n ast.Node) bool {
			literal := sequenceConstruction(n)
			if literal == nil {
				return true
			}
			if origin, found := p.constructions[literal]; found {
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
