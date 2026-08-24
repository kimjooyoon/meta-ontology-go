package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
)

const rawMetricID = "gooo.metric.evidence.raw-reconstruction.v1"

func transformRegistry(path string, spec metricSpec) ([]byte, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	files, file, err := parseSource(path, source)
	if err != nil {
		return nil, err
	}
	if !hasString(file, spec.MetricID) {
		return nil, fmt.Errorf("write-set denominator entry not found")
	}
	if err := addOperatingOperation(file, spec); err != nil {
		return nil, err
	}
	if err := addMetaOperation(file, spec); err != nil {
		return nil, err
	}
	return formatSource(files, file)
}

func addOperatingOperation(file *ast.File, spec metricSpec) error {
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, item := range general.Specs {
			value, ok := item.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || value.Names[0].Name != "operatingOperations" || len(value.Values) != 1 {
				continue
			}
			literal, ok := value.Values[0].(*ast.CompositeLit)
			if !ok {
				return fmt.Errorf("operatingOperations is not a literal")
			}
			if hasString(literal, spec.MetricID) {
				return nil
			}
			parsed, err := parser.ParseExpr(fmt.Sprintf("map[string]string{%q:%q}", spec.MetricID, spec.MetaOperation))
			if err != nil {
				return err
			}
			literal.Elts = append(literal.Elts, parsed.(*ast.CompositeLit).Elts[0])
			return nil
		}
	}
	return fmt.Errorf("operatingOperations not found")
}

func addMetaOperation(file *ast.File, spec metricSpec) error {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "CanonicalMetaOperations" {
			continue
		}
		if hasString(function, spec.MetaOperation) {
			return nil
		}
		for _, statement := range function.Body.List {
			result, ok := statement.(*ast.ReturnStmt)
			if !ok || len(result.Results) != 1 {
				continue
			}
			literal, ok := result.Results[0].(*ast.CompositeLit)
			if !ok {
				continue
			}
			expression, err := parser.ParseExpr(fmt.Sprintf("MetaOperation{ID:%q, Activity:%q, ProofChoice:%q}", spec.MetaOperation, spec.Activity, spec.ProofChoice))
			if err != nil {
				return err
			}
			literal.Elts = append(literal.Elts, expression)
			return nil
		}
	}
	return fmt.Errorf("CanonicalMetaOperations return literal not found")
}
