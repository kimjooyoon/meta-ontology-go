package selfimprovementloop

import (
	"fmt"
	"strings"
)

func BindActivities(graph Graph) ([]ActivityBinding, error) {
	if graph.SchemaVersion != GraphSchemaVersion {
		return nil, fmt.Errorf("released semantic graph schema = %q, want %q", graph.SchemaVersion, GraphSchemaVersion)
	}
	if !validDigest(graph.GraphHash) || !validDigest(graph.SourceDigest) {
		return nil, fmt.Errorf("released semantic graph digests are not sha256 values")
	}
	activities := make(map[string]GraphNode)
	for _, node := range graph.Nodes {
		if node.Kind != "Activity" {
			continue
		}
		name := strings.TrimSpace(node.Name)
		if name == "" || strings.TrimSpace(node.ID) == "" || strings.TrimSpace(node.Namespace) != "meta" {
			return nil, fmt.Errorf("released semantic graph contains an incomplete activity")
		}
		if _, exists := activities[name]; exists {
			return nil, fmt.Errorf("released semantic graph duplicates activity %q", name)
		}
		activities[name] = node
	}
	if len(activities) != len(fixedCells) {
		return nil, fmt.Errorf("released semantic graph activities = %d, want %d", len(activities), len(fixedCells))
	}
	bindings := make([]ActivityBinding, 0, len(fixedCells))
	ids := make(map[string]string, len(fixedCells))
	for _, cell := range fixedCells {
		node, exists := activities[cell]
		if !exists {
			return nil, fmt.Errorf("semantic cell %q has no released Gooo activity", cell)
		}
		if previous, exists := ids[node.ID]; exists {
			return nil, fmt.Errorf("activities %q and %q share id %q", previous, cell, node.ID)
		}
		ids[node.ID] = cell
		bindings = append(bindings, ActivityBinding{Cell: cell, Activity: node.Name, ActivityID: node.ID})
	}
	return bindings, nil
}
