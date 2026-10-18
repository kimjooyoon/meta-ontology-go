package bodycodegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// CollectionFilterProjection recognizes the whole generated loop, including
// its bounded array, stable order, predicate call and capacity-limited result.
func CollectionFilterProjection(call *ast.CallExpr) (ast.Expr, *ast.Ident, bool) {
	function, ok := call.Fun.(*ast.FuncLit)
	if !ok || len(call.Args) != 2 || call.Ellipsis.IsValid() || function.Type.Params == nil || len(function.Type.Params.List) != 2 {
		return nil, nil, false
	}
	predicate, ok := call.Args[1].(*ast.Ident)
	if !ok {
		return nil, nil, false
	}
	inputType := mapScalarSequence(function.Type.Params.List[0].Type)
	if inputType == "" {
		return nil, nil, false
	}
	// This independent inverse template is compared with the entire function
	// literal. Changing one index, branch or bound cannot retain the primitive.
	expected := fmt.Sprintf(`func(goooFilterInput []%s, goooFilterApply func(%s) bool) []%s {
 var goooFilterResult [%d]%s
 var goooFilterLength int
 for goooFilterIndex := range goooFilterInput {
  if goooFilterApply(goooFilterInput[goooFilterIndex]) {
   goooFilterResult[goooFilterLength] = goooFilterInput[goooFilterIndex]
   goooFilterLength++
  }
 }
 return goooFilterResult[:goooFilterLength:goooFilterLength]
}`, inputType, inputType, inputType, RepeatedFieldLimit, inputType)
	parsed, err := parser.ParseExpr(expected)
	if err != nil {
		return nil, nil, false
	}
	var actualText, expectedText bytes.Buffer
	if format.Node(&actualText, token.NewFileSet(), function) != nil ||
		format.Node(&expectedText, token.NewFileSet(), parsed) != nil || actualText.String() != expectedText.String() {
		return nil, nil, false
	}
	return call.Args[0], predicate, true
}

func restoreCollectionFilters(file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if input, predicate, recognized := CollectionFilterProjection(call); recognized {
				call.Fun = ast.NewIdent("filter_values")
				call.Args = []ast.Expr{input, ast.NewIdent(predicate.Name)}
			}
		}
		return true
	})
}

func hasCollectionFilters(file *ast.File) bool {
	if _, shadowed := findFunction(file, "filter_values"); shadowed {
		return false
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && mapName(call.Fun, "filter_values") {
			found = true
		}
		return !found
	})
	return found
}
