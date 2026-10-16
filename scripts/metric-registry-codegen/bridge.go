package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type bridgeBinding struct{ record, target, activity, function, owner string }

type bridgeProjection struct{ recordID, code string }

func renderBridge(records []bodycodegen.RecordType, b bridgeBinding, directory string) (bridgeProjection, error) {
	var selected *bodycodegen.RecordType
	for i := range records {
		if records[i].Name == b.record {
			if selected != nil {
				return bridgeProjection{}, fmt.Errorf("duplicated record %s", b.record)
			}
			selected = &records[i]
		}
	}
	if selected == nil {
		return bridgeProjection{}, fmt.Errorf("record %s unavailable", b.record)
	}
	fields, err := targetFields(filepath.Join(directory, b.owner), b.target)
	if err != nil {
		return bridgeProjection{}, err
	}
	if len(fields) != len(selected.Fields) {
		return bridgeProjection{}, fmt.Errorf("record/target fields differ")
	}
	var assignments []string
	for _, field := range selected.Fields {
		kind, ok := fields[field.Name]
		if !ok {
			return bridgeProjection{}, fmt.Errorf("record field %s has no target", field.Name)
		}
		value := "row." + field.GoName
		if kind == "int" && field.TypeID == "urn:gooo:type:integer" {
			value = "int(" + value + ")"
		} else if !compatibleType(kind, field.TypeID) {
			return bridgeProjection{}, fmt.Errorf("record field %s changes type", field.Name)
		}
		assignments = append(assignments, field.Name+": "+value)
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

func targetFields(path, name string) (map[string]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok || spec.Name.Name != name {
			return true
		}
		shape, ok := spec.Type.(*ast.StructType)
		if !ok || result != nil {
			err = fmt.Errorf("target type is ambiguous or not a record")
			return false
		}
		result = map[string]string{}
		for _, field := range shape.Fields.List {
			kind, ok := field.Type.(*ast.Ident)
			if !ok || len(field.Names) != 1 {
				err = fmt.Errorf("target field is not one named scalar")
				return false
			}
			result[field.Names[0].Name] = kind.Name
		}
		return false
	})
	if result == nil && err == nil {
		err = fmt.Errorf("target type unavailable")
	}
	return result, err
}
