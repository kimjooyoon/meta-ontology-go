package analyzer

import (
	"go/ast"
)

func collectVariableType(result map[string]typeReference, declaration *ast.ValueSpec, file parsedFile, imports importTable) {
	ref, ok := typeReferenceFor(declaration.Type, file, imports)
	if !ok {
		return
	}
	for _, name := range declaration.Names {
		result[name.Name] = ref
	}
}
func typeReferenceFor(expr ast.Expr, file parsedFile, imports importTable) (typeReference, bool) {
	switch current := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return typeReference{packagePath: file.packagePath, packageName: file.packageName, name: current.Name}, true
	case *ast.SelectorExpr:
		base, ok := unwrapExpr(current.X).(*ast.Ident)
		if !ok {
			return typeReference{}, false
		}
		path, ok := imports.aliases[base.Name]
		if !ok {
			return typeReference{}, false
		}
		return typeReference{packagePath: path, packageName: base.Name, name: current.Sel.Name}, true
	case *ast.StarExpr:
		return typeReferenceFor(current.X, file, imports)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return typeReferenceFor(indexBase(expr), file, imports)
	default:
		return typeReference{}, false
	}
}
func receiverTypeName(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	return typeName(fields.List[0].Type)
}
func typeName(expr ast.Expr) string {
	switch current := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return current.Name
	case *ast.StarExpr:
		return typeName(current.X)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return typeName(indexBase(expr))
	default:
		return ""
	}
}
func unwrapExpr(expr ast.Expr) ast.Expr {
	for expr != nil {
		switch current := expr.(type) {
		case *ast.ParenExpr:
			expr = current.X
		case *ast.IndexExpr, *ast.IndexListExpr:
			expr = indexBase(expr)
		default:
			return expr
		}
	}
	return nil
}
