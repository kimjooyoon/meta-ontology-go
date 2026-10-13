package bodycodegen

import (
	"go/ast"
)

// BodyConditionOrigin is captured before lowering and follows the retained
// condition expression node through supported structural route transformations.
// Offsets locate the decoded, let-normalized body, not a generated Go file.
type BodyConditionOrigin struct {
	ActivityID    string                 `json:"activity_id"`
	NativeOrdinal int                    `json:"native_ordinal"`
	SourceOrdinal int                    `json:"source_ordinal"`
	BodySHA256    string                 `json:"body_sha256"`
	Start         int                    `json:"decoded_start"`
	End           int                    `json:"decoded_end"`
	Expression    string                 `json:"expression"`
	Relation      string                 `json:"relation"`
	Fields        []ConditionSourceField `json:"fields,omitempty"`
}

type ConditionSourceField struct {
	Declaration *ConditionDeclarationSource `json:"declaration,omitempty"`
	EntityID    string                      `json:"entity_id"`
	FieldID     string                      `json:"field_id"`
	Name        string                      `json:"name"`
}

func (p bodyProjection) captureConditionOrigins(activityID, body string) map[ast.Expr]BodyConditionOrigin {
	bodies := map[string]string{p.function.Name.Name: body}
	ids := map[string]string{p.function.Name.Name: activityID}
	for _, call := range p.calls {
		bodies[call.identity.Name] = call.body
		ids[call.identity.Name] = call.identity.ActivityID
	}
	origins := make(map[ast.Expr]BodyConditionOrigin)
	for _, declaration := range p.file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		text, id := bodies[f.Name.Name], ids[f.Name.Name]
		base := p.fset.Position(f.Body.Lbrace).Offset + 2
		ordinal := 0
		ast.Inspect(f.Body, func(node ast.Node) bool {
			stmt, ok := node.(*ast.IfStmt)
			if !ok {
				return true
			}
			a, b := p.fset.Position(stmt.Cond.Pos()).Offset-base, p.fset.Position(stmt.Cond.End()).Offset-base
			if a >= 0 && b >= a && b <= len(text) {
				origins[stmt.Cond] = BodyConditionOrigin{ActivityID: id, SourceOrdinal: ordinal, BodySHA256: digest([]byte(text)), Start: a, End: b, Expression: text[a:b], Relation: "retained_condition_node"}
			}
			ordinal++
			return true
		})
	}
	return origins
}

func (p bodyProjection) emittedConditionOrigins() []BodyConditionOrigin {
	var result []BodyConditionOrigin
	for _, declaration := range p.file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ordinal := 0
		ast.Inspect(f.Body, func(node ast.Node) bool {
			stmt, ok := node.(*ast.IfStmt)
			if !ok {
				return true
			}
			if origin, found := p.conditions[stmt.Cond]; found {
				origin.NativeOrdinal = ordinal
				result = append(result, origin)
			}
			ordinal++
			return true
		})
	}
	return result
}
