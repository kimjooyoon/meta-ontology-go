package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type r4Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	PathID string `json:"path_id"`
}

type r4Component struct {
	Digest  string
	Members []string
	Edges   []r4Edge
	Cyclic  bool
}

type r4Graph struct {
	nodes              []string
	edges              []r4Edge
	reachableNodes     []string
	reachableEdges     []r4Edge
	reachablePaths     []RepairPath
	components         []r4Component
	condensationEdges  [][2]string
	graphDigest        string
	sccDigest          string
	condensationDigest string
	ruleDigest         string
	receipt            R4WorkReceipt
}

type r4GraphPayload struct {
	SchemaVersion string   `json:"schema_version"`
	Roots         []string `json:"roots"`
	Nodes         []string `json:"nodes"`
	Edges         []r4Edge `json:"edges"`
}

type r4SCCPayload struct {
	Members []string `json:"members"`
	Edges   []r4Edge `json:"edges"`
}

type r4SCCDigestPayload struct {
	SchemaVersion string   `json:"schema_version"`
	Components    []string `json:"components"`
}

type r4CondensationPayload struct {
	SchemaVersion string      `json:"schema_version"`
	Components    []string    `json:"components"`
	Edges         [][2]string `json:"edges"`
}

type r4RuleDigestPayload struct {
	SchemaVersion string   `json:"schema_version"`
	Rules         []R4Rule `json:"rules"`
}

func buildR4Graph(input R4Input) (r4Graph, string) {
	graph := r4Graph{}
	pathByID := make(map[string]RepairPath, len(input.Paths))
	nodeSet := make(map[string]struct{}, len(input.Paths))
	edges := make([]r4Edge, 0, len(input.Paths))
	for _, path := range input.Paths {
		pathID := path.StableID
		if pathID == "" || path.ObligationID == "" {
			return graph, R4ReasonMalformedGraph
		}
		if _, exists := pathByID[pathID]; exists {
			return graph, R4ReasonMalformedGraph
		}
		pathByID[pathID] = path
		nodeSet[path.ObligationID] = struct{}{}
		seenPrerequisites := make(map[string]struct{}, len(path.PrerequisiteObligationIDs))
		for _, prerequisite := range path.PrerequisiteObligationIDs {
			if prerequisite == "" {
				return graph, R4ReasonMalformedGraph
			}
			if _, duplicate := seenPrerequisites[prerequisite]; duplicate {
				return graph, R4ReasonMalformedGraph
			}
			seenPrerequisites[prerequisite] = struct{}{}
			nodeSet[prerequisite] = struct{}{}
			edges = append(edges, r4Edge{From: prerequisite, To: path.ObligationID, PathID: pathID})
		}
	}
	graph.nodes = sortedKeys(nodeSet)
	sort.Slice(edges, func(i, j int) bool { return edgeKey(edges[i]) < edgeKey(edges[j]) })
	graph.edges = edges
	if len(input.RootObligationIDs) == 0 {
		return graph, R4ReasonRequiredInputMissing
	}
	if len(sortedUnique(input.RootObligationIDs)) != len(input.RootObligationIDs) {
		return graph, R4ReasonMalformedGraph
	}
	roots := sortedCopy(input.RootObligationIDs)
	for _, root := range roots {
		if _, ok := nodeSet[root]; !ok || root == "" {
			return graph, R4ReasonMalformedGraph
		}
	}

	adjacency := make(map[string][]string, len(graph.nodes))
	for _, node := range graph.nodes {
		adjacency[node] = nil
	}
	for _, edge := range graph.edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	for node := range adjacency {
		adjacency[node] = sortedUnique(adjacency[node])
	}
	reachable := reachableFrom(roots, adjacency)
	graph.reachableNodes = sortedKeys(reachable)
	for _, edge := range graph.edges {
		if _, fromOK := reachable[edge.From]; fromOK {
			if _, toOK := reachable[edge.To]; toOK {
				graph.reachableEdges = append(graph.reachableEdges, edge)
			}
		}
	}
	for _, path := range input.Paths {
		if _, ok := reachable[path.ObligationID]; ok {
			graph.reachablePaths = append(graph.reachablePaths, path)
		}
	}
	sort.Slice(graph.reachablePaths, func(i, j int) bool {
		return r4PathKey(graph.reachablePaths[i]) < r4PathKey(graph.reachablePaths[j])
	})

	graph.components = deriveR4Components(graph.reachableNodes, graph.reachableEdges)
	graph.condensationEdges = deriveR4Condensation(graph.components, graph.reachableEdges)
	graph.graphDigest = digestR4(r4GraphPayload{
		SchemaVersion: R4SchemaVersion,
		Roots:         roots,
		Nodes:         graph.reachableNodes,
		Edges:         graph.reachableEdges,
	})
	componentDigests := make([]string, 0, len(graph.components))
	for _, component := range graph.components {
		componentDigests = append(componentDigests, component.Digest)
	}
	sort.Strings(componentDigests)
	graph.sccDigest = digestR4(r4SCCDigestPayload{SchemaVersion: R4SchemaVersion, Components: componentDigests})
	graph.condensationDigest = digestR4(r4CondensationPayload{
		SchemaVersion: R4SchemaVersion,
		Components:    componentDigests,
		Edges:         graph.condensationEdges,
	})
	normalizedRules := normalizeR4Rules(input.Rules)
	graph.ruleDigest = digestR4(r4RuleDigestPayload{SchemaVersion: R4SchemaVersion, Rules: normalizedRules})
	graph.receipt = R4WorkReceipt{
		GraphNodes:        uint64(len(graph.nodes)),
		GraphEdges:        uint64(len(graph.edges)),
		ReachableNodes:    uint64(len(graph.reachableNodes)),
		ReachableEdges:    uint64(len(graph.reachableEdges)),
		SCCs:              uint64(len(graph.components)),
		CyclicSCCs:        uint64(countCyclicR4Components(graph.components)),
		CondensationEdges: uint64(len(graph.condensationEdges)),
		RuleChecks:        uint64(len(graph.components)),
		IterationChecks:   uint64(countCyclicR4Components(graph.components)),
		PathChecks:        uint64(len(graph.reachablePaths)),
	}
	return graph, ""
}

// AnalyzeR4Graph derives only graph facts. It is useful for binding a rule's
// scc_digest and does not authorize selection.
func AnalyzeR4Graph(input R4Input) (R4GraphSummary, error) {
	graph, reason := buildR4Graph(input)
	if reason != "" {
		return R4GraphSummary{}, fmt.Errorf("r4 graph: %s", reason)
	}
	components := make([]R4SCC, 0, len(graph.components))
	for _, component := range graph.components {
		components = append(components, R4SCC{Digest: component.Digest, Members: append([]string(nil), component.Members...), Cyclic: component.Cyclic})
	}
	return R4GraphSummary{
		GraphDigest:        graph.graphDigest,
		SCCDigest:          graph.sccDigest,
		CondensationDigest: graph.condensationDigest,
		SCCs:               components,
		ReachableNodes:     append([]string(nil), graph.reachableNodes...),
	}, nil
}

func deriveR4Components(nodes []string, edges []r4Edge) []r4Component {
	adjacency := make(map[string][]string, len(nodes))
	for _, node := range nodes {
		adjacency[node] = nil
	}
	for _, edge := range edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	for node := range adjacency {
		adjacency[node] = sortedUnique(adjacency[node])
	}

	index := 0
	indices := make(map[string]int, len(nodes))
	lowlink := make(map[string]int, len(nodes))
	onStack := make(map[string]bool, len(nodes))
	stack := make([]string, 0, len(nodes))
	components := make([][]string, 0)
	var visit func(string)
	visit = func(node string) {
		indices[node] = index
		lowlink[node] = index
		index++
		stack = append(stack, node)
		onStack[node] = true
		for _, next := range adjacency[node] {
			if _, seen := indices[next]; !seen {
				visit(next)
				if lowlink[next] < lowlink[node] {
					lowlink[node] = lowlink[next]
				}
			} else if onStack[next] && indices[next] < lowlink[node] {
				lowlink[node] = indices[next]
			}
		}
		if lowlink[node] != indices[node] {
			return
		}
		component := make([]string, 0)
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == node {
				break
			}
		}
		sort.Strings(component)
		components = append(components, component)
	}
	for _, node := range nodes {
		if _, seen := indices[node]; !seen {
			visit(node)
		}
	}
	sort.Slice(components, func(i, j int) bool { return components[i][0] < components[j][0] })

	componentIndex := make(map[string]int, len(nodes))
	for index, members := range components {
		for _, member := range members {
			componentIndex[member] = index
		}
	}
	result := make([]r4Component, 0, len(components))
	for index, members := range components {
		internalEdges := make([]r4Edge, 0)
		for _, edge := range edges {
			if componentIndex[edge.From] == index && componentIndex[edge.To] == index {
				internalEdges = append(internalEdges, edge)
			}
		}
		sort.Slice(internalEdges, func(i, j int) bool { return edgeKey(internalEdges[i]) < edgeKey(internalEdges[j]) })
		cyclic := len(members) > 1
		if !cyclic {
			for _, edge := range internalEdges {
				if edge.From == edge.To {
					cyclic = true
					break
				}
			}
		}
		result = append(result, r4Component{
			Digest:  digestR4(r4SCCPayload{Members: members, Edges: internalEdges}),
			Members: members,
			Edges:   internalEdges,
			Cyclic:  cyclic,
		})
	}
	return result
}

func deriveR4Condensation(components []r4Component, edges []r4Edge) [][2]string {
	componentByMember := make(map[string]string)
	for _, component := range components {
		for _, member := range component.Members {
			componentByMember[member] = component.Digest
		}
	}
	seen := make(map[[2]string]struct{})
	for _, edge := range edges {
		from, to := componentByMember[edge.From], componentByMember[edge.To]
		if from == to {
			continue
		}
		seen[[2]string{from, to}] = struct{}{}
	}
	result := make([][2]string, 0, len(seen))
	for edge := range seen {
		result = append(result, edge)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i][0] != result[j][0] {
			return result[i][0] < result[j][0]
		}
		return result[i][1] < result[j][1]
	})
	return result
}

func normalizeR4Rules(rules []R4Rule) []R4Rule {
	if rules == nil {
		return nil
	}
	result := make([]R4Rule, len(rules))
	copy(result, rules)
	sort.Slice(result, func(i, j int) bool {
		if result[i].SCCDigest != result[j].SCCDigest {
			return result[i].SCCDigest < result[j].SCCDigest
		}
		if result[i].MaxIterations != result[j].MaxIterations {
			return result[i].MaxIterations < result[j].MaxIterations
		}
		return result[i].IterationsUsed < result[j].IterationsUsed
	})
	return result
}

func digestR4(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("r4 canonical digest: %v", err))
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func edgeKey(edge r4Edge) string {
	return edge.From + "\x00" + edge.To + "\x00" + edge.PathID
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func reachableFrom(roots []string, adjacency map[string][]string) map[string]struct{} {
	seen := make(map[string]struct{}, len(adjacency))
	stack := append([]string(nil), roots...)
	for len(stack) != 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, exists := seen[node]; exists {
			continue
		}
		seen[node] = struct{}{}
		for _, next := range adjacency[node] {
			stack = append(stack, next)
		}
	}
	return seen
}

func countCyclicR4Components(components []r4Component) int {
	count := 0
	for _, component := range components {
		if component.Cyclic {
			count++
		}
	}
	return count
}
