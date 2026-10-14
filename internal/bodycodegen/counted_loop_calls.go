package bodycodegen

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type pureCallSite struct {
	name    string
	maximum int
}

func directPureCallSites(body string) ([]pureCallSite, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "calls.go", "package p\nfunc body(){\n"+body+"\n}", parser.AllErrors)
	if err != nil {
		return nil, err
	}
	var sites []pureCallSite
	err = collectPureCallSites(file, 1, &sites)
	return sites, err
}

func collectPureCallSites(root ast.Node, maximum int, sites *[]pureCallSite) error {
	var err error
	ast.Inspect(root, func(node ast.Node) bool {
		if err != nil {
			return false
		}
		switch value := node.(type) {
		case *ast.RangeStmt:
			if _, shapeErr := collectionLoopShape(value); shapeErr != nil {
				err = shapeErr
			} else {
				err = collectPureCallSites(value.X, maximum, sites)
				if err == nil {
					err = collectPureCallSites(value.Body, pureCallMultiplicity(maximum, RepeatedFieldLimit), sites)
				}
			}
			return false
		case *ast.ForStmt:
			_, bound, shapeErr := countedLoopShape(value)
			if shapeErr != nil {
				err = shapeErr
			} else {
				err = collectPureCallSites(value.Body, pureCallMultiplicity(maximum, bound), sites)
			}
			return false
		case *ast.ReturnStmt:
			for _, expression := range value.Results {
				if err = collectPureCallSites(expression, 1, sites); err != nil {
					break
				}
			}
			return false
		case *ast.CallExpr:
			if err = validatePureCallExpression(value); err == nil {
				*sites = append(*sites, pureCallSite{name: value.Fun.(*ast.Ident).Name, maximum: maximum})
			}
		}
		return err == nil
	})
	return err
}

// Saturate loop multiplicity above the existing invocation budget. A return
// expression executes at most once in this activity, even inside nested loops.
// Primitive-only loops do not acquire a named-activity call budget.
func pureCallMultiplicity(maximum, bound int) int {
	if maximum > pureCallLimit/bound {
		return pureCallLimit + 1
	}
	return maximum * bound
}
