package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateModelFields(nodes []Node, registry semantic.TypeRegistry) error {
	fieldOwners := make(map[ID]ID)
	nodeIDs := make(map[ID]Kind, len(nodes))
	for _, node := range nodes {
		nodeIDs[node.ID] = node.Kind
	}
	for _, node := range nodes {
		if len(node.Fields) == 0 {
			continue
		}
		if node.Kind != EntityKind {
			return fmt.Errorf("%w: fields are only valid on Entity nodes", ErrInvalidField)
		}
		nameOwners := make(map[string]ID, len(node.Fields)*2)
		for index, field := range node.Fields {
			normalized, err := normalizeField(field, node.ID, node.Kind, registry)
			if err != nil {
				return fmt.Errorf("node %q field %d: %w", node.ID, index, err)
			}
			if owner, exists := fieldOwners[normalized.ID]; exists {
				return entityFieldsError(EntityFieldsIDCollisionDiagnostic, fmt.Sprintf("duplicate field ID %s owned by %s and %s", normalized.ID, owner, node.ID), field.Span, ErrInvalidField)
			}
			if kind, exists := nodeIDs[normalized.ID]; exists {
				return entityFieldsError(EntityFieldsIDCollisionDiagnostic, fmt.Sprintf("field ID %s collides with %s ID", normalized.ID, kind), field.Span, ErrInvalidField)
			}
			fieldOwners[normalized.ID] = node.ID
			for _, name := range append([]string{normalized.Name}, normalized.Aliases...) {
				if owner, exists := nameOwners[name]; exists && owner != normalized.ID {
					return fmt.Errorf("%w: field name %q is shared by %s and %s", ErrInvalidField, name, owner, normalized.ID)
				}
				nameOwners[name] = normalized.ID
			}
		}
	}
	return nil
}
func validateFieldOrderStability(before, after Model) error {
	beforeNodes, afterNodes := nodeMap(before.Nodes), nodeMap(after.Nodes)
	for id, source := range beforeNodes {
		updated, exists := afterNodes[id]
		if !exists || len(source.Fields) < 2 {
			continue
		}
		positions := make(map[ID]int, len(updated.Fields))
		for index, field := range updated.Fields {
			positions[field.ID] = index
		}
		last := -1
		for _, field := range source.Fields {
			position, exists := positions[field.ID]
			if !exists {
				continue
			}
			if position < last {
				return entityFieldsError(EntityFieldsIllegalReorderDiagnostic, fmt.Sprintf("field order for entity %s changed", id), field.Span, ErrInvalidField)
			}
			last = position
		}
	}
	return nil
}
