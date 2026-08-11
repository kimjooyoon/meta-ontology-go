package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

func isGoIdentifier(name string) bool {
	if name == "" {
		return false
	}
	file, err := parser.ParseFile(token.NewFileSet(), "", "package p\nvar "+name+" = 0\n", 0)
	if err != nil || file == nil || len(file.Decls) != 1 {
		return false
	}
	declaration, ok := file.Decls[0].(*ast.GenDecl)
	if !ok || len(declaration.Specs) != 1 {
		return false
	}
	value, ok := declaration.Specs[0].(*ast.ValueSpec)
	return ok && len(value.Names) == 1 && value.Names[0].Name == name
}

func lowerCamel(name string) string {
	if name == "" {
		return name
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func quotePath(path string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(path, `\`, `\\`), `"`, `\"`) + `"`
}
