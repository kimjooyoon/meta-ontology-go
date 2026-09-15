package query

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
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
// identity; Namespace and the display fields are lookup metadata copied from
// SemanticIR. They never replace ID when facts are keyed or matched.
type Node struct {
	ID        ID
	Kind      NodeKind
	Namespace string
	Name      string
	Aliases   []string
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
	if node.Namespace != "" {
		namespace := strings.TrimSpace(node.Namespace)
		if namespace == "" || strings.IndexFunc(namespace, unicode.IsSpace) >= 0 {
			return Node{}, fmt.Errorf("%w: namespace must not contain whitespace", ErrInvalidNode)
		}
		node.Namespace = namespace
	}
	if node.Name != "" {
		node.Name = normalizeDisplayName(node.Name)
	}
	aliases := make([]string, 0, len(node.Aliases))
	seen := make(map[string]struct{}, len(node.Aliases))
	for _, raw := range node.Aliases {
		alias := normalizeDisplayName(raw)
		if alias == "" || alias == node.Name {
			continue
		}
		if _, exists := seen[alias]; exists {
			continue
		}
		seen[alias] = struct{}{}
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	node.Aliases = aliases
	node.ID = id
	return node, nil
}
func normalizeDisplayName(raw string) string {
	return strings.Join(strings.Fields(raw), " ")
}
