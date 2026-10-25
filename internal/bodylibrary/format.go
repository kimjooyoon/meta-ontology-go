package bodylibrary

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
)

func formatLibraryProjection(source string) (string, error) {
	if len(source) > 256<<10 {
		return "", fmt.Errorf("library projection exceeds 256 KiB")
	}
	raw, err := format.Source([]byte(source))
	if err != nil {
		return "", err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "library.go", raw, 0)
	if err != nil {
		return "", err
	}
	if _, err := new(types.Config).Check("gooo.library", fset, []*ast.File{file}, nil); err != nil {
		return "", fmt.Errorf("library projection typecheck: %w", err)
	}
	return string(raw), nil
}
