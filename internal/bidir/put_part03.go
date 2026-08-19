package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateFieldSourceSnapshot(before, after Model) error {
	source := make(map[ID]Field)
	for _, node := range before.Nodes {
		for _, field := range node.Fields {
			source[field.ID] = field
		}
	}
	for _, node := range after.Nodes {
		for _, field := range node.Fields {
			previous, exists := source[field.ID]
			if exists && previous.Span.File != field.Span.File {
				return entityFieldsError(EntityFieldsIncompleteDiagnostic, fmt.Sprintf("field %q provenance crosses source snapshots", field.ID), field.Span, ErrUnrepresentableField)
			}
		}
	}
	return nil
}
func putError(code string, err error) error {
	return &PutError{Code: code, NoWrite: true, Err: err}
}
func validatePutProvenance(source, updated Model) error {
	delta := Diff(source, updated)
	for _, node := range append(delta.AddedNodes, delta.RemovedNodes...) {
		if !node.Span.Valid() && len(node.Fields) == 0 {
			return fmt.Errorf("node %q semantic change has no source span", node.ID)
		}
		for _, field := range node.Fields {
			if !field.Span.Valid() {
				return fmt.Errorf("field %q semantic change has no source span", field.ID)
			}
		}
	}
	for _, relation := range append(delta.AddedRelations, delta.RemovedRelations...) {
		if !relation.Span.Valid() {
			return fmt.Errorf("relation %s %q -> %q semantic change has no source span", relation.Kind, relation.Source, relation.Target)
		}
	}
	return nil
}
func appendSurvivingDeclarations(result *Document, declarations []Declaration, nodes map[ID]Node, updated Model, registry semantic.TypeRegistry) (map[ID]struct{}, error) {
	ids := make(map[ID]struct{}, len(declarations))
	for _, declaration := range declarations {
		id, err := declarationIdentity(result.Namespace, declaration)
		if err != nil {
			return nil, err
		}
		if _, exists := ids[id]; exists {
			return nil, fmt.Errorf("duplicate source declaration ID %q", id)
		}
		ids[id] = struct{}{}
		if node, exists := nodes[id]; exists {
			declaration, err := declarationFromNode(node, updated, registry)
			if err != nil {
				return nil, err
			}
			result.Declarations = append(result.Declarations, declaration)
		}
	}
	return ids, nil
}
