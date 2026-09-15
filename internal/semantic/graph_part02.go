package semantic

import (
	"fmt"
	"sort"
)

func (g Graph) validateEntityFieldsIdentityDomain(node Node) error {
	for existingID, existing := range g.nodes {
		for _, field := range existing.Fields {
			if field.ID == node.ID {
				return fmt.Errorf("%w: node %s collides with field %s on %s", ErrInvalidField, node.ID, field.ID, existingID)
			}
			if existingID == node.ID {
				continue
			}
			for _, incoming := range node.Fields {
				if field.ID == incoming.ID {
					return fmt.Errorf("%w: field %s cannot move from %s to %s", ErrInvalidField, incoming.ID, existingID, node.ID)
				}
			}
		}
	}
	for _, field := range node.Fields {
		if field.ID == node.ID {
			return fmt.Errorf("%w: field %s collides with parent node", ErrInvalidField, field.ID)
		}
		if _, exists := g.nodes[field.ID]; exists {
			return fmt.Errorf("%w: field %s collides with declaration %s", ErrInvalidField, field.ID, field.ID)
		}
	}
	return nil
}
func nodeNameRefs(node Node) []NameRef {
	refs := make([]NameRef, 0, 1+len(node.Aliases))
	refs = append(refs, NameRef{Namespace: node.Namespace, Name: node.Name})
	for _, alias := range node.Aliases {
		refs = append(refs, NameRef{Namespace: node.Namespace, Name: alias})
	}
	return refs
}
func (g Graph) Node(id ID) (Node, bool) {
	canonical, err := ParseIdentity(id.String())
	if err != nil {
		return Node{}, false
	}
	node, ok := g.nodes[canonical]
	if !ok {
		return Node{}, false
	}
	return copyNode(node), true
}
func (g Graph) Nodes() []Node {
	nodes := make([]Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, copyNode(node))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes
}

// SortedNodes is an explicit alias for adapters that prefer sorted snapshots.
func (g Graph) SortedNodes() []Node {
	return g.Nodes()
}
