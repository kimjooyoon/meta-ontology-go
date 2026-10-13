package bodycodegen

import (
	"go/ast"
	"reflect"
)

// Local origins name lexical variables and their write expressions. Ordinals
// are scoped to one activity body; repeated runtime writes are separate work.
type BodyLocalOrigin struct {
	ActivityID      string `json:"activity_id"`
	NativeOrdinal   int    `json:"native_ordinal"`
	VariableOrdinal int    `json:"variable_ordinal"`
	Name            string `json:"name"`
	Kind            string `json:"kind"`
	Operation       string `json:"operation"`
	BodySHA256      string `json:"body_sha256"`
	BodyElementOrigin
	Declaration BodyElementOrigin `json:"declaration"`
}

type bodyLocalWrite struct {
	name            *ast.Ident
	value           ast.Node
	kind, operation string
}

// LocalWriteExpression defines the projection traversal shared by source capture
// and native instrumentation. Compiler-created writes consume ordinals too.
func LocalWriteExpression(node ast.Node) (*ast.Ident, ast.Node, string, string) {
	w := localWrite(node)
	if w.name == nil || w.name.Name == "_" {
		return nil, nil, "", ""
	}
	return w.name, w.value, w.kind, w.operation
}

func localWrite(node ast.Node) bodyLocalWrite {
	switch value := node.(type) {
	case *ast.ValueSpec:
		if len(value.Names) == 1 && len(value.Values) == 1 {
			return bodyLocalWrite{value.Names[0], value.Values[0], "declaration", "var"}
		}
	case *ast.AssignStmt:
		if len(value.Lhs) == 1 && len(value.Rhs) == 1 {
			if name, ok := value.Lhs[0].(*ast.Ident); ok {
				return bodyLocalWrite{name, value.Rhs[0], "assignment", value.Tok.String()}
			}
		}
	case *ast.IncDecStmt:
		if name, ok := value.X.(*ast.Ident); ok {
			return bodyLocalWrite{name, value, "update", value.Tok.String()}
		}
	case *ast.RangeStmt:
		if name, ok := value.Value.(*ast.Ident); ok {
			return bodyLocalWrite{name, value.X, "iteration", value.Tok.String()}
		}
	}
	return bodyLocalWrite{}
}

func retainedLocalWrite(node ast.Node) bodyLocalWrite {
	w := localWrite(node)
	if w.name == nil || w.name.Name == "_" || w.name.Obj == nil {
		return bodyLocalWrite{}
	}
	return w
}

func EqualRetainedLocalOrigins(prior, expected []BodyLocalOrigin) bool {
	return prior == nil || reflect.DeepEqual(prior, expected)
}
