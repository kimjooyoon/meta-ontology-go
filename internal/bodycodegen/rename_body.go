package bodycodegen

import (
	"fmt"
	"go/ast"
	"slices"
	"sort"
)

// renameBodyCalls edits resolved free call names, retaining every other byte.
// Parser objects distinguish parameters and local declarations from activities.
func renameBodyCalls(body string, arity int, from, to string) (string, int, error) {
	file, set, header, err := parseRenameBody(body, arity)
	if err != nil {
		return "", 0, err
	}
	_, mapping, _ := rewriteBodySurface(body)
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
		if !ok || id.Obj != nil || ignored[id] || set.Position(id.Pos()).Offset < header {
			return true
		}
		if id.Name == to {
			err = fmt.Errorf("new activity name %q would bind an existing free name", to)
		}
		if id.Name == from {
			if !calls[id] {
				err = fmt.Errorf("activity %q is used outside a direct call", from)
			} else {
				offsets = append(offsets, mapping.originalOffset(set.Position(id.Pos()).Offset-header))
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
	after, _, _, err := parseRenameBody(updated, arity)
	if err != nil {
		return "", 0, err
	}
	freeRenamed := countFreeBodyCalls(after, to)
	if freeRenamed != len(offsets) {
		return "", 0, fmt.Errorf("new activity name %q is captured by a parameter or local declaration", to)
	}
	return updated, len(offsets), nil
}
