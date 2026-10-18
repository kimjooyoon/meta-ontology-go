package bodycodegen

import (
	"go/ast"
	"reflect"
)

// BodyReadOrigin follows the same syntax node through route and record lowering.
// Its offsets refer to the decoded, let-normalized source body.
type BodyReadOrigin struct {
	ActivityID    string `json:"activity_id"`
	Kind          string `json:"kind"`
	NativeOrdinal int    `json:"native_ordinal"`
	BodySHA256    string `json:"body_sha256"`
	Start         int    `json:"decoded_start"`
	End           int    `json:"decoded_end"`
	Expression    string `json:"expression"`
}

func readExpression(node ast.Node) (ast.Expr, string) {
	switch n := node.(type) {
	case *ast.IndexExpr:
		return n, "index"
	case *ast.RangeStmt:
		return n.X, "range_value"
	case *ast.CallExpr:
		if input, _, mapped := CollectionMapProjection(n); mapped {
			return input, "map_value"
		}
	}
	return nil, ""
}

func (p bodyProjection) captureReadOrigins(activityID, body string) map[ast.Node]BodyReadOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	ids := map[string]string{p.function.Name.Name: activityID}
	for _, call := range p.calls {
		bodies[call.identity.Name], ids[call.identity.Name] = call.body, call.identity.ActivityID
	}
	origins := map[ast.Node]BodyReadOrigin{}
	_, mapShadowed := findFunction(p.file, "map_values")
	for _, declaration := range p.file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		text, id := bodies[f.Name.Name], ids[f.Name.Name]
		base := p.fset.Position(f.Body.Lbrace).Offset + 2
		ast.Inspect(f.Body, func(node ast.Node) bool {
			expression, kind := readExpression(node)
			if call, ok := node.(*ast.CallExpr); ok && !mapShadowed && mapName(call.Fun, "map_values") && len(call.Args) == 2 {
				expression, kind = call.Args[0], "map_value"
			}
			if expression == nil {
				return true
			}
			a, b := p.fset.Position(expression.Pos()).Offset-base, p.fset.Position(expression.End()).Offset-base
			if a >= 0 && b >= a && b <= len(text) {
				origins[node] = BodyReadOrigin{ActivityID: id, Kind: kind, BodySHA256: digest([]byte(text)), Start: a, End: b, Expression: text[a:b]}
			}
			return true
		})
	}
	return origins
}

func (p bodyProjection) emittedReadOrigins() []BodyReadOrigin {
	var result []BodyReadOrigin
	for _, declaration := range p.file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ordinal := 0
		InspectBodyProjection(f.Body, func(node ast.Node) bool {
			if expression, _ := readExpression(node); expression == nil {
				return true
			}
			if origin, ok := p.reads[node]; ok {
				origin.NativeOrdinal = ordinal
				result = append(result, origin)
			}
			ordinal++
			return true
		})
	}
	return result
}

func EqualRetainedReadOrigins(prior, expected []BodyReadOrigin) bool {
	return prior == nil || reflect.DeepEqual(prior, expected)
}
