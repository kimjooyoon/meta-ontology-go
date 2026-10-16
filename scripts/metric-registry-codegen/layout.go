package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
	"strings"
)

// Keep source records readable while bounding generated function declarations.
// Only expression layout changes; surrounding origin markers remain in place.
func compactRecordLiterals(raw []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "registry.generated.go", raw, 0)
	if err != nil {
		return nil, err
	}
	type replacement struct {
		start, end int
		text       string
	}
	var edits []replacement
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		name, ok := literal.Type.(*ast.Ident)
		if !ok || !strings.HasPrefix(name.Name, "GoooRecord") {
			return true
		}
		var output bytes.Buffer
		if printErr := printer.Fprint(&output, token.NewFileSet(), literal); printErr != nil {
			err = printErr
			return false
		}
		edits = append(edits, replacement{fset.Position(literal.Pos()).Offset,
			fset.Position(literal.End()).Offset, output.String()})
		return false
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	output := string(raw)
	for _, edit := range edits {
		output = output[:edit.start] + edit.text + output[edit.end:]
	}
	return []byte(output), nil
}
