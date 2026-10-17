package semantic

import (
	"fmt"
)

// Graph stores normalized semantic nodes and two explicitly separate fact
// sets. The zero value is ready for use.
type Graph struct {
	nodes       map[ID]Node
	names       map[NameRef]ID
	fieldOwners map[ID]ID
	facts       map[FactKey]Fact
	candidates  map[FactKey]Fact
}

func NewGraph() Graph {
	return Graph{
		nodes:       make(map[ID]Node),
		names:       make(map[NameRef]ID),
		fieldOwners: make(map[ID]ID),
		facts:       make(map[FactKey]Fact),
		candidates:  make(map[FactKey]Fact),
	}
}
func (g *Graph) ensure() {
	if g.nodes == nil {
		g.nodes = make(map[ID]Node)
	}
	if g.names == nil {
		g.names = make(map[NameRef]ID)
	}
	if g.fieldOwners == nil {
		g.fieldOwners = make(map[ID]ID)
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
		for _, field := range old.Fields {
			delete(g.fieldOwners, field.ID)
		}
		for _, ref := range nodeNameRefs(old) {
			if owner, occupied := g.names[ref]; occupied && owner == normalized.ID {
				delete(g.names, ref)
			}
		}
	}
	g.nodes[normalized.ID] = normalized
	for _, field := range normalized.Fields {
		g.fieldOwners[field.ID] = normalized.ID
	}
	for _, ref := range nodeNameRefs(normalized) {
		g.names[ref] = normalized.ID
	}
	return nil
}
