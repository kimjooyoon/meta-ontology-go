package policyprojection

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

func checkNativeProjection(projection, target []byte) error {
	set := token.NewFileSet()
	var files []*ast.File
	for index, source := range [][]byte{projection, target} {
		file, err := parser.ParseFile(set, fmt.Sprintf("disposition-%d.go", index), source, 0)
		if err != nil {
			return err
		}
		files = append(files, file)
	}
	config := types.Config{}
	if _, err := config.Check("unresolvedpolicy", set, files, nil); err != nil {
		return fmt.Errorf("policy native bridge typecheck: %w", err)
	}
	return nil
}
