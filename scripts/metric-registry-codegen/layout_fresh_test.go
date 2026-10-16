package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

func TestBoundedRegistryProjectionFresh1645(t *testing.T) {
	directory := "../../internal/meta/metricprogram"
	output := filepath.Join(t.TempDir(), "registry.generated.go")
	if err := generate(filepath.Join(directory, "registry.gooo"), output); err != nil {
		t.Fatal(err)
	}
	actual, expected := projectionFiles(t, filepath.Dir(output)), projectionFiles(t, directory)
	for name, raw := range actual {
		if !bytes.Equal(raw, expected[name]) {
			t.Fatal("bounded registry projection differs:", name)
		}
		if filepath.Ext(name) != ".go" {
			continue
		}
		if bytes.Count(raw, []byte{'\n'}) > 75 {
			t.Fatal("generated file exceeds the physical line budget:", name)
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, name, raw, 0)
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
