package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRegistryBridgeFresh1637(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "row.go"),
		[]byte("package fixture; type Row struct { Value string `gooo:\"registry://declared/value\"` }"), 0600); err != nil {
		t.Fatal(err)
	}
	record := bodycodegen.RecordType{Name: "Declared", ID: "registry://declared", Fields: []bodycodegen.RecordField{
		{Name: "Value", ID: "registry://declared/value", GoName: "ProjectedValue", TypeID: "urn:gooo:type:string"},
	}}
	for _, item := range []struct {
		name    string
		records []bodycodegen.RecordType
		allowed bool
	}{
		{"typed-field", []bodycodegen.RecordType{record}, true},
		{"missing-record", nil, false},
		{"duplicate-record", []bodycodegen.RecordType{record, record}, false},
		{"type-mismatch", []bodycodegen.RecordType{{Name: "Declared", ID: "registry://declared", Fields: []bodycodegen.RecordField{
			{Name: "Value", ID: "registry://declared/value", GoName: "ProjectedValue", TypeID: "urn:gooo:type:boolean"},
		}}}, false},
	} {
		t.Run(item.name, func(t *testing.T) {
			projection, err := renderBridge(item.records, bridgeBinding{
				"registry://declared", "Row", "Item", "items", "row.go",
			}, directory)
			if (err == nil) != item.allowed {
				t.Fatalf("bridge admission: %v", err)
			}
			if item.allowed && projection.recordID != record.ID {
				t.Fatal("adapter lost source record identity")
			}
		})
	}
}
