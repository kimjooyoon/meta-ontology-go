package bodycodegen

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// RecordType retains source names, stable identities and the typed value layout.
type RecordType struct {
	Name   string        `json:"name"`
	ID     string        `json:"id"`
	GoName string        `json:"go_name"`
	Fields []RecordField `json:"fields"`
}

type RecordField struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	TypeID string `json:"type_id"`
	GoName string `json:"go_name"`
}

// RecordTypesFromModel accepts a model already validated under the body profile.
// Native value construction additionally requires names usable in body syntax.
func RecordTypesFromModel(model bidir.Model) ([]RecordType, error) {
	var result []RecordType
	for _, node := range model.Nodes {
		if node.Kind != bidir.EntityKind || len(node.Fields) == 0 {
			continue
		}
		if _, scalar := goTypeForEntity(node.Name); scalar {
			return nil, fmt.Errorf("scalar entity %q cannot also declare record fields", node.Name)
		}
		if !recordIdentifier(node.Name) || len(node.Fields) > 16 || len(result) >= 16 {
			return nil, fmt.Errorf("record bodies require 1..16 records, 1..16 fields and identifier names")
		}
		record := RecordType{Name: node.Name, ID: string(node.ID), GoName: recordGoName("Record", string(node.ID))}
		for _, field := range node.Fields {
			if !recordIdentifier(field.Name) {
				return nil, fmt.Errorf("record %q field %q requires an identifier in body syntax", node.Name, field.Name)
			}
			typeID := string(field.TypeRef.ID)
			if typeID == "" {
				typeID = string(field.TypeRefUse.ResolvedID)
			}
			if typeID != string(semantic.BuiltinStringTypeID) && typeID != string(semantic.BuiltinBooleanTypeID) {
				return nil, fmt.Errorf("record %q field %q has unsupported type %q", node.Name, field.Name, typeID)
			}
			record.Fields = append(record.Fields, RecordField{Name: field.Name, ID: string(field.ID),
				TypeID: typeID, GoName: recordGoName("Field", string(field.ID))})
		}
		result = append(result, record)
	}
	return result, nil
}

func recordIdentifier(name string) bool {
	return token.IsIdentifier(name) && name != "_" && !strings.HasPrefix(name, "_gooo") &&
		name != "int64" && name != "bool" && name != "string" && name != "true" && name != "false"
}

func recordGoName(kind, id string) string {
	return "Gooo" + kind + strings.TrimPrefix(digest([]byte(id)), "sha256:")
}

// ParseBodyFile activates the existing field profile only for the pure body
// entry points. The ordinary parser and other profiles keep their own contracts.
func ParseBodyFile(filename string, source []byte) (*syntax.File, syntax.Diagnostics) {
	return syntax.ParseFileWithEntityFieldsSupport(filename, string(source), syntax.EntityFieldsV2Support())
}

func bodyEntityType(name string, records []RecordType) (string, bool) {
	if scalar, ok := goTypeForEntity(name); ok {
		return scalar, true
	}
	for _, record := range records {
		if record.Name == name {
			return name, true
		}
	}
	return "", false
}

func supportedBodyType(name string, records []RecordType) bool {
	if name == "int64" || name == "bool" || name == "string" {
		return true
	}
	_, ok := bodyEntityType(name, records)
	return ok
}

func resolveBodyModel(file *syntax.File) (bidir.Model, []RecordType, error) {
	support := bidir.EntityFieldsV2Support()
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		return bidir.Model{}, nil, fmt.Errorf("lower activity identity: %w", err)
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		return model, nil, fmt.Errorf("resolve activity identity: %w", err)
	}
	records, err := RecordTypesFromModel(model)
	return model, records, err
}
