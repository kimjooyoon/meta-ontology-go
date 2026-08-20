package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
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
