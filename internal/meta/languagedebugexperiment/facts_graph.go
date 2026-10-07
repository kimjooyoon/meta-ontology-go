package languagedebugexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"

var canonicalDebugActivities = []string{
	"languageutility://activity/observe-debugging-deterministic-replay",
	"languageutility://activity/observe-debugging-resource-observed",
}

func expectedGraphActivityCount() int {
	return languageutility.ExpectedCellCount() + len(canonicalDebugActivities)
}

func expectedGraphEdgeCount() int {
	return expectedGraphActivityCount() * 2
}

func validDebugGraph(graph GraphObservation) bool {
	activities := make(map[string]bool, len(canonicalDebugActivities))
	for _, activity := range canonicalDebugActivities {
		activities[activity] = true
	}
	if len(graph.DebugActivityIDs) != len(activities) {
		return false
	}
	seenActivities := map[string]bool{}
	for _, activity := range graph.DebugActivityIDs {
		if !activities[activity] || seenActivities[activity] {
			return false
		}
		seenActivities[activity] = true
	}
	edges := map[string]bool{
		"used\x00languageutility://activity/observe-debugging-deterministic-replay\x00gooo://meta/language-utility/entity/cell":               true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-deterministic-replay": true,
		"used\x00languageutility://activity/observe-debugging-resource-observed\x00gooo://meta/language-utility/entity/cell":                  true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-resource-observed":    true,
	}
	if len(graph.DebugCausalEdges) != len(edges) {
		return false
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

func validHexDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}
