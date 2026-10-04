package bodyexecution

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func compositionStepFunction(generation bodycodegen.Result, node CompositionActivity, index int) (string, error) {
	if len(generation.Report.RecordTypes) == 0 {
		return compositionFunction(generation.Source, node.Name, index)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "activity.go", generation.Source, parser.ParseComments)
	if err != nil || len(file.Imports) != 0 || len(file.Decls) != len(generation.Report.RecordTypes)+1 {
		return "", fmt.Errorf("record composition requires declared value structs and one pure function")
	}
	if err := checkCompositionRecordDeclarations(file, fset, generation.Report.RecordTypes); err != nil {
		return "", err
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != node.Name || function.Recv != nil {
			continue
		}
		start, end := fset.Position(function.Pos()).Offset, fset.Position(function.End()).Offset
		nameStart, nameEnd := fset.Position(function.Name.Pos()).Offset, fset.Position(function.Name.End()).Offset
		text := generation.Source[start:nameStart] + fmt.Sprintf("GoooComposedActivity%d", index) +
			generation.Source[nameEnd:end]
		return fmt.Sprintf("\n//gooo:generated:start id=%q kind=\"activity\"\n%s\n"+
			"//gooo:generated:end id=%q kind=\"activity\"\n", node.ID, text, node.ID), nil
	}
	return "", fmt.Errorf("record composition function differs from declared activity")
}

func checkCompositionRecordDeclarations(file *ast.File, fset *token.FileSet, records []bodycodegen.RecordType) error {
	expectedSet := token.NewFileSet()
	expected, err := parser.ParseFile(expectedSet, "records.go", "package p\n"+
		bodycodegen.RecordDeclarations(records, true), 0)
	if err != nil {
		return err
	}
	shapes := make(map[string]string)
	for _, declaration := range expected.Decls {
		spec := declaration.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
		shapes[spec.Name.Name] = recordTypeShape(expectedSet, spec)
	}
	seen := 0
	for _, declaration := range file.Decls {
		if _, ok := declaration.(*ast.FuncDecl); ok {
			continue
		}
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE || len(general.Specs) != 1 {
			return fmt.Errorf("composition accepts only source-declared record types")
		}
		spec, ok := general.Specs[0].(*ast.TypeSpec)
		if !ok || shapes[spec.Name.Name] == "" || shapes[spec.Name.Name] != recordTypeShape(fset, spec) {
			return fmt.Errorf("composition record layout differs from its declared fields")
		}
		delete(shapes, spec.Name.Name)
		seen++
	}
	if seen != len(records) {
		return fmt.Errorf("composition record declarations are incomplete")
	}
	return nil
}

func recordTypeShape(fset *token.FileSet, spec *ast.TypeSpec) string {
	var out bytes.Buffer
	if err := format.Node(&out, fset, spec); err != nil {
		return ""
	}
	return out.String()
}
