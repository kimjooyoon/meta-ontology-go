package main

import (
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorbinding"
)

func classify(expression ast.Expr, constants map[string]ast.Expr, parameters map[string]bool) predecessorbinding.State {
	if isConstant(expression, constants, make(map[string]bool)) {
		return predecessorbinding.StateStaticLiteral
	}
	selector, ok := unparen(expression).(*ast.SelectorExpr)
	if ok {
		root, rootOK := unparen(selector.X).(*ast.Ident)
		if rootOK && parameters[root.Name] {
			return predecessorbinding.StateDynamicInput
		}
	}
	return predecessorbinding.StateUnknown
}

func isConstant(expression ast.Expr, constants map[string]ast.Expr, seen map[string]bool) bool {
	switch value := unparen(expression).(type) {
	case *ast.BasicLit:
		return true
	case *ast.UnaryExpr:
		return isConstant(value.X, constants, seen)
	case *ast.BinaryExpr:
		return isConstant(value.X, constants, seen) && isConstant(value.Y, constants, seen)
	case *ast.Ident:
		if seen[value.Name] {
			return false
		}
		next, ok := constants[value.Name]
		if !ok {
			return false
		}
		seen[value.Name] = true
		result := isConstant(next, constants, seen)
		delete(seen, value.Name)
		return result
	default:
		return false
	}
}

func unparen(expression ast.Expr) ast.Expr {
	for {
		parenthesized, ok := expression.(*ast.ParenExpr)
		if !ok {
			return expression
		}
		expression = parenthesized.X
	}
}

func isBaselineReference(expression ast.Expr) bool {
	identifier, ok := unparen(expression).(*ast.Ident)
	return ok && identifier.Name == "BaselineReference"
}
