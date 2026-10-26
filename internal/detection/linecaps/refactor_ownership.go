package linecaps

import (
	"go/ast"
	"strings"
)

// Whole-file Go formatting cannot preserve a different source owner's exact
// projection. Keep measuring it, but repair the owner before regenerating Go.
func AllowsWholeFileGoRefactor(file *ast.File) bool {
	if ast.IsGenerated(file) {
		return false
	}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			for _, prefix := range []string{"//gooo:generated:", "//gooo:generated ", "//gooo:slot:"} {
				if strings.HasPrefix(comment.Text, prefix) {
					return false
				}
			}
		}
	}
	return true
}
