package cycles

import (
	"fmt"
	"sort"
	"strings"
)

type nodeTable struct {
	nodes       map[ID]Node
	orderedIDs  []ID
	diagnostics Diagnostics
}

type normalizedEdge struct {
	subject   ID
	predicate Relation
	object    ID
	span      Span
	known     bool
}

// Detect reports every supported structural problem in graph. Diagnostics
// are sorted by a stable key, so equivalent graphs produce equivalent output
// regardless of declaration or relation insertion order.
func Detect(graph Graph) Diagnostics {
	table := indexNodes(graph.Nodes)
	edges, edgeDiagnostics := indexEdges(graph.edges(), table.nodes)
	result := append(table.diagnostics, edgeDiagnostics...)
	result = append(result, detectCycles(table.nodes, edges)...)
	sortDiagnostics(result)
	return result
}

// Analyze is a descriptive alias for Detect.
func Analyze(graph Graph) Diagnostics {
	return Detect(graph)
}

// DetectCycles is a descriptive alias for Detect. It retains the package's
// historical name while still returning all graph diagnostics.
func DetectCycles(graph Graph) Diagnostics {
	return Detect(graph)
}

// Validate is a descriptive alias for Detect.
func Validate(graph Graph) Diagnostics {
	return Detect(graph)
}

// Diagnostics returns the diagnostics for graph.
func (g Graph) Diagnostics() Diagnostics {
	return Detect(g)
}

// Validate returns the diagnostics for graph.
func (g Graph) Validate() Diagnostics {
	return Detect(g)
}

// Check returns nil for a valid graph, or its deterministic diagnostics as an
// error when one or more invariants fail.
func Check(graph Graph) error {
	diagnostics := Detect(graph)
	if len(diagnostics) == 0 {
		return nil
	}
	return diagnostics
}

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
