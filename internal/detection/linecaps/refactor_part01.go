package linecaps

import (
	"fmt"
	"go/ast"
	"go/token"
)

func refactorCandidateFinding(fset *token.FileSet, node ast.Node, path string) (Finding, bool) {
	body, name, start, end, ok := functionBodyForRefactor(fset, node)
	if !ok {
		return Finding{}, false
	}
	if rule, detail, actual, limit := classifyRefactorCandidate(body); rule != "" {
		return Finding{
			Path:      path,
			Rule:      rule,
			Name:      name,
			StartLine: start,
			EndLine:   end,
			Actual:    actual,
			Limit:     limit,
			Detail:    detail,
		}, true
	}
	return Finding{}, false
}
func functionBodyForRefactor(fset *token.FileSet, node ast.Node) (*ast.BlockStmt, string, int, int, bool) {
	switch function := node.(type) {
	case *ast.FuncDecl:
		name := function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
		return function.Body, name, fset.Position(function.Pos()).Line, fset.Position(function.End()).Line, true
	case *ast.FuncLit:
		return function.Body, "function literal", fset.Position(function.Pos()).Line, fset.Position(function.End()).Line, true
	default:
		return nil, "", 0, 0, false
	}
}
func classifyRefactorCandidate(body *ast.BlockStmt) (rule Rule, detail string, actual int, limit int) {
	if body == nil || len(body.List) == 0 {
		return "", "", 0, 0
	}
	if len(body.List) == 1 {
		ret, ok := body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 || ret.Results[0] == nil {
			return "", "", 0, 0
		}
		return RuleRefactorReturn, fmt.Sprintf("single return %s", expressionKind(ret.Results[0])), 1, 1
	}
	if len(body.List) != 2 {
		return "", "", 0, 0
	}
	assign, ok := body.List[0].(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return "", "", 0, 0
	}
	lhs, ok := assign.Lhs[0].(*ast.Ident)
	if !ok || lhs.Name == "_" {
		return "", "", 0, 0
	}
	ret, ok := body.List[1].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return "", "", 0, 0
	}
	result, ok := ret.Results[0].(*ast.Ident)
	if !ok || result.Name != lhs.Name {
		return "", "", 0, 0
	}
	return RuleRefactorAssign, "assignment then return " + lhs.Name, 2, 2
}
