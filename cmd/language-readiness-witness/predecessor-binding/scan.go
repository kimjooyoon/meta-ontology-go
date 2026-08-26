package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorbinding"
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

func unknownObservations() []predecessorbinding.Observation {
	result := make([]predecessorbinding.Observation, 0, predecessorbinding.Total)
	for _, coordinate := range predecessorbinding.Registry() {
		result = append(result, predecessorbinding.Observation{ID: coordinate.ID,
			GoField: coordinate.GoField, SourcePath: predecessorbinding.SourcePath,
			Provider: predecessorbinding.Provider, State: predecessorbinding.StateUnknown})
	}
	return result
}
