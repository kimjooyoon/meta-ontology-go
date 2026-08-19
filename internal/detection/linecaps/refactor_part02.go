package linecaps

import (
	"fmt"
	"go/ast"
	"strings"
)

func expressionKind(expression ast.Expr) string {
	kind := strings.TrimPrefix(fmt.Sprintf("%T", expression), "*ast.")
	if kind == "" {
		return "expression"
	}
	return kind
}
