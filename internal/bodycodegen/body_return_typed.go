package bodycodegen

import (
	"go/ast"
	"go/types"
)

func (c *bodyReturnCapture) captureTyped(file *ast.File, info *types.Info, records []RecordType) {
	if c == nil {
		return
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		WalkReturnSyntax(function.Body, func(statement *ast.ReturnStmt) {
			named, ok := info.TypeOf(statement.Results[0]).(*types.Named)
			if !ok {
				return
			}
			for _, record := range records {
				if named.Obj().Name() != record.Name {
					continue
				}
				origin, retained := c.origins[statement]
				if !retained {
					origin.ActivityID = c.activities[function.Name.Name]
					origin.Reason = "RETURN_EXPRESSION_NOT_RETAINED"
				}
				origin.RecordID = record.ID
				c.origins[statement] = origin
			}
		})
	}
}
