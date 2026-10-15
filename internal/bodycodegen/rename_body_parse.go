package bodycodegen

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Both reference lookup and renaming use identical parameter scopes and offsets.
func parseRenameBody(body string, arity int, declared ...string) (*ast.File, *token.FileSet, int, error) {
	if arity < 0 || arity > 16 {
		return nil, nil, 0, fmt.Errorf("rename supports bodies with 0..16 inputs")
	}
	names, err := inputbinding.Resolve(arity, declared)
	if err != nil {
		return nil, nil, 0, err
	}
	parameters := make([]string, arity)
	for i, name := range names {
		parameters[i] = name + " any"
	}
	header := "package p\nvar _ = func(" + strings.Join(parameters, ",") + "){\n"
	lowered, err := rewriteLetDeclarations(body)
	if err != nil {
		return nil, nil, 0, err
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "rename.go", header+lowered+"\n}", 0)
	return file, set, len(header), err
}
