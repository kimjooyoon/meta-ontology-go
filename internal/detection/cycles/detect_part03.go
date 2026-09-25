package cycles

import (
	"fmt"
	"sort"
)

func (t *nodeTable) addNode(node Node) {
	if old, exists := t.nodes[node.ID]; exists {
		if !sameDeclaration(old, node) {
			t.diagnostics = append(t.diagnostics, Diagnostic{
				Code: NamespaceCollision, NodeID: node.ID, Namespace: node.Namespace,
				Name: node.Name, Span: node.Span,
				Message: fmt.Sprintf("stable ID %q has conflicting declarations in namespace %q", node.ID, node.Namespace),
			})
		}
		return
	}
	t.nodes[node.ID] = node
}
func sameDeclaration(left, right Node) bool {
	if left.ID != right.ID || left.Kind != right.Kind || left.Namespace != right.Namespace || left.Name != right.Name {
		return false
	}
	if len(left.Aliases) != len(right.Aliases) {
		return false
	}
	for i := range left.Aliases {
		if left.Aliases[i] != right.Aliases[i] {
			return false
		}
	}
	return true
}
func nodeNames(node Node) []string {
	result := make([]string, 0, 1+len(node.Aliases))
	if node.Name != "" {
		result = append(result, node.Name)
	}
	for _, alias := range node.Aliases {
		if alias != "" && alias != node.Name {
			result = append(result, alias)
		}
	}
	sort.Strings(result)
	return result
}
func addNameOwner(table *nodeTable, owners map[string]ID, node Node, name string) {
	key := node.Namespace + "\x00" + name
	if owner, exists := owners[key]; exists {
		if owner != node.ID {
			table.diagnostics = append(table.diagnostics, Diagnostic{
				Code: NamespaceCollision, NodeID: node.ID, Namespace: node.Namespace,
				Name: name, Span: node.Span,
				Message: fmt.Sprintf("name %q in namespace %q is owned by both %q and %q", name, node.Namespace, owner, node.ID),
			})
		}
		return
	}
	owners[key] = node.ID
}
