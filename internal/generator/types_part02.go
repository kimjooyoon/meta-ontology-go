package generator

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	gotypes "go/types"
)

func validateGeneratedSource(source []byte, packageName string) error {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "generated.gooo.go", source, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("generator: generated source is not valid Go: %w", err)
	}
	config := gotypes.Config{Importer: importer.Default()}
	if _, err := config.Check(packageName, fileSet, []*ast.File{file}, nil); err != nil {
		return fmt.Errorf("generator: generated source does not compile: %w", err)
	}
	return nil
}
