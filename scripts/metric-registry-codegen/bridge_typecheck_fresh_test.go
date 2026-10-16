package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRenamedBridgeTypesFresh1705(t *testing.T) {
	for _, tc := range []struct{ kind, native, target, literal string }{
		{"string", "string", "string", `"sample"`},
		{"boolean", "bool", "bool", "true"},
		{"integer", "int64", "int", "7"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			directory := t.TempDir()
			target := "type Row struct { NativeValue " + tc.target + " `gooo:\"registry://value\"` }"
			writeBridgeFixtureFresh1699(t, filepath.Join(directory, "row.go"), "package fixture; "+target)
			record := bodycodegen.RecordType{Name: "SourceLabel", ID: "registry://record", Fields: []bodycodegen.RecordField{
				{Name: "SourceValue", ID: "registry://value", TypeID: "urn:gooo:type:" + tc.kind, GoName: "ProjectedValue"},
			}}
			projection, err := renderBridge([]bodycodegen.RecordType{record}, bridgeBinding{
				"registry://record", "Row", "Item", "items", "row.go",
			}, directory)
			if err != nil {
				t.Fatal(err)
			}
			source := fmt.Sprintf("package fixture; %s; type Projected struct { ProjectedValue %s }; "+
				"func ItemCount(int64) int64 { return 1 }; func ItemAt(int64) Projected { return Projected{%s} }; %s",
				target, tc.native, tc.literal, projection.code)
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "bridge.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			configuration := types.Config{GoVersion: "go1.27"}
			if _, err := configuration.Check("fixture", fset, []*ast.File{file}, nil); err != nil {
				t.Fatalf("renamed bridge does not typecheck: %v", err)
			}
		})
	}
}
