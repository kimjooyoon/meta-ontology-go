package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"unicode"
)

// ValidateWithTypes validates field TypeRefs against an explicit registry.
// The receiver and all nested slices remain untouched.
func (m Model) ValidateWithTypes(registry semantic.TypeRegistry) error {
	seenNodes := make(map[ID]Kind, len(m.Nodes))
	for _, node := range m.Nodes {
		if err := node.Span.Validate(); err != nil {
			return fmt.Errorf("node %q: %w", node.ID, err)
		}
		if err := validateID(node.ID); err != nil {
			return fmt.Errorf("node %q: %w", node.ID, err)
		}
		if node.Kind == "" {
			return fmt.Errorf("node %q has empty kind", node.ID)
		}
		if previous, exists := seenNodes[node.ID]; exists {
			return fmt.Errorf("duplicate node ID %q (%s and %s)", node.ID, previous, node.Kind)
		}
		seenNodes[node.ID] = node.Kind
	}
	if err := validateModelFields(m.Nodes, registry); err != nil {
		return err
	}
	seenRelations := make(map[string]struct{}, len(m.Relations))
	for _, relation := range m.Relations {
		if err := relation.Span.Validate(); err != nil {
			return fmt.Errorf("relation %s %q -> %q: %w", relation.Kind, relation.Source, relation.Target, err)
		}
		if relation.Kind == "" {
			return fmt.Errorf("relation %q -> %q has empty predicate", relation.Source, relation.Target)
		}
		if _, exists := seenNodes[relation.Source]; !exists {
			return fmt.Errorf("relation %s references unknown source %q", relation.Kind, relation.Source)
		}
		if _, exists := seenNodes[relation.Target]; !exists {
			return fmt.Errorf("relation %s references unknown target %q", relation.Kind, relation.Target)
		}
		key := relationKey(relation.Kind, relation.Source, relation.Target)
		if _, exists := seenRelations[key]; exists {
			return fmt.Errorf("duplicate relation %s", key)
		}
		seenRelations[key] = struct{}{}
	}
	for _, candidate := range m.Candidates {
		if candidate.Layer != CandidateFact {
			return fmt.Errorf("model candidate %q has non-candidate layer %s", candidate.SemanticKey(), candidate.Layer)
		}
		if _, exists := seenRelations[candidate.SemanticKey()]; exists {
			return fmt.Errorf("candidate %q is shadowed by a deterministic relation", candidate.SemanticKey())
		}
	}
	return nil
}
func validateID(id ID) error {
	if strings.TrimSpace(string(id)) == "" {
		return fmt.Errorf("empty semantic ID")
	}
	for _, r := range string(id) {
		if unicode.IsSpace(r) {
			return fmt.Errorf("semantic ID contains whitespace")
		}
	}
	return nil
}
