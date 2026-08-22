package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/predecessorbinding"
)

func scan(root string) ([]predecessorbinding.Observation, error) {
	source := filepath.Join(root, filepath.FromSlash(predecessorbinding.SourcePath))
	parsed, err := parser.ParseFile(token.NewFileSet(), source, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse predecessor provider: %w", err)
	}
	constants := constantExpressions(parsed)
	var providers []*ast.FuncDecl
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == predecessorbinding.Provider {
			providers = append(providers, function)
		}
	}
	if len(providers) != 1 {
		return unknownObservations(), nil
	}
	parameters := parameterNames(providers[0])
	bindings := make(map[string][]predecessorbinding.State)
	ast.Inspect(providers[0].Body, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || !isBaselineReference(literal.Type) {
			return true
		}
		for _, element := range literal.Elts {
			binding, ok := element.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			field, ok := binding.Key.(*ast.Ident)
			if !ok {
				continue
			}
			bindings[field.Name] = append(bindings[field.Name], classify(binding.Value, constants, parameters))
		}
		return true
	})
	result := make([]predecessorbinding.Observation, 0, predecessorbinding.Total)
	for _, coordinate := range predecessorbinding.Registry() {
		state := predecessorbinding.StateUnknown
		if observed := bindings[coordinate.GoField]; len(observed) == 1 {
			state = observed[0]
		}
		result = append(result, predecessorbinding.Observation{ID: coordinate.ID,
			GoField: coordinate.GoField, SourcePath: predecessorbinding.SourcePath,
			Provider: predecessorbinding.Provider, State: state})
	}
	return result, nil
}

func constantExpressions(file *ast.File) map[string]ast.Expr {
	result := make(map[string]ast.Expr)
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, specification := range group.Specs {
			value, ok := specification.(*ast.ValueSpec)
			if !ok || len(value.Values) != len(value.Names) {
				continue
			}
			for index, name := range value.Names {
				result[name.Name] = value.Values[index]
			}
		}
	}
	return result
}

func parameterNames(function *ast.FuncDecl) map[string]bool {
	result := make(map[string]bool)
	if function.Type.Params == nil {
		return result
	}
	for _, field := range function.Type.Params.List {
		for _, name := range field.Names {
			result[name.Name] = true
		}
	}
	return result
}

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

func unknownObservations() []predecessorbinding.Observation {
	result := make([]predecessorbinding.Observation, 0, predecessorbinding.Total)
	for _, coordinate := range predecessorbinding.Registry() {
		result = append(result, predecessorbinding.Observation{ID: coordinate.ID,
			GoField: coordinate.GoField, SourcePath: predecessorbinding.SourcePath,
			Provider: predecessorbinding.Provider, State: predecessorbinding.StateUnknown})
	}
	return result
}
