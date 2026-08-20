package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func renderPart(fset *token.FileSet, file *ast.File, declarations []ast.Decl) ([]byte, error) {
	importDecls, imports := importsFor(file, declarations)
	allDecls := make([]ast.Decl, 0, len(importDecls)+len(declarations))
	allDecls = append(allDecls, importDecls...)
	allDecls = append(allDecls, declarations...)
	part := *file
	part.Decls = allDecls
	part.Imports = imports
	part.Comments = ast.NewCommentMap(fset, file, file.Comments).Filter(&part).Comments()
	var output bytes.Buffer
	if err := format.Node(&output, fset, &part); err != nil {
		return nil, err
	}
	if output.Len() != 0 && output.Bytes()[output.Len()-1] != '\n' {
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func parseSource(target, subject string) (os.FileInfo, *token.FileSet, *ast.File, error) {
	source, err := os.ReadFile(target)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read %s: %w", subject, err)
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("stat %s: %w", subject, err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, subject, source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse %s: %w", subject, err)
	}
	return info, fset, file, nil
}

func sourceDeclarations(file *ast.File) []ast.Decl {
	declarations := make([]ast.Decl, 0, len(file.Decls))
	for _, declaration := range file.Decls {
		general, isImport := declaration.(*ast.GenDecl)
		if !isImport || general.Tok != token.IMPORT {
			declarations = append(declarations, declaration)
		}
	}
	return declarations
}

func positionLine(fset *token.FileSet, declaration ast.Decl) int {
	return fset.Position(declaration.Pos()).Line
}
