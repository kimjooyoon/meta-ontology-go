package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
)

func renderGenericImports(fset *token.FileSet, file *ast.File, decls []ast.Decl, imports []astImport, includeBlank bool) (map[*ast.GenDecl][]byte, []byte, error) {
	selected, err := genericImportSpecs(decls, imports, includeBlank)
	if err != nil {
		return nil, nil, err
	}
	replacements := map[*ast.GenDecl][]byte{}
	var helper bytes.Buffer
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if !ok || group.Tok != token.IMPORT || len(selected[group]) == 0 {
			continue
		}
		data, err := formatGenericImport(fset, group, selected[group])
		if err != nil {
			return nil, nil, err
		}
		replacements[group] = data
		helper.Write(data)
	}
	return replacements, helper.Bytes(), nil
}

func formatGenericImport(fset *token.FileSet, group *ast.GenDecl, specs []*ast.ImportSpec) ([]byte, error) {
	copyGroup := *group
	copyGroup.Specs = make([]ast.Spec, len(specs))
	for i, spec := range specs {
		copyGroup.Specs[i] = spec
	}
	if len(specs) == 1 {
		copyGroup.Lparen, copyGroup.Rparen = token.NoPos, token.NoPos
	}
	var out bytes.Buffer
	if err := format.Node(&out, fset, &copyGroup); err != nil {
		return nil, extractionError("rewrite-source", "render-imports", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", []string{})
	}
	return out.Bytes(), nil
}
