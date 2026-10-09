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
	Name     string `json:"name"`
	ID       string `json:"id"`
	TypeID   string `json:"type_id"`
	GoName   string `json:"go_name"`
	Presence string `json:"presence,omitempty"`
}

// RecordTypesFromModel accepts a model already validated under the body profile.
// Native value construction additionally requires names usable in body syntax.
func RecordTypesFromModel(model bidir.Model) ([]RecordType, error) {
	var result []RecordType
	for _, node := range model.Nodes {
		if node.Kind != bidir.EntityKind || len(node.Fields) == 0 {
			continue
		}
		if ScalarEntityKind(node.Name, string(node.ID)) != "" {
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
			if typeID != string(semantic.BuiltinStringTypeID) && typeID != string(semantic.BuiltinBooleanTypeID) && typeID != string(semantic.BuiltinIntegerTypeID) {
				return nil, fmt.Errorf("record %q field %q has unsupported type %q", node.Name, field.Name, typeID)
			}
			presence := ""
			if field.Presence == semantic.Optional {
				presence = string(field.Presence)
			}
			record.Fields = append(record.Fields, RecordField{Name: field.Name, ID: string(field.ID),
				TypeID: typeID, GoName: recordGoName("Field", string(field.ID)), Presence: presence})
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
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), syntax.EntityFieldsV3Support())
	if !diagnostics.HasErrors() {
		return file, diagnostics
	}
	// Preserve the V3 source profile for existing required-field programs. Only
	// use V4 when V3 rejected the source and V4 accepts its optional scalar fields.
	v4File, v4Diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), syntax.EntityFieldsV4Support())
	if v4File != nil && !v4Diagnostics.HasErrors() {
		return v4File, v4Diagnostics
	}
	return file, diagnostics
}

func bodyEntityType(name string, records []RecordType, files ...*syntax.File) (string, bool) {
	var file *syntax.File
	if len(files) != 0 {
		file = files[0]
	}
	if scalar, ok := goTypeForEntity(sourceScalarKind(file, name)); ok {
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
	support := bodyEntityFieldsSupport(file)
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

func bodyEntityFieldsSupport(file *syntax.File) bidir.EntityFieldsSupport {
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		for _, field := range entity.Fields {
			if field.Presence == syntax.FieldPresenceOptional {
				return bidir.EntityFieldsV4Support()
			}
		}
	}
	return bidir.EntityFieldsV3Support()
}

// BodyEntityFieldsSupport selects V4 only when the source declares optional
// fields, keeping existing required-field body fingerprints on V3.
func BodyEntityFieldsSupport(file *syntax.File) bidir.EntityFieldsSupport {
	return bodyEntityFieldsSupport(file)
}
