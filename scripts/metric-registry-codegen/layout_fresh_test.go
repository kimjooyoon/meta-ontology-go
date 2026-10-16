package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestBoundedRegistryProjectionFresh1645(t *testing.T) {
	directory := "../../internal/meta/metricprogram"
	output := filepath.Join(t.TempDir(), "registry.generated.go")
	if err := generate(filepath.Join(directory, "registry.gooo"), output); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join(directory, "registry.generated.go"))
	if err != nil || !bytes.Equal(raw, expected) {
		t.Fatal("bounded registry projection differs", err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, output, raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			lines := fset.Position(function.End()).Line - fset.Position(function.Pos()).Line + 1
			if lines > 75 {
				t.Fatalf("generated %s exceeds the declaration bound: %d", function.Name.Name, lines)
			}
		}
	}
}

func TestRecordLayoutTokensFresh1645(t *testing.T) {
	before := []byte("package demo\nfunc read() any { return GoooRecordFixture{\n Name: \"origin\",\n N: 37,\n} }\n")
	after, err := compactRecordLiterals(before)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(after, []byte(`GoooRecordFixture{Name: "origin", N: 37}`)) {
		t.Fatalf("record fields changed during layout projection: %s", after)
	}
}
