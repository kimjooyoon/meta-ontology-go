package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"sort"
	"strings"
)

// renameBodyCalls edits resolved free call names, retaining every other byte.
// Parser objects distinguish parameters and local declarations from activities.
func renameBodyCalls(body string, arity int, from, to string) (string, int, error) {
	if arity < 0 || arity > 16 {
		return "", 0, fmt.Errorf("rename supports bodies with 0..16 inputs")
	}
	parameters := make([]string, arity)
	for i := range parameters {
		name := "input"
		if arity > 1 {
			name = fmt.Sprintf("input%d", i)
		}
		parameters[i] = name + " any"
	}
	header := "package p\nvar _ = func(" + strings.Join(parameters, ",") + "){\n"
	parse := func(value string) (*ast.File, *token.FileSet, error) {
		lowered, err := rewriteLetDeclarations(value)
		if err != nil {
			return nil, nil, err
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, "rename.go", header+lowered+"\n}", 0)
		return file, set, err
	}
	file, set, err := parse(body)
	if err != nil {
		return "", 0, err
	}
	calls, ignored := map[*ast.Ident]bool{}, map[*ast.Ident]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			if id, ok := ast.Unparen(n.Fun).(*ast.Ident); ok {
				calls[id] = true
			}
		case *ast.SelectorExpr:
			ignored[n.Sel] = true
		case *ast.KeyValueExpr:
			if id, ok := n.Key.(*ast.Ident); ok {
				ignored[id] = true
			}
		}
		return true
	})
	var offsets []int
	ast.Inspect(file, func(node ast.Node) bool {
		id, ok := node.(*ast.Ident)
		if !ok || id.Obj != nil || ignored[id] || set.Position(id.Pos()).Offset < len(header) {
			return true
		}
		if id.Name == to {
			err = fmt.Errorf("new activity name %q would bind an existing free name", to)
		}
		if id.Name == from {
			if !calls[id] {
				err = fmt.Errorf("activity %q is used outside a direct call", from)
			} else {
				offsets = append(offsets, set.Position(id.Pos()).Offset-len(header))
			}
		}
		return true
	})
	if err != nil {
		return "", 0, err
	}
	sort.Ints(offsets)
	updated := body
	for _, offset := range slices.Backward(offsets) {
		updated = updated[:offset] + to + updated[offset+len(from):]
	}
	after, _, err := parse(updated)
	if err != nil {
		return "", 0, err
	}
	freeRenamed := 0
	ast.Inspect(after, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok && id.Name == to && id.Obj == nil {
			freeRenamed++
		}
		return true
	})
	if freeRenamed != len(offsets) {
		return "", 0, fmt.Errorf("new activity name %q is captured by a parameter or local declaration", to)
	}
	return updated, len(offsets), nil
}
