package bodycodegen

import (
	"fmt"
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SelectActivityRecords bounds the executable closure while retaining the full
// validated source declaration view. Record IDs and source order stay intact.
func SelectActivityRecords(file *syntax.File, declarations []RecordType, roots []string) ([]RecordType, error) {
	if file == nil || len(roots) == 0 {
		return nil, fmt.Errorf("record scope requires source and selected activities")
	}
	closure, err := planCallConstructionFile(file, roots)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{}
	for _, name := range closure.order {
		activity := closure.activities[name]
		used[activity.Output] = true
		for _, input := range activity.Inputs {
			used[input.Name] = true
		}
		bodies, err := activityConstructionSyntax(activity)
		if err != nil {
			return nil, err
		}
		for _, body := range bodies {
			ast.Inspect(body, func(node ast.Node) bool {
				var expression ast.Expr
				switch value := node.(type) {
				case *ast.CompositeLit:
					expression = value.Type
				case *ast.ValueSpec:
					expression = value.Type
				case *ast.CallExpr:
					if ident, ok := value.Fun.(*ast.Ident); ok && closure.activities[ident.Name] == nil {
						expression = ident
					}
				}
				if expression != nil {
					ast.Inspect(expression, func(part ast.Node) bool {
						if ident, ok := part.(*ast.Ident); ok && ident.Obj == nil {
							used[ident.Name] = true
						}
						return true
					})
				}
				return true
			})
		}
	}
	return checkedRecordScope(declarations, used, "selected activity closure", false)
}
