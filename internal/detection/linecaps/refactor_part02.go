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
	if _, ok := body.List[1].(*ast.ReturnStmt); !ok {
		return false
	}
	for _, comment := range comments {
		if comment.Pos() >= assign.Pos() && comment.Pos() <= body.Rbrace {
			return false
		}
	}
	// Re-anchor the generated statement and closing brace to the removed
	// assignment. Keeping the original return and brace positions makes the
	// printer preserve their old line gap after the assignment is removed.
	body.List = []ast.Stmt{&ast.ReturnStmt{Return: assign.Pos(), Results: assign.Rhs}}
	body.Rbrace = assign.End()
	return true
}
