package main

import (
	"fmt"
	"go/ast"
	"go/parser"
)

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
			expression, err := parser.ParseExpr(fmt.Sprintf("[]MetaOperation{{ID:%q, Activity:%q, ProofChoice:ProofRegression}}", spec.MetaOperation, spec.Activity))
			if err != nil {
				return err
			}
			literal.Elts = append(literal.Elts, expression.(*ast.CompositeLit).Elts[0])
			return nil
		}
	}
	return fmt.Errorf("CanonicalMetaOperations return literal not found")
}
