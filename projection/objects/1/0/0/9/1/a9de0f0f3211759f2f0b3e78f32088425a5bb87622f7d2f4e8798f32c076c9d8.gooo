package analyzer

import (
	"go/ast"
)

type resolution struct {
	state   resolutionState
	entries []Registration
}

func (r *resolver) resolve(ref SymbolRef) resolution {
	return makeResolution(r.lookup(ref))
}
func (r *resolver) resolveExpression(expr ast.Expr, file parsedFile, varTypes map[string]typeReference) resolution {
	switch expression := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return r.resolve(SymbolRef{
			PackagePath: file.packagePath,
			PackageName: file.packageName,
			Name:        expression.Name,
		})
	case *ast.SelectorExpr:
		return r.resolveSelector(expression, file, varTypes)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return r.resolveExpression(indexBase(expr), file, varTypes)
	default:
		return resolution{state: unresolved}
	}
}
func (r *resolver) resolveSelector(selector *ast.SelectorExpr, file parsedFile, varTypes map[string]typeReference) resolution {
	base := unwrapExpr(selector.X)
	if ident, ok := base.(*ast.Ident); ok {
		if typeRef, typed := varTypes[ident.Name]; typed {
			return r.resolve(SymbolRef{
				PackagePath: typeRef.packagePath,
				PackageName: typeRef.packageName,
				Receiver:    typeRef.name,
				Name:        selector.Sel.Name,
			})
		}
		if path, imported := r.imports[file.file].aliases[ident.Name]; imported {
			return r.resolve(SymbolRef{PackagePath: path, PackageName: ident.Name, Name: selector.Sel.Name})
		}
	}
	return resolution{state: unresolved}
}
