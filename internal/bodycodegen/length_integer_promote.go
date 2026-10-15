package bodycodegen

import (
	"go/ast"
	"go/types"
)

// Promotion occurs only after an actual length/Integer mismatch was repaired.
// Already accepted programs keep their projection; newly accepted inferred
// length locals use Integer and participate in the native local-origin journal.
func promoteLengthIntegerLocals(file *ast.File, info *types.Info, lengths map[types.Object]bool) int {
	units := 0
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || spec.Type != nil || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		object := info.Defs[spec.Names[0]]
		if object != nil && lengths[object] && types.Identical(object.Type(), types.Typ[types.Int]) {
			spec.Type = ast.NewIdent("int64")
			units++
		}
		return true
	})
	return units
}
