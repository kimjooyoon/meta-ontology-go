package semantic

import (
	"fmt"
	"sort"
)

// Graph stores normalized semantic nodes and two explicitly separate fact
// sets. The zero value is ready for use.
type Graph struct {
	nodes      map[ID]Node
	names      map[NameRef]ID
	facts      map[FactKey]Fact
	candidates map[FactKey]Fact
}

func NewGraph() Graph {
	return Graph{
		nodes:      make(map[ID]Node),
		names:      make(map[NameRef]ID),
		facts:      make(map[FactKey]Fact),
		candidates: make(map[FactKey]Fact),
	}
}

func (g *Graph) ensure() {
	if g.nodes == nil {
		g.nodes = make(map[ID]Node)
	}
	if g.names == nil {
		g.names = make(map[NameRef]ID)
	}
	if g.facts == nil {
		g.facts = make(map[FactKey]Fact)
	}
	if g.candidates == nil {
		g.candidates = make(map[FactKey]Fact)
	}
}

func (g *Graph) AddNode(node Node) error {
	normalized, err := node.Normalized()
	if err != nil {
		return err
	}

	old, exists := g.nodes[normalized.ID]
	if exists && (old.Kind != normalized.Kind || old.Namespace != normalized.Namespace) {
		return fmt.Errorf("%w: %s cannot change kind or namespace", ErrIdentityConflict, normalized.ID)
	}
	if err := g.validateEntityFieldsIdentityDomain(normalized); err != nil {
		return err
	}
	for _, ref := range nodeNameRefs(normalized) {
		if owner, occupied := g.names[ref]; occupied && owner != normalized.ID {
			return fmt.Errorf("%w: %s/%s is already owned by %s", ErrNameCollision, ref.Namespace, ref.Name, owner)
		}
	}
	g.ensure()
	if exists {
		for _, ref := range nodeNameRefs(old) {
			if owner, occupied := g.names[ref]; occupied && owner == normalized.ID {
				delete(g.names, ref)
			}
		}
	}
	g.nodes[normalized.ID] = normalized
	for _, ref := range nodeNameRefs(normalized) {
		g.names[ref] = normalized.ID
	}
	return nil
}

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

// NodeByName performs a namespace-qualified lookup. An unqualified name can
// never resolve across namespace boundaries.
func (g Graph) NodeByName(namespace Namespace, name string) (Node, bool) {
	ref, err := NewNameRef(namespace, name)
	if err != nil {
		return Node{}, false
	}
	id, ok := g.names[ref]
	if !ok {
		return Node{}, false
	}
	return g.Node(id)
}

func (g Graph) ResolveName(namespace, name string) (Node, error) {
	ns, err := ParseNamespace(namespace)
	if err != nil {
		return Node{}, err
	}
	node, ok := g.NodeByName(ns, name)
	if !ok {
		return Node{}, fmt.Errorf("%w: %s/%s", ErrNodeNotFound, ns, name)
	}
	return node, nil
}

func copyNode(node Node) Node {
	node.Aliases = append([]string(nil), node.Aliases...)
	node.Fields = copyFields(node.Fields)
	return node
}
