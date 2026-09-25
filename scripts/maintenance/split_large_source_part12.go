package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

func filterTopLevelDecls(decls []ast.Decl) []ast.Decl {
	result := make([]ast.Decl, 0, len(decls))
	for _, decl := range decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
			continue
		}
		result = append(result, decl)
	}
	return result
}

func validateSplitSafety(file *ast.File, decls []ast.Decl, imports []importSpec) error {
	for _, imp := range imports {
		if imp.path == "C" {
			return fmt.Errorf("cgo preamble requires a dedicated transformer")
		}
	}
	for _, decl := range decls {
		switch node := decl.(type) {
		case *ast.FuncDecl:
			if node.Name.Name == "init" {
				return fmt.Errorf("init order requires a coherence proof")
			}
		case *ast.GenDecl:
			if node.Tok == token.VAR {
				return fmt.Errorf("package variable order requires a coherence proof")
			}
		}
	}
	used := make(map[string]struct{})
	for _, decl := range decls {
		used = mergeImportSet(used, usedImportsForDecl(file, decl))
	}
	for _, imp := range imports {
		if _, ok := used[imp.key()]; !ok {
			return fmt.Errorf("cannot safely attribute import %q", imp.path)
		}
	}
	return nil
}
