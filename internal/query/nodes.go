package query

import (
	"fmt"
	"sort"
)

// NodeKind is the PROV role carried by a query projection endpoint.
type NodeKind string

const (
	UnknownNodeKind  NodeKind = "Unknown"
	EntityNodeKind   NodeKind = "Entity"
	ActivityNodeKind NodeKind = "Activity"
	AgentNodeKind    NodeKind = "Agent"
)

// Node is a detached typed endpoint in the derived query view. Stable ID is
// identity; Kind is a validation hint copied from SemanticIR.
type Node struct {
	ID   ID
	Kind NodeKind
}

func (node Node) normalized() (Node, error) {
	id, err := ParseID(node.ID.String())
	if err != nil {
		return Node{}, fmt.Errorf("%w: %v", ErrInvalidNode, err)
	}
	if node.Kind == "" {
		node.Kind = UnknownNodeKind
	}
	if node.Kind != UnknownNodeKind && node.Kind != EntityNodeKind &&
		node.Kind != ActivityNodeKind && node.Kind != AgentNodeKind {
		return Node{}, fmt.Errorf("%w: unknown kind %q", ErrInvalidNode, node.Kind)
	}
	node.ID = id
	return node, nil
}

func sortNodes(nodes []Node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].ID != nodes[j].ID {
			return nodes[i].ID < nodes[j].ID
		}
		return nodes[i].Kind < nodes[j].Kind
	})
}

func relationNodeKinds(relation Relation) (NodeKind, NodeKind, bool) {
	switch relation {
	case Used:
		return ActivityNodeKind, EntityNodeKind, true
	case WasGeneratedBy:
		return EntityNodeKind, ActivityNodeKind, true
	case WasDerivedFrom:
		return EntityNodeKind, EntityNodeKind, true
	case WasAssociatedWith:
		return ActivityNodeKind, AgentNodeKind, true
	default:
		return "", "", false
	}
}
