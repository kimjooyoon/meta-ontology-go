package bodycodegen

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SourceControlColumns copies retained IR facts. Gooo owns their interpretation.
type SourceControlColumns struct {
	Schema         string   `json:"schema"`
	Scope          string   `json:"scope"`
	SourceSHA256   string   `json:"source_sha256"`
	ActivityID     string   `json:"activity_id"`
	GraphSHA256    string   `json:"graph_sha256"`
	Count          int      `json:"count"`
	IDs            []uint16 `json:"ids"`
	Bits           []int64  `json:"bits"`
	ParentA        []uint16 `json:"parent_a"`
	ParentB        []uint16 `json:"parent_b"`
	Guards         []uint16 `json:"guards"`
	Conditions     []uint16 `json:"conditions"`
	Kinds          []string `json:"kinds"`
	Operators      []string `json:"operators"`
	KnownTruth     []string `json:"known_truth"`
	NodeActivities []string `json:"node_activities"`
	FieldIDs       []string `json:"field_ids"`
	InputPorts     []string `json:"input_ports"`
}

func TypedSourceControlInputs(r TypedValueOrigins, target string) ([]byte, error) {
	if r.Flow != nil && len(r.Flow.RecordVersions) != 0 {
		return nil, fmt.Errorf("record-version origins require their layout, value and version tables; flat columns cannot retain them")
	}

	pkg, activity, ok := strings.Cut(target, ":")
	if !ok || pkg == "" || activity == "" || strings.ContainsAny(target, " \t\r\n") || strings.Contains(activity, ":") {
		return nil, fmt.Errorf("consumer input needs package/path:Activity")
	}
	if r.Status != "RESOLVED" || r.Flow == nil || r.Flow.Status != "RESOLVED" || len(r.Flow.Nodes) < 1 || len(r.Flow.Nodes) > 60 {
		return nil, fmt.Errorf("source control input requires a resolved graph with 1..60 nodes")
	}
	graph, err := json.Marshal(r.Flow)
	if err != nil {
		return nil, err
	}
	c := newSourceControlColumns(r, digest(graph))
	ports := make(map[uint16]string, len(r.Inputs))
	for _, input := range r.Inputs {
		ports[input.Node] = input.Port
	}
	for i, node := range r.Flow.Nodes {
		if int(node.ID) != i+1 || node.Parents[0] >= node.ID || node.Parents[1] >= node.ID || node.Guard >= node.ID || node.Condition >= node.ID {
			return nil, fmt.Errorf("source control graph is not topologically ordered")
		}
		c.IDs[i], c.Bits[i] = node.ID, int64(1)<<i
		c.ParentA[i], c.ParentB[i] = node.Parents[0], node.Parents[1]
		c.Guards[i], c.Conditions[i] = node.Guard, node.Condition
		c.Kinds[i], c.Operators[i], c.KnownTruth[i] = node.Kind, node.Operator, node.KnownTruth
		c.NodeActivities[i], c.FieldIDs[i], c.InputPorts[i] = node.Span.ActivityID, node.FieldID, ports[node.ID]
	}
	output := struct {
		Schema string                            `json:"schema"`
		Inputs []map[string]SourceControlColumns `json:"inputs"`
	}{Schema: "gooo/body-composition-inputs/v1", Inputs: []map[string]SourceControlColumns{{target: c}}}
	raw, err := json.Marshal(output)
	if err == nil && len(raw) > 32<<10 {
		err = fmt.Errorf("source control input exceeds 32 KiB")
	}
	return raw, err
}

func newSourceControlColumns(r TypedValueOrigins, graphSHA string) SourceControlColumns {
	n := len(r.Flow.Nodes)
	return SourceControlColumns{Schema: "gooo/source-control-columns/v1", Scope: "retained_source_graph",
		SourceSHA256: r.SourceSHA256, ActivityID: r.ActivityID, GraphSHA256: graphSHA, Count: n,
		IDs: make([]uint16, n), Bits: make([]int64, n), ParentA: make([]uint16, n), ParentB: make([]uint16, n),
		Guards: make([]uint16, n), Conditions: make([]uint16, n), Kinds: make([]string, n),
		Operators: make([]string, n), KnownTruth: make([]string, n), NodeActivities: make([]string, n),
		FieldIDs: make([]string, n), InputPorts: make([]string, n)}
}
