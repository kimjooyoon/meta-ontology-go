package semantic

import (
	"fmt"
)

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
