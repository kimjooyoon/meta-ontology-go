package semanticbinding

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func recordsForResult(t *testing.T, name string, source []byte, result Result) []fixtureRecord {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, name+".go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse fixture for oracle comparison: %v", err)
	}
	nameAt := func(span Span) string {
		for _, declaration := range file.Decls {
			if spanFor(fileSet, declaration) == span {
				return declarationName(declaration)
			}
			if group, ok := declaration.(*ast.GenDecl); ok {
				for _, specification := range group.Specs {
					if spanFor(fileSet, specification) == span {
						return declarationName(specification)
					}
				}
			}
		}
		t.Fatalf("no declaration for result span %s", span)
		return ""
	}
	return appendBindingRecords(result.Bindings, result.Obligations, nameAt)
}
func appendBindingRecords(bindings []Binding, obligations []Obligation, nameAt func(Span) string) []fixtureRecord {
	records := make([]fixtureRecord, 0, len(bindings)+len(obligations))
	for _, binding := range bindings {
		records = append(records, fixtureRecord{
			Directive: "bind", Name: nameAt(binding.Span), ID: binding.ID,
			Location: locationForSpan(binding.Span),
		})
	}
	for _, obligation := range obligations {
		records = append(records, fixtureRecord{
			Directive: "obligation", Name: nameAt(obligation.Span), ID: obligation.ID,
			Subject: obligation.Subject, Pressure: obligation.Pressure,
			Location: locationForSpan(obligation.Span),
		})
	}
	return records
}
func declarationName(node ast.Node) string {
	switch current := node.(type) {
	case *ast.FuncDecl:
		return current.Name.Name
	case *ast.GenDecl:
		if len(current.Specs) == 1 {
			return declarationName(current.Specs[0])
		}
	case *ast.TypeSpec:
		return current.Name.Name
	case *ast.ValueSpec:
		if len(current.Names) == 1 {
			return current.Names[0].Name
		}
	}
	return ""
}
