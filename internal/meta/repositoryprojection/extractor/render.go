package extractor

import (
	"bytes"
	"fmt"
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

func capacityRender(fset *token.FileSet, file *ast.File, source []byte, all []declaration, list []importSpec, limit int) (rendered, [][]declaration, error) {
	output, err := render(fset, file, source, all, list)
	if err != nil {
		return rendered{}, nil, err
	}
	if physicalLines(output.source) > limit {
		return rendered{}, nil, failWithDiagnostics("derive-recipe", "select-declaration", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{fmt.Sprintf("remaining_lines=%d", physicalLines(output.source))})
	}
	partitions := make([][]declaration, 0, len(all))
	current := make([]declaration, 0)
	for _, candidate := range all {
		trial := append(append([]declaration{}, current...), candidate)
		trialRendered, trialErr := render(fset, file, source, trial, list)
		if trialErr != nil {
			return rendered{}, nil, trialErr
		}
		if physicalLines(trialRendered.helper) <= limit {
			current = trial
			continue
		}
		if len(current) == 0 {
			return rendered{}, nil, failWithDiagnostics("derive-recipe", "select-declaration", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{fmt.Sprintf("declaration=%s", candidate.identity), fmt.Sprintf("helper_lines=%d", physicalLines(trialRendered.helper))})
		}
		partitions = append(partitions, current)
		current = []declaration{candidate}
		single, singleErr := render(fset, file, source, current, list)
		if singleErr != nil {
			return rendered{}, nil, singleErr
		}
		if physicalLines(single.helper) > limit {
			return rendered{}, nil, failWithDiagnostics("derive-recipe", "select-declaration", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{fmt.Sprintf("declaration=%s", candidate.identity), fmt.Sprintf("helper_lines=%d", physicalLines(single.helper))})
		}
	}
	if len(current) > 0 {
		partitions = append(partitions, current)
	}
	return output, partitions, nil
}
