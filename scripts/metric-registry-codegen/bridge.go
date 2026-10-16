package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type bridgeBinding struct{ recordID, target, activity, function, owner string }

type bridgeProjection struct{ recordID, code string }

func renderBridge(records []bodycodegen.RecordType, b bridgeBinding, directory string) (bridgeProjection, error) {
	if b.recordID == "" {
		return bridgeProjection{}, fmt.Errorf("record binding requires a stable ID")
	}
	var selected *bodycodegen.RecordType
	for i := range records {
		if records[i].ID == b.recordID {
			if selected != nil {
				return bridgeProjection{}, fmt.Errorf("duplicated record ID %s", b.recordID)
			}
			selected = &records[i]
		}
	}
	if selected == nil {
		return bridgeProjection{}, fmt.Errorf("record ID %s unavailable", b.recordID)
	}
	fields, err := targetFields(filepath.Join(directory, b.owner), b.target)
	if err != nil {
		return bridgeProjection{}, err
	}
	if len(fields) != len(selected.Fields) {
		return bridgeProjection{}, fmt.Errorf("record/target fields differ")
	}
	var assignments []string
	seen := map[string]bool{}
	for _, field := range selected.Fields {
		target, ok := fields[field.ID]
		if field.ID == "" || field.GoName == "" || seen[field.ID] || !ok {
			return bridgeProjection{}, fmt.Errorf("record field ID %q has no unique target", field.ID)
		}
		seen[field.ID] = true
		if field.Presence != "" || field.Cardinality != "" {
			return bridgeProjection{}, fmt.Errorf("record field ID %q requires one mandatory value", field.ID)
		}
		value := "row." + field.GoName
		if target.kind == "int" && field.TypeID == "urn:gooo:type:integer" {
			value = "int(" + value + ")"
		} else if !compatibleType(target.kind, field.TypeID) {
			return bridgeProjection{}, fmt.Errorf("record field ID %q changes type", field.ID)
		}
		assignments = append(assignments, target.name+": "+value)
	}
	code := fmt.Sprintf("func %s() []%s {\n values := make([]%s, %sCount(0))\n for index := range values {\n"+
		"row := %sAt(int64(index))\n values[index] = %s{%s}\n }\n return values\n }\n"+
		"\n", b.function, b.target, b.target, b.activity, b.activity, b.target,
		strings.Join(assignments, ", "))
	return bridgeProjection{recordID: selected.ID, code: code}, nil
}

func compatibleType(kind, typeID string) bool {
	return kind == "string" && typeID == "urn:gooo:type:string" ||
		kind == "bool" && typeID == "urn:gooo:type:boolean"
}
