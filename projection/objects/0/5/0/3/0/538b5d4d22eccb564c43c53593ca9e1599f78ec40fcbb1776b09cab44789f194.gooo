package query

import (
	"slices"
	"sort"
)

func (node Node) hasName(name string) bool {
	name = normalizeDisplayName(name)
	if name == "" || node.Name == name {
		return node.Name == name && name != ""
	}
	return slices.Contains(node.Aliases, name)
}
func displayNames(node Node) []string {
	names := make([]string, 0, 1+len(node.Aliases))
	if node.Name != "" {
		names = append(names, node.Name)
	}
	names = append(names, node.Aliases...)
	return names
}
func sortNodes(nodes []Node) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Namespace != nodes[j].Namespace {
			return nodes[i].Namespace < nodes[j].Namespace
		}
		if nodes[i].ID != nodes[j].ID {
			return nodes[i].ID < nodes[j].ID
		}
		if nodes[i].Kind != nodes[j].Kind {
			return nodes[i].Kind < nodes[j].Kind
		}
		return nodes[i].Name < nodes[j].Name
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
