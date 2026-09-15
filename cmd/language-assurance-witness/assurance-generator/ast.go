package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
)

func parseSource(path string, source []byte) (*token.FileSet, *ast.File, error) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, path, source, parser.ParseComments)
	return files, file, err
}

func formatSource(files *token.FileSet, file *ast.File) ([]byte, error) {
	var output bytes.Buffer
	err := format.Node(&output, files, file)
	return output.Bytes(), err
}

func hasString(node ast.Node, wanted string) bool {
	found := false
	ast.Inspect(node, func(current ast.Node) bool {
		literal, ok := current.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err == nil && value == wanted {
			found = true
		}
		return !found
	})
	return found
}

func hasIdentifier(node ast.Node, wanted string) bool {
	found := false
	ast.Inspect(node, func(current ast.Node) bool {
		identifier, ok := current.(*ast.Ident)
		if ok && identifier.Name == wanted {
			found = true
		}
		return !found
	})
	return found
}
