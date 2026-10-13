package bodycodegen

import (
	"go/ast"
	"go/types"
)

// Typed field objects retain identity before native field names are substituted.
// Repeated accesses remain separate references in their source traversal order.
func captureConditionFields(file *ast.File, info *types.Info, records []RecordType, origins map[ast.Expr]BodyConditionOrigin) {
	fields := map[types.Object]ConditionSourceField{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range general.Specs {
			named, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			for _, record := range records {
				if record.Name != named.Name.Name {
					continue
				}
				object := info.Defs[named.Name]
				if object == nil {
					continue
				}
				shape, ok := object.Type().Underlying().(*types.Struct)
				if !ok {
					continue
				}
				for i, field := range record.Fields {
					fields[shape.Field(i)] = ConditionSourceField{EntityID: record.ID, FieldID: field.ID, Name: field.Name}
				}
			}
		}
	}
	for expression, origin := range origins {
		ast.Inspect(expression, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if field, found := fields[info.Uses[selector.Sel]]; found {
				origin.Fields = append(origin.Fields, field)
			}
			return true
		})
		origins[expression] = origin
	}
}
