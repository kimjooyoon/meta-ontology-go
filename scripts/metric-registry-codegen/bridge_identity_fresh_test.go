package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestStableBridgeIdentityFresh1699(t *testing.T) {
	for _, tc := range []struct{ mode, reason string }{
		{"original", ""}, {"source-field-rename", ""}, {"source-record-rename", ""}, {"target-field-rename", ""},
		{"changed-record-ID", "record ID registry://record unavailable"},
		{"changed-field-ID", "record field ID \"registry://other\" has no unique target"},
		{"missing-field-ID", "record field ID \"\" has no unique target"},
		{"missing-native-name", "record field ID \"registry://value\" has no unique target"},
		{"duplicate-record-ID", "duplicated record ID registry://record"},
		{"duplicate-field-ID", "record field ID \"registry://value\" has no unique target"},
		{"optional-field", "record field ID \"registry://value\" requires one mandatory value"},
		{"repeated-field", "record field ID \"registry://value\" requires one mandatory value"},
		{"changed-type", "record field ID \"registry://value\" changes type"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			records, fields, targetName := bridgeIdentityCaseFresh1699(tc.mode)
			directory := t.TempDir()
			writeBridgeFixtureFresh1699(t, filepath.Join(directory, "row.go"), "package fixture; type Row struct {"+fields+"}")
			projection, err := renderBridge(records, bridgeBinding{
				"registry://record", "Row", "Item", "items", "row.go",
			}, directory)
			if tc.reason != "" {
				if err == nil || err.Error() != tc.reason {
					t.Fatalf("expected %q, got %v", tc.reason, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if projection.recordID != "registry://record" || !strings.Contains(projection.code, targetName+": row.StableValue") {
				t.Fatalf("stable-ID bridge lost its source field: %+v", projection)
			}
		})
	}
}

func bridgeIdentityCaseFresh1699(mode string) ([]bodycodegen.RecordType, string, string) {
	field := bodycodegen.RecordField{Name: "Value", ID: "registry://value", GoName: "StableValue", TypeID: "urn:gooo:type:string"}
	record := bodycodegen.RecordType{Name: "Declared", ID: "registry://record", Fields: []bodycodegen.RecordField{field}}
	fields, targetName := "Value string `gooo:\"registry://value\"`", "Value"
	switch mode {
	case "source-field-rename":
		record.Fields[0].Name = "Renamed"
	case "source-record-rename":
		record.Name = "OtherDisplayName"
	case "target-field-rename":
		fields, targetName = "NativeName string `gooo:\"registry://value\"`", "NativeName"
	case "changed-record-ID":
		record.ID = "registry://other-record"
	case "changed-field-ID":
		record.Fields[0].ID = "registry://other"
	case "missing-field-ID":
		record.Fields[0].ID = ""
	case "missing-native-name":
		record.Fields[0].GoName = ""
	case "duplicate-record-ID":
		other := record
		other.Name = "OtherDisplayName"
		return []bodycodegen.RecordType{record, other}, fields, targetName
	case "duplicate-field-ID":
		record.Fields = append(record.Fields, field)
		fields += "; Other string `gooo:\"registry://other\"`"
	case "optional-field":
		record.Fields[0].Presence = "optional"
	case "repeated-field":
		record.Fields[0].Cardinality = "many"
	case "changed-type":
		record.Fields[0].TypeID = "urn:gooo:type:boolean"
	}
	return []bodycodegen.RecordType{record}, fields, targetName
}

func writeBridgeFixtureFresh1699(t *testing.T, path, source string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}
