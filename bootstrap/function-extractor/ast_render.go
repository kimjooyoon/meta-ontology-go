package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"sort"
)

func renderGeneric(fset *token.FileSet, file *ast.File, source []byte, candidates, selected []astDecl, imports []astImport) (astRendered, error) {
	moved := map[ast.Decl]bool{}
	for _, item := range selected {
		moved[item.decl] = true
	}
	var remaining, helpers []ast.Decl
	for _, decl := range file.Decls {
		if moved[decl] {
			helpers = append(helpers, decl)
		} else {
			remaining = append(remaining, decl)
		}
	}
	replacements, _, err := renderGenericImports(fset, file, remaining, imports, true)
	if err != nil {
		return astRendered{}, err
	}
	_, helperImports, err := renderGenericImports(fset, file, helpers, imports, false)
	if err != nil {
		return astRendered{}, err
	}
	var edits []astEdit
	for _, item := range selected {
		edits = append(edits, astEdit{item.start, item.end, nil})
	}
	for _, decl := range file.Decls {
		group, ok := decl.(*ast.GenDecl)
		if ok && group.Tok == token.IMPORT {
			replacement, exists := replacements[group]
			if !exists {
				replacement = nil
			}
			edits = append(edits, astEdit{fset.Position(group.Pos()).Offset, fset.Position(group.End()).Offset, replacement})
		}
	}
	result, err := applyGenericEdits(source, edits)
	if err != nil {
		return astRendered{}, err
	}
	formattedSource, err := format.Source(result)
	if err != nil {
		return astRendered{}, extractionError("rewrite-source", "format-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", []string{})
	}
	sort.SliceStable(selected, func(i, j int) bool { return selected[i].start < selected[j].start })
	var helper bytes.Buffer
	packageOffset := fset.Position(file.Package).Offset
	helper.Write(source[:packageOffset])
	helper.WriteString("package ")
	helper.WriteString(file.Name.Name)
	helper.WriteString("\n\n")
	helper.Write(helperImports)
	for _, item := range selected {
		helper.Write(source[item.start:item.end])
		if helper.Len() > 0 && helper.Bytes()[helper.Len()-1] != '\n' {
			helper.WriteByte('\n')
		}
	}
	formattedHelper, err := format.Source(helper.Bytes())
	if err != nil {
		return astRendered{}, extractionError("generate-helpers", "format-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", []string{})
	}
	return astRendered{formattedSource, formattedHelper}, nil
}
