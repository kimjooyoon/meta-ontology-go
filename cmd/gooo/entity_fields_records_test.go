package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecordFieldsOrdinaryCLIGenerationAndUnchangedSource(t *testing.T) {
	const source = `package recordprojection
namespace recordprojection
entity Basket id "urn:projection:basket" fields {
 field One id "urn:projection:one" type "urn:projection:line" required one
 field Optional id "urn:projection:optional" type Line optional one
 field Many id "urn:projection:many" type Line required many
 field OptionalMany id "urn:projection:optional-many" type Line optional many
}
entity Line id "urn:projection:line" fields {
 field Quantity id "urn:projection:quantity" type integer required one
}
activity Pass(Basket) -> Basket id "urn:projection:pass" computes "return input"
activity Deliver(Basket) -> Basket id "urn:projection:deliver" computes "return input"
bind Pass.result -> Deliver.input
`
	directory := t.TempDir()
	input, output := filepath.Join(directory, "기원.gooo"), filepath.Join(directory, "output")
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runCheck([]string{input}, OSFileReader{}, EntityFieldsCLIParser{}, &stdout, &stderr); code != exitOK {
		t.Fatal("ordinary source check", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{input, "--out", output}, OSFileReader{}, EntityFieldsCLIParser{}, &stdout, &stderr); code != exitOK {
		t.Fatal("ordinary record projection", code, stdout.String(), stderr.String())
	}
	generated, err := os.ReadFile(filepath.Join(output, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	assertRecordCLIFieldTypes(t, generated)
	if _, err := os.Stat(filepath.Join(output, "runtime-plan.json")); err != nil {
		t.Fatal("source binding plan was not projected", err)
	}
	after, err := os.ReadFile(input)
	if err != nil || string(after) != source {
		t.Fatal("projection changed original source", err)
	}
}

func assertRecordCLIFieldTypes(t *testing.T, source []byte) {
	t.Helper()
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "projection.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if ok && spec.Name.Name == "Basket" {
			for _, field := range spec.Type.(*ast.StructType).Fields.List {
				var spelling bytes.Buffer
				if err := format.Node(&spelling, set, field.Type); err != nil {
					t.Fatal(err)
				}
				got[field.Names[0].Name] = spelling.String()
			}
		}
		return true
	})
	want := map[string]string{"One": "Line", "Optional": "*Line", "Many": "[]Line", "OptionalMany": "*[]Line"}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("nominal shapes missing from ordinary Go projection", got)
	}
}
