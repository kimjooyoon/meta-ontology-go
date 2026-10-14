package bodycodegen

import (
	"go/ast"
	"reflect"
)

// Native ordinals include generated returns. An absent authored value remains
// explicit rather than attributing a merged result to one source branch.
type BodyReturnOrigin struct {
	ActivityID       string             `json:"activity_id"`
	RecordID         string             `json:"record_id"`
	NativeOrdinal    int                `json:"native_ordinal"`
	NativePopulation int                `json:"native_population"`
	BodySHA256       string             `json:"body_sha256,omitempty"`
	Value            *BodyElementOrigin `json:"value,omitempty"`
	Reason           string             `json:"reason,omitempty"`
}

func WalkReturnSyntax(body *ast.BlockStmt, visit func(*ast.ReturnStmt)) {
	ast.Inspect(body, func(node ast.Node) bool {
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}
		if statement, ok := node.(*ast.ReturnStmt); ok && len(statement.Results) == 1 {
			visit(statement)
		}
		return true
	})
}

func EqualRetainedReturnOrigins(prior, expected []BodyReturnOrigin) bool {
	return prior == nil || reflect.DeepEqual(prior, expected)
}
