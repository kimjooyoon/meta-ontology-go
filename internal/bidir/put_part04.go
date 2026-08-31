package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"slices"
	"sort"
)

func appendNewDeclarations(result *Document, updated Model, nodes map[ID]Node, existing map[ID]struct{}, registry semantic.TypeRegistry) error {
	implicitIDs := implicitActivityPortIDs(result)
	ids := make([]ID, 0, len(updated.Nodes))
	for _, node := range updated.Nodes {
		if _, exists := existing[node.ID]; !exists {
			if _, implicit := implicitIDs[node.ID]; implicit {
				continue
			}
			ids = append(ids, node.ID)
		}
	}
	slices.Sort(ids)
	for _, id := range ids {
		declaration, err := declarationFromNode(nodes[id], updated, registry)
		if err != nil {
			return err
		}
		result.Declarations = append(result.Declarations, declaration)
	}
	return nil
}

func implicitActivityPortIDs(document *Document) map[ID]struct{} {
	result := make(map[ID]struct{})
	if document == nil || !document.ImplicitActivityPorts {
		return result
	}
	for _, declaration := range implicitEntityDeclarations(document.Declarations, document.Namespace) {
		id, err := declarationIdentity(document.Namespace, declaration)
		if err == nil {
			result[id] = struct{}{}
		}
	}
	return result
}
func appendUpdatedRelations(result *Document, original []Relation, updated Model) {
	implicit := implicitRelationKeys(result.Declarations, result.Namespace)
	relations := make(map[string]Relation, len(updated.Relations))
	for _, relation := range updated.Relations {
		relation = relation.normalized()
		key := relationKey(relation.Kind, relation.Source, relation.Target)
		if _, representable := implicit[key]; representable && len(relation.Attributes) == 0 {
			continue
		}
		relations[key] = relation
	}
	seen := make(map[string]struct{}, len(relations))
	for _, relation := range original {
		key := relationKey(relation.Kind, relation.Source, relation.Target)
		if next, exists := relations[key]; exists {
			result.Relations = append(result.Relations, next)
			seen[key] = struct{}{}
		}
	}
	remaining := make([]Relation, 0, len(relations))
	for key, relation := range relations {
		if _, exists := seen[key]; !exists {
			remaining = append(remaining, relation)
		}
	}
	sort.Slice(remaining, func(i, j int) bool { return relationLess(remaining[i], remaining[j]) })
	result.Relations = append(result.Relations, remaining...)
}

// CheckGetPut verifies that a generic source view survives a write-back.
func CheckGetPut(document Document) error {
	model, err := Get(document)
	if err != nil {
		return err
	}
	written, err := Put(document, model)
	if err != nil {
		return err
	}
	observed, err := Get(written)
	if err != nil {
		return err
	}
	if !SemanticEquivalent(model, observed) {
		return fmt.Errorf("Get-Put violated: semantic model changed after write-back")
	}
	return nil
}
