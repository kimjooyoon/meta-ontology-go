package bodycodegen

import (
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
)

// Existing well-typed length expressions retain their Go projection. Only a
// runtime length flowing into an Integer context needs the widening conversion.
func normalizeLengthIntegerContexts(packageName string, file *ast.File, fset *token.FileSet) int {
	if !bodyMayUseLength(file) {
		return 0
	}
	units := 0
	for remaining := semanticUnitCount(file) + 1; remaining > 0; remaining-- {
		info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue),
			Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}
		configuration := types.Config{Importer: importer.Default(), Error: func(error) {}}
		if _, err := configuration.Check(packageName, fset, []*ast.File{file}, info); err == nil {
			return units
		}
		context := lengthIntegerContext{info: info, lengths: runtimeLengthBindings(file, info)}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			context.output = nil
			if object := info.Defs[function.Name]; object != nil {
				if signature, ok := object.Type().(*types.Signature); ok && signature.Results().Len() == 1 {
					context.output = signature.Results().At(0).Type()
				}
			}
			ast.Inspect(function.Body, context.visit)
		}
		if context.units == 0 {
			return units
		}
		units += context.units + promoteLengthIntegerLocals(file, info, context.lengths)
	}
	return units
}

type lengthIntegerContext struct {
	info    *types.Info
	lengths map[types.Object]bool
	output  types.Type
	units   int
}

func (c *lengthIntegerContext) widen(expression ast.Expr, expected types.Type) ast.Expr {
	value, known := c.info.Types[expression]
	if !known || value.Value != nil || !types.Identical(value.Type, types.Typ[types.Int]) ||
		!types.Identical(expected, types.Typ[types.Int64]) ||
		!dependsOnRuntimeLength(expression, c.info, c.lengths) {
		return expression
	}
	c.units += 2
	return &ast.CallExpr{Fun: &ast.Ident{Name: "int64", NamePos: expression.Pos()},
		Args: []ast.Expr{expression}, Rparen: expression.End() - 1}
}

func isImplicitLengthConversion(call *ast.CallExpr) bool {
	name, ok := call.Fun.(*ast.Ident)
	return ok && name.Name == "int64" && len(call.Args) == 1 && call.Lparen == token.NoPos &&
		name.Pos() == call.Args[0].Pos() && call.Rparen == call.Args[0].End()-1
}
