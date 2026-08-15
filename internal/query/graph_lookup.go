package query

// NodeByName performs a namespace-qualified display-name lookup. An empty
// namespace is rejected so equal names from separate contexts cannot merge.
func (graph Graph) NodeByName(namespace, name string) (Node, bool) {
	namespace = normalizeDisplayName(namespace)
	name = normalizeDisplayName(name)
	if namespace == "" || name == "" {
		return Node{}, false
	}
	var found Node
	for _, node := range graph.nodes {
		if node.Namespace != namespace || !node.hasName(name) {
			continue
		}
		if found.ID != "" && found.ID != node.ID {
			// A malformed or manually assembled graph must fail closed rather
			// than choosing a map-order-dependent result.
			return Node{}, false
		}
		found = copyQueryNode(node)
	}
	return found, found.ID != ""
}

// Search returns namespace-qualified display-name matches in stable order.
// It is a lookup projection only; facts remain keyed by stable IDs.
func (graph Graph) Search(namespace, name string) []Node {
	namespace = normalizeDisplayName(namespace)
	name = normalizeDisplayName(name)
	if namespace == "" || name == "" {
		return nil
	}
	results := make([]Node, 0)
	for _, node := range graph.nodes {
		if node.Namespace == namespace && node.hasName(name) {
			results = append(results, copyQueryNode(node))
		}
	}
	sortNodes(results)
	return results
}

func copyQueryNode(node Node) Node {
	node.Aliases = append([]string(nil), node.Aliases...)
	return node
}
