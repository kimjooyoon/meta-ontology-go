package cycles

import (
	"fmt"
	"sort"
	"strings"
)

func indexNodes(rawNodes []Node) nodeTable {
	records := make([]Node, 0, len(rawNodes))
	diagnostics := make(Diagnostics, 0)
	for _, raw := range rawNodes {
		node, err := normalizeNode(raw)
		if err != nil {
			diagnostics = append(diagnostics, invalidNodeDiagnostic(raw, err))
			continue
		}
		records = append(records, node)
	}
	sortNodes(records)
	table := nodeTable{nodes: make(map[ID]Node), diagnostics: diagnostics}
	nameOwners := make(map[string]ID)
	for _, node := range records {
		table.addNode(node)
		for _, name := range nodeNames(node) {
			addNameOwner(&table, nameOwners, node, name)
		}
	}
	table.orderedIDs = make([]ID, 0, len(table.nodes))
	for id := range table.nodes {
		table.orderedIDs = append(table.orderedIDs, id)
	}
	sort.Strings(table.orderedIDs)
	return table
}
func normalizeNode(raw Node) (Node, error) {
	id, err := canonicalID(raw.ID)
	if err != nil {
		return Node{}, err
	}
	raw.ID = id
	raw.Namespace = normalizedNamespace(raw.Namespace)
	raw.Name = normalizedName(raw.Name)
	raw.Aliases = append([]string(nil), raw.Aliases...)
	for i, alias := range raw.Aliases {
		raw.Aliases[i] = normalizedName(alias)
	}
	sort.Strings(raw.Aliases)
	return raw, nil
}
func invalidNodeDiagnostic(node Node, err error) Diagnostic {
	return Diagnostic{
		Code: InvalidStableID, NodeID: strings.TrimSpace(node.ID),
		Namespace: normalizedNamespace(node.Namespace), Name: normalizedName(node.Name),
		Span:    node.Span,
		Message: fmt.Sprintf("node stable ID %q is invalid: %v", node.ID, err),
	}
}
func sortNodes(nodes []Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		left, right := nodes[i], nodes[j]
		return strings.Join([]string{left.ID, left.Namespace, left.Name, string(left.Kind),
			strings.Join(left.Aliases, "\x00")}, "\x00") <
			strings.Join([]string{right.ID, right.Namespace, right.Name, string(right.Kind),
				strings.Join(right.Aliases, "\x00")}, "\x00")
	})
}
