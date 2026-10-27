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
	records := make(map[ID]Node)
	for _, node := range nodes {
		if node.Kind == EntityKind && len(node.Fields) != 0 {
			records[node.ID] = node
		}
	}
	heights, visiting := make(map[ID]int), make(map[ID]bool)
	var visit func(ID, int) (int, error)
	visit = func(id ID, depth int) (int, error) {
		if h := heights[id]; h != 0 {
			return h, nil
		}
		visiting[id] = true
		height := 1
		for _, field := range records[id].Fields {
			normalized, err := normalizeField(field, id, EntityKind, registry)
			if err != nil {
				return 0, err
			}
			target := normalized.TypeRefUse.ResolvedID
			if target == ID(semantic.BuiltinStringTypeID) || target == ID(semantic.BuiltinBooleanTypeID) || target == ID(semantic.BuiltinIntegerTypeID) {
				continue
			}
			message := ""
			if _, ok := records[target]; !ok {
				message = "record field requires a declared record type"
			}
			if visiting[target] {
				message = "cyclic record reference"
			}
			if message != "" {
				return 0, recordReferenceError(field, message)
			}
			if depth >= syntax.RecordReferenceDepthLimit {
				return 0, recordReferenceError(field, fmt.Sprintf("record depth exceeds %d", syntax.RecordReferenceDepthLimit))
			}
			childHeight, err := visit(target, depth+1)
			if err != nil {
				return 0, err
			}
			if depth+childHeight > syntax.RecordReferenceDepthLimit {
				return 0, recordReferenceError(field, fmt.Sprintf("record depth exceeds %d", syntax.RecordReferenceDepthLimit))
			}
			if childHeight+1 > height {
				height = childHeight + 1
			}
		}
		visiting[id] = false
		heights[id] = height
		return height, nil
	}
	for _, node := range nodes {
		if _, ok := records[node.ID]; ok {
			if _, err := visit(node.ID, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

func recordReferenceError(field Field, message string) error {
	span := field.TypeRefSpan
	if !span.Valid() {
		span = field.Span
	}
	return entityFieldsError(EntityFieldsUnsupportedShapeDiagnostic, fmt.Sprintf("field %q: %s", field.ID, message), span, ErrInvalidField)
}
