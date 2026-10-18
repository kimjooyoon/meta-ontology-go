package bodycodegen

import (
	"go/ast"
	"reflect"
)

// A field origin retains authored syntax independently of native field spelling.
// Ordinals and populations include synthetic/unretained syntax of the same kind.
type BodyFieldOrigin struct {
	ActivityID       string            `json:"activity_id"`
	Kind             string            `json:"kind"`
	Operation        string            `json:"operation,omitempty"`
	NativeOrdinal    int               `json:"native_ordinal"`
	NativePopulation int               `json:"native_population"`
	RecordID         string            `json:"record_id"`
	FieldID          string            `json:"field_id"`
	BodySHA256       string            `json:"body_sha256"`
	Value            BodyElementOrigin `json:"value"`
	Target           BodyElementOrigin `json:"target"`
}

// FieldSyntax is a syntactic slot, not a claim of record/field identity.
// Only a successful typed capture may associate it with a Gooo stable ID.
type FieldSyntax struct {
	Node      ast.Node
	Field     *ast.Ident
	Value     ast.Node
	Target    ast.Node
	Kind      string
	Operation string
}

// WalkFieldSyntax is shared by source retention and native observation. Compiler
// unused-value markers do not acquire field-read ordinals.
func WalkFieldSyntax(body *ast.BlockStmt, visit func(FieldSyntax)) {
	writes := map[*ast.SelectorExpr]bool{}
	InspectBodyProjection(body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.AssignStmt:
			if len(v.Lhs) == 1 && len(v.Rhs) == 1 {
				if id, ok := v.Lhs[0].(*ast.Ident); ok && id.Name == "_" {
					return false
				}
				if field, ok := v.Lhs[0].(*ast.SelectorExpr); ok {
					_, compound := CompoundAssignmentOperator(v.Tok)
					writes[field] = !compound
					slot := FieldSyntax{Node: n, Field: field.Sel, Value: v.Rhs[0], Target: field, Kind: "field_assignment"}
					if compound {
						slot.Operation = v.Tok.String()
					}
					visit(slot)
				}
			}
		case *ast.KeyValueExpr:
			if key, ok := v.Key.(*ast.Ident); ok {
				visit(FieldSyntax{Node: n, Field: key, Value: v.Value, Target: key, Kind: "record_constructor"})
			}
		case *ast.SelectorExpr:
			if !writes[v] {
				visit(FieldSyntax{Node: n, Field: v.Sel, Value: v, Target: v, Kind: "field_read"})
			}
		}
		return true
	})
}

func EqualRetainedFieldOrigins(prior, expected []BodyFieldOrigin) bool {
	return prior == nil || reflect.DeepEqual(prior, expected)
}
