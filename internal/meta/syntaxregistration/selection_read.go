package syntaxregistration

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
)

type selectionEntry struct {
	capability int
	result     string
}

func selectionEntries(source *goSource, root *ast.FuncDecl, kind, empty string) ([]selectionEntry, []*ast.FuncDecl, error) {
	helpers := map[int]*ast.FuncDecl{}
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, root.Name.Name+"Chunk") {
			continue
		}
		index, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, root.Name.Name+"Chunk"))
		if err != nil || index < 0 || index >= 256 || helpers[index] != nil || function.Recv != nil {
			return nil, nil, fmt.Errorf("denominator selection chunks are ambiguous")
		}
		helpers[index] = function
	}
	if len(helpers) == 0 {
		entries, err := readSelectionSwitch(source, root, "languagesyntax.FixedCapabilityTotal", empty, true)
		return entries, nil, err
	}
	owned := make([]*ast.FuncDecl, len(helpers))
	var entries []selectionEntry
	for index := range owned {
		if helpers[index] == nil {
			return nil, nil, fmt.Errorf("denominator selection chunk sequence has a gap")
		}
		owned[index] = helpers[index]
		part, err := readSelectionSwitch(source, helpers[index], "total", empty, false)
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, part...)
	}
	wrapper, _ := renderSelection(root.Name.Name, kind, empty, append([]selectionEntry(nil), entries...))
	if strings.TrimSpace(source.text(root)) != strings.TrimSpace(formatSelection(wrapper)) {
		return nil, nil, fmt.Errorf("denominator selection wrapper differs from its owned chunks")
	}
	return entries, owned, nil
}

func readSelectionSwitch(source *goSource, function *ast.FuncDecl, tag, empty string, legacy bool) ([]selectionEntry, error) {
	want := 2
	if legacy {
		want = 1
	}
	if function.Body == nil || len(function.Body.List) != want {
		return nil, fmt.Errorf("denominator selection body has an unsupported shape")
	}
	selection, ok := function.Body.List[0].(*ast.SwitchStmt)
	if !ok || selection.Tag == nil || source.text(selection.Tag) != tag || selection.Init != nil {
		return nil, fmt.Errorf("denominator selection switch is unbound")
	}
	if !legacy && returnSelection(source, function.Body.List[1]) != empty {
		return nil, fmt.Errorf("denominator selection fallback differs")
	}
	var entries []selectionEntry
	seen := map[int]bool{}
	defaults := 0
	for _, statement := range selection.Body.List {
		part, fallback, err := readSelectionClause(source, statement, empty)
		if err != nil {
			return nil, err
		}
		if fallback {
			defaults++
			if !legacy || defaults > 1 {
				return nil, fmt.Errorf("denominator selection default differs")
			}
		}
		for _, entry := range part {
			if seen[entry.capability] {
				return nil, fmt.Errorf("denominator selection case differs or repeats")
			}
			seen[entry.capability] = true
			entries = append(entries, entry)
		}
	}
	if legacy && defaults != 1 {
		return nil, fmt.Errorf("denominator selection lacks its exact fallback")
	}
	return entries, nil
}

func readSelectionClause(source *goSource, statement ast.Stmt, empty string) ([]selectionEntry, bool, error) {
	clause, ok := statement.(*ast.CaseClause)
	if !ok || len(clause.Body) != 1 {
		return nil, false, fmt.Errorf("denominator selection case is not one return")
	}
	value := returnSelection(source, clause.Body[0])
	if len(clause.List) == 0 {
		if value != empty {
			return nil, false, fmt.Errorf("denominator selection fallback differs")
		}
		return nil, true, nil
	}
	if value == "" || value == empty {
		return nil, false, fmt.Errorf("denominator selection case lacks its value")
	}
	returned := clause.Body[0].(*ast.ReturnStmt)
	if _, ok := returned.Results[0].(*ast.Ident); !ok {
		return nil, false, fmt.Errorf("denominator selection value is not a bound symbol")
	}
	var entries []selectionEntry
	for _, expression := range clause.List {
		capability, ok := integer(expression)
		if !ok {
			return nil, false, fmt.Errorf("denominator selection case is not literal")
		}
		entries = append(entries, selectionEntry{capability, value})
	}
	return entries, false, nil
}

func returnSelection(source *goSource, statement ast.Stmt) string {
	value, ok := statement.(*ast.ReturnStmt)
	if !ok || len(value.Results) != 1 {
		return ""
	}
	return source.text(value.Results[0])
}
