package linecaps

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func expressionKind(expression ast.Expr) string {
	kind := strings.TrimPrefix(fmt.Sprintf("%T", expression), "*ast.")
	if kind == "" {
		return "expression"
	}
	return kind
}

func FunctionIdentity(fset *token.FileSet, node ast.Node) (string, int, bool) {
	name, start, _, ok := functionSpan(fset, node)
	return name, start, ok
}

func CollapseAssignReturn(fset *token.FileSet, node ast.Node, comments []*ast.CommentGroup) bool {
	body, _, _, _, ok := functionBodyForRefactor(fset, node)
	if !ok {
		return false
	}
	rule, _, _, _ := classifyRefactorCandidate(body)
	if rule != RuleRefactorAssign {
		return false
	}
	assign := body.List[0].(*ast.AssignStmt)
	result := body.List[1].(*ast.ReturnStmt)
	for _, comment := range comments {
		if comment.Pos() >= assign.Pos() && comment.Pos() <= result.End() {
			return false
		}
	}
	body.List = []ast.Stmt{&ast.ReturnStmt{Return: result.Return, Results: assign.Rhs}}
	return true
}
