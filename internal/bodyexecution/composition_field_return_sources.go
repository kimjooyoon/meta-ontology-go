package bodyexecution

import (
	"fmt"
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type compositionReturnSource struct {
	origin bodycodegen.BodyReturnOrigin
	source *bodycodegen.BodyExpressionSource
}

func compositionSourceReturns(prior Composition) (map[conditionOriginKey]compositionReturnSource, error) {
	result := map[conditionOriginKey]compositionReturnSource{}
	for _, step := range prior.Steps {
		origins := step.Generation.Report.BodyReturns
		if len(origins) == 0 {
			continue
		}
		bound, err := bodycodegen.BindBodyReturnSources([]byte(prior.GoooSource), origins)
		if err != nil {
			return nil, err
		}
		for i, origin := range origins {
			key := conditionOriginKey{step.Generation.Report.ActivityID, origin.ActivityID, origin.NativeOrdinal}
			if _, exists := result[key]; exists {
				return nil, fmt.Errorf("record return origin has duplicate identity")
			}
			result[key] = compositionReturnSource{origin, bound[i]}
		}
	}
	return result, nil
}

func fieldFunctionReturns(function *ast.FuncDecl) []*ast.ReturnStmt {
	var result []*ast.ReturnStmt
	bodycodegen.WalkReturnSyntax(function.Body, func(statement *ast.ReturnStmt) { result = append(result, statement) })
	return result
}

func unparenthesizedFieldValue(value ast.Expr) ast.Expr {
	for {
		wrapped, ok := value.(*ast.ParenExpr)
		if !ok {
			return value
		}
		value = wrapped.X
	}
}
