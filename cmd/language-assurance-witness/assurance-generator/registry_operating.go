package main

import (
	"fmt"
	"go/ast"
	"go/parser"
)

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
