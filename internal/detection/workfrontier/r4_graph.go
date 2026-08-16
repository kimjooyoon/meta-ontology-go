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
