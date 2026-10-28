package generator

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// EntityFieldTypeNames projects stable type IDs to native names under the exact
// profile. Field provenance and graph validity remain separate obligations.
func EntityFieldTypeNames(ir SemanticIR, profile syntax.EntityFieldsProfile) (map[string]string, error) {
	capabilities, err := profile.Capabilities()
	if err != nil {
		return nil, err
	}
	types := map[string]string{entityFieldsStringTypeID: "string"}
	if capabilities.Boolean {
		types[entityFieldsBooleanTypeID] = "bool"
	}
	if capabilities.Integer {
		types[entityFieldsIntegerTypeID] = "int64"
	}
	if !capabilities.Record {
		return types, nil
	}
	reserved := map[string]bool{"_": true}
	for _, name := range types {
		reserved[name] = true
	}
	for _, entity := range ir.Entities {
		name := entity.GoName
		if name == "" {
			name = entity.Name
		}
		if reserved[name] {
			return nil, entityFieldsError(entityFieldsGoNameCollisionDiagnostic, Field{}, fmt.Sprintf("entity %q shadows native type %q", entity.ID, name))
		}
		if len(entity.Fields) == 0 {
			continue
		}
		if !isGoIdentifier(name) {
			return nil, entityFieldsError(entityFieldsGoNameCollisionDiagnostic, Field{}, fmt.Sprintf("record %q has invalid Go name %q", entity.ID, name))
		}
		if _, exists := types[entity.ID]; exists {
			return nil, entityFieldsError(entityFieldsIDCollisionDiagnostic, Field{}, fmt.Sprintf("record identity %q is already a type", entity.ID))
		}
		types[entity.ID] = name
	}
	return types, nil
}

func projectedEntityFieldType(field Field, base string, capabilities syntax.EntityFieldsCapabilities) string {
	if field.Cardinality == "many" && capabilities.Many {
		base = "[]" + base
	}
	if field.Presence == "optional" && capabilities.Optional {
		base = "*" + base
	}
	return base
}

func validateRecordFieldReferences(ir SemanticIR, capabilities syntax.EntityFieldsCapabilities) error {
	if !capabilities.Record {
		return nil
	}
	nodes := make([]semantic.Node, 0, len(ir.Entities))
	fields := map[semantic.ID]Field{}
	for _, entity := range ir.Entities {
		node := semantic.Node{ID: semantic.ID(entity.ID), Kind: semantic.Entity}
		for _, field := range entity.Fields {
			id := semantic.ID(field.ID)
			fields[id] = field
			node.Fields = append(node.Fields, semantic.Field{ID: id, TypeRef: semantic.TypeRef{ID: semantic.ID(field.TypeRefID)}})
		}
		nodes = append(nodes, node)
	}
	err := semantic.ValidateRecordReferences(nodes, syntax.RecordReferenceDepthLimit)
	if issues, ok := err.(*semantic.ValidationErrors); ok && len(issues.Issues) != 0 {
		issue := issues.Issues[0]
		return entityFieldsError(entityFieldsUnsupportedShapeDiagnostic, fields[issue.Subject], issue.Message)
	}
	return err
}
