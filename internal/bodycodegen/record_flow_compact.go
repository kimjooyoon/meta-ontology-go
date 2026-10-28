package bodycodegen

import (
	"go/ast"
	"go/types"
)

// Select compaction from used typed expressions, never from an unused layout.
// Legacy flat graphs retain their original node identity and feature bytes.
func recordFlowUsesNominalValues(file *ast.File, info *types.Info, records []RecordType) bool {
	recordIDs := make(map[string]bool, len(records))
	for _, record := range records {
		recordIDs[record.ID] = true
	}
	nominal := map[string]bool{}
	for _, record := range records {
		for _, field := range record.Fields {
			if recordIDs[field.TypeID] {
				nominal[record.Name] = true
				nominal[record.GoName] = true
				break
			}
		}
	}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ast.Inspect(fn, func(node ast.Node) bool {
			if found {
				return false
			}
			if expr, ok := node.(ast.Expr); ok {
				if named, ok := info.Types[expr].Type.(*types.Named); ok && nominal[named.Obj().Name()] {
					found = true
				}
				if literal, ok := expr.(*ast.CompositeLit); ok {
					if sequence, ok := info.Types[literal].Type.(*types.Slice); ok {
						if _, named := sequence.Elem().(*types.Named); named {
							found = true
						}
					}
				}
			}
			return !found
		})
	}
	return found
}

// Record history is separate from a field's data and control dependencies.
func (c *recordFlowBuilder) linkRecordVersion(root, version uint16) {
	if root != 0 && version != 0 {
		c.nodes[root-1].RecordVersion = version
	}
}
