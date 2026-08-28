package languageutility

import "strings"

func refutedResult(result CellResult, step, reason string) CellResult {
	result.State, result.ClaimStatus, result.Resolution = StateRefuted, "REFUTED", "EXACT"
	result.Step, result.Reason = step, reason
	return result
}

func validateDebugBinding(observed CellObservation, graph GraphObservation, stage string) string {
	activity := "languageutility://activity/observe-debugging-" + strings.ToLower(strings.ReplaceAll(stage, "_", "-"))
	if !validDebugGraph(graph) {
		return "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"
	}
	if observed.MetaActivityID != activity || observed.MetaInputID != "gooo://meta/language-utility/entity/cell" || observed.MetaOutputID != "gooo://meta/language-utility/entity/evidence" || observed.ActivityMatches != 1 || observed.OutputMatches != 1 || observed.UsedEdgeMatches != 1 || observed.GeneratedEdgeMatches != 1 || len(observed.CausalEdges) != 2 {
		return "GOOO_META_ACTIVITY_OUTPUT_OR_EDGE_MISSING"
	}
	seen := map[string]bool{}
	for _, edge := range observed.CausalEdges {
		key := edge.Relation + "\x00" + edge.Subject + "\x00" + edge.Object
		if seen[key] {
			return "GOOO_CAUSAL_EDGE_DUPLICATE"
		}
		seen[key] = true
	}
	if !seen["used\x00"+activity+"\x00gooo://meta/language-utility/entity/cell"] || !seen["wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00"+activity] {
		return "GOOO_CAUSAL_EDGE_MISSING"
	}
	return ""
}

func validDebugGraph(graph GraphObservation) bool {
	activities := map[string]bool{"languageutility://activity/observe-debugging-deterministic-replay": true, "languageutility://activity/observe-debugging-resource-observed": true}
	if graph.Schema != "gooo-graph/v1" || graph.ActivityCount != 44 || graph.EdgeCount != 88 || graph.DebugActivityCount != 2 || graph.DebugOutputCount != 2 || graph.DebugUsedEdgeCount != 2 || graph.DebugGeneratedEdgeCount != 2 || len(graph.DebugActivityIDs) != 2 || len(graph.DebugCausalEdges) != 4 {
		return false
	}
	seenActivities := map[string]bool{}
	for _, value := range graph.DebugActivityIDs {
		if !activities[value] || seenActivities[value] {
			return false
		}
		seenActivities[value] = true
	}
	edges := map[string]bool{
		"used\x00languageutility://activity/observe-debugging-deterministic-replay\x00gooo://meta/language-utility/entity/cell":               true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-deterministic-replay": true,
		"used\x00languageutility://activity/observe-debugging-resource-observed\x00gooo://meta/language-utility/entity/cell":                  true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-resource-observed":    true,
	}
	seenEdges := map[string]bool{}
	for _, edge := range graph.DebugCausalEdges {
		key := edge.Relation + "\x00" + edge.Subject + "\x00" + edge.Object
		if !edges[key] || seenEdges[key] {
			return false
		}
		seenEdges[key] = true
	}
	return len(seenEdges) == len(edges)
}
