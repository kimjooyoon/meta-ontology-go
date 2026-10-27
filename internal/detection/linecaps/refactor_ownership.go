package linecaps

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourceownership"
)

// Whole-file Go formatting cannot preserve a different source owner's exact
// projection. Keep measuring it, but repair the owner before regenerating Go.
func AllowsWholeFileGoRefactor(file *ast.File) bool {
	return GoSourceRewritePermission(file).F_may_rewrite_go
}

func GoSourceRewritePermission(file *ast.File) sourceownership.R_Permission {
	return sourceownership.Evaluate(hasGeneratorOwner(file))
}

func SourceHasGeneratorOwner(name string, source []byte) bool {
	file, err := parser.ParseFile(token.NewFileSet(), name, source, parser.ParseComments)
	return err == nil && hasGeneratorOwner(file)
}

func hasGeneratorOwner(file *ast.File) bool {
	if ast.IsGenerated(file) {
		return true
	}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			for _, prefix := range []string{"//gooo:generated:", "//gooo:generated ", "//gooo:slot:"} {
				if strings.HasPrefix(comment.Text, prefix) {
					return true
				}
			}
		}
	}
	return false
}
