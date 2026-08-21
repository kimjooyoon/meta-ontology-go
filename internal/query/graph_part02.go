package query

import (
	"fmt"
)

// AddNode registers a typed endpoint in the derived view. It never changes the
// SemanticIR source; FromSemanticIR uses it while constructing a detached graph.
func (graph *Graph) AddNode(node Node) error {
	normalized, err := node.normalized()
	if err != nil {
		return err
	}
	graph.ensure()
	if existing, ok := graph.nodes[normalized.ID]; ok && existing.Kind != UnknownNodeKind &&
		normalized.Kind != UnknownNodeKind && existing.Kind != normalized.Kind {
		return fmt.Errorf("%w: %s is both %s and %s", ErrInvalidNode, normalized.ID, existing.Kind, normalized.Kind)
	}
	if existing, ok := graph.nodes[normalized.ID]; ok && existing.Namespace != "" &&
		normalized.Namespace != "" && existing.Namespace != normalized.Namespace {
		return fmt.Errorf("%w: %s changes namespace from %s to %s", ErrInvalidNode, normalized.ID, existing.Namespace, normalized.Namespace)
	}
	if len(displayNames(normalized)) > 0 && normalized.Namespace != "" {
		for _, existing := range graph.nodes {
			if existing.ID == normalized.ID || existing.Namespace != normalized.Namespace {
				continue
			}
			for _, incomingName := range displayNames(normalized) {
				if existing.hasName(incomingName) {
					return fmt.Errorf("%w: %s/%s is already attached to %s", ErrInvalidNode, normalized.Namespace, incomingName, existing.ID)
				}
			}
		}
	}
	if existing, ok := graph.nodes[normalized.ID]; ok && existing.Kind != UnknownNodeKind {
		return nil
	}
	graph.nodes[normalized.ID] = copyQueryNode(normalized)
	return nil
}

// Node returns a detached endpoint snapshot.
func (graph Graph) Node(id ID) (Node, bool) {
	canonical, err := ParseID(id.String())
	if err != nil {
		return Node{}, false
	}
	node, ok := graph.nodes[canonical]
	if ok {
		node = copyQueryNode(node)
	}
	return node, ok
}

// Nodes returns endpoints in stable ID/kind order.
func (graph Graph) Nodes() []Node {
	nodes := make([]Node, 0, len(graph.nodes))
	for _, node := range graph.nodes {
		nodes = append(nodes, node)
	}
	sortNodes(nodes)
	return nodes
}
