package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
)

func isHistorySelector(source *goSource, statement ast.Stmt) bool {
	selection, ok := statement.(*ast.IfStmt)
	if !ok || selection.Else != nil || source.text(selection.Cond) != "known" || len(selection.Body.List) != 1 {
		return false
	}
	initial, ok := selection.Init.(*ast.AssignStmt)
	if !ok || initial.Tok != token.DEFINE || len(initial.Lhs) != 2 || len(initial.Rhs) != 1 ||
		source.text(initial.Lhs[0]) != "target" || source.text(initial.Lhs[1]) != "known" ||
		!isHistoryCall(source, initial.Rhs[0], historyChunkPrefix+"0") {
		return false
	}
	assignment, ok := selection.Body.List[0].(*ast.AssignStmt)
	return ok && assignment.Tok == token.ASSIGN && len(assignment.Lhs) == 1 && len(assignment.Rhs) == 1 &&
		source.text(assignment.Lhs[0]) == "expected[0].Target" && source.text(assignment.Rhs[0]) == "target"
}

func historyHelperSignature(source *goSource, helper *ast.FuncDecl) bool {
	if helper.Type.TypeParams != nil || helper.Type.Params.NumFields() != 1 ||
		helper.Type.Results.NumFields() != 2 || len(helper.Type.Results.List) != 2 {
		return false
	}
	parameter := helper.Type.Params.List[0]
	return len(parameter.Names) == 1 && parameter.Names[0].Name == "version" &&
		source.text(parameter.Type) == "int" && source.text(helper.Type.Results.List[0].Type) == "int" &&
		source.text(helper.Type.Results.List[1].Type) == "bool"
}

func historyContinuation(source *goSource, statement ast.Stmt, index int) (bool, error) {
	result, ok := statement.(*ast.ReturnStmt)
	if ok && len(result.Results) == 2 {
		zero, literal := integer(result.Results[0])
		if literal && zero == 0 && source.text(result.Results[1]) == "false" {
			return true, nil
		}
	}
	if ok && len(result.Results) == 1 &&
		isHistoryCall(source, result.Results[0], fmt.Sprintf("%s%d", historyChunkPrefix, index+1)) {
		return false, nil
	}
	return false, fmt.Errorf("historical syntax helper continuation differs")
}

func isHistoryCall(source *goSource, expression ast.Expr, name string) bool {
	call, ok := expression.(*ast.CallExpr)
	return ok && call.Ellipsis == token.NoPos && len(call.Args) == 1 &&
		source.text(call.Fun) == name && source.text(call.Args[0]) == "version"
}
