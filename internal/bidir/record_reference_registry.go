package bidir

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func documentRecordRegistry(document Document, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.TypeRegistry, []Node, error) {
	nodes := make([]Node, 0, len(document.Declarations))
	for _, declaration := range document.Declarations {
		if declaration.Kind != EntityKind || len(declaration.Fields) == 0 {
			continue
		}
		id, err := declarationIdentity(document.Namespace, declaration)
		if err != nil {
			return registry, nil, err
		}
		nodes = append(nodes, Node{ID: id, Kind: EntityKind, Name: declaration.Name, Fields: declaration.Fields, Span: declaration.Span})
	}
	result, err := modelRecordRegistry(document.Namespace, nodes, registry, support)
	return result, nodes, err
}

func modelRecordRegistry(namespace string, nodes []Node, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.TypeRegistry, error) {
	capabilities, err := support.Profile.Capabilities()
	if err != nil || !capabilities.Record {
		return registry, err
	}
	result := registry.Clone()
	for _, node := range nodes {
		if node.Kind != EntityKind || len(node.Fields) == 0 {
			continue
		}
		if err := result.Register(semantic.TypeDef{ID: semantic.ID(node.ID), Namespace: semantic.Namespace(namespace), Name: node.Name}); err != nil {
			return registry, entityFieldsError(EntityFieldsUnsupportedTypeDiagnostic, fmt.Sprintf("record %q: %v", node.ID, err), node.Span, err)
		}
	}
	return result, nil
}

func validateRecordReferenceGraph(nodes []Node, registry semantic.TypeRegistry, support EntityFieldsSupport) error {
	capabilities, err := support.Profile.Capabilities()
	if err != nil || !capabilities.Record {
		return err
	}
	var records []semantic.Node
	fields := make(map[semantic.ID]Field)
	for _, node := range nodes {
		if node.Kind != EntityKind || len(node.Fields) == 0 {
			continue
		}
		record := semantic.Node{ID: semantic.ID(node.ID), Kind: semantic.Entity}
		for _, field := range node.Fields {
			normalized, err := normalizeField(field, node.ID, EntityKind, registry)
			if err != nil {
				return err
			}
			id := semantic.ID(field.ID)
			fields[id] = field
			record.Fields = append(record.Fields, semantic.Field{ID: id, TypeRef: semantic.TypeRef{ID: semantic.ID(normalized.TypeRefUse.ResolvedID)}})
		}
		records = append(records, record)
	}
	err = semantic.ValidateRecordReferences(records, syntax.RecordReferenceDepthLimit)
	if issues, ok := err.(*semantic.ValidationErrors); ok && len(issues.Issues) != 0 {
		issue := issues.Issues[0]
		if field, exists := fields[issue.Subject]; exists {
			return recordReferenceError(field, issue.Message)
		}
	}
	return err
}

func recordReferenceError(field Field, message string) error {
	span := field.TypeRefSpan
	if !span.Valid() {
		span = field.Span
	}
	return entityFieldsError(EntityFieldsUnsupportedShapeDiagnostic, fmt.Sprintf("field %q: %s", field.ID, message), span, ErrInvalidField)
}
