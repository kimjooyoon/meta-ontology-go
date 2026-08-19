package workfrontier

import (
	"fmt"
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
