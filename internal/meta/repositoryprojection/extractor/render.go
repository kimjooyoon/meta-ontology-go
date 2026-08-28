package extractor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"sort"
)

func render(fset *token.FileSet, file *ast.File, source []byte, selected []declaration, list []importSpec) (rendered, error) {
	moved := map[ast.Decl]bool{}
	for _, item := range selected {
		moved[item.node] = true
	}
	var remaining, helpers []ast.Decl
	for _, node := range file.Decls {
		if moved[node] {
			helpers = append(helpers, node)
		} else {
			remaining = append(remaining, node)
		}
	}
	replacements, _, err := renderImports(fset, file, remaining, list, true)
	if err != nil {
		return rendered{}, err
	}
	_, helperImports, err := renderImports(fset, file, helpers, list, false)
	if err != nil {
		return rendered{}, err
	}
	var edits []edit
	for _, item := range selected {
		edits = append(edits, edit{item.start, item.end, nil})
	}
	for _, node := range file.Decls {
		if group, ok := node.(*ast.GenDecl); ok && group.Tok == token.IMPORT {
			edits = append(edits, edit{fset.Position(group.Pos()).Offset, fset.Position(group.End()).Offset, replacements[group]})
		}
	}
	result, err := applyEdits(source, edits)
	if err != nil {
		return rendered{}, err
	}
	formattedSource, err := format.Source(result)
	if err != nil {
		return rendered{}, fail("rewrite-source", "format-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
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
		return rendered{}, fail("generate-helpers", "format-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return rendered{formattedSource, formattedHelper}, nil
}
