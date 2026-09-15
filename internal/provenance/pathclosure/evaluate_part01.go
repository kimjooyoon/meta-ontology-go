package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func indexEdges(edges []semantic.InferenceEdge) (map[semantic.ID]semantic.InferenceEdge, []semantic.ID) {
	indexed := make(map[semantic.ID]semantic.InferenceEdge, len(edges))
	duplicates := make([]semantic.ID, 0)
	for _, edge := range edges {
		if _, exists := indexed[edge.RecordID]; exists {
			duplicates = appendID(duplicates, edge.RecordID)
			continue
		}
		indexed[edge.RecordID] = edge
	}
	sortIDs(duplicates)
	return indexed, duplicates
}
func evaluateRequirement(requirement Requirement, edges map[semantic.ID]semantic.InferenceEdge) issueClass {
	selected := make([]semantic.InferenceEdge, 0, len(requirement.RecordIDs))
	for i, recordID := range requirement.RecordIDs {
		edge, exists := edges[recordID]
		if !exists {
			return issueMissingEvidence
		}
		if edge.Kind != requirement.ExpectedKinds[i] {
			return issueMalformed
		}
		selected = append(selected, edge)
	}
	chain, err := semantic.NewInferencePathChain(selected...)
	if err != nil {
		return classifyChainError(err)
	}
	if !sameCanonicalEdges(chain.Edges, selected) ||
		chain.Edges[0].SubjectID != requirement.StartID ||
		chain.Edges[len(chain.Edges)-1].ObjectID != requirement.EndID {
		return issueMalformed
	}
	return 0
}
func classifyChainError(err error) issueClass {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "path_orphan") || strings.Contains(message, "path_ambiguity") {
		return issueMissingEvidence
	}
	return issueMalformed
}
func sameCanonicalEdges(left, right []semantic.InferenceEdge) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].Canonical() != right[i].Canonical() {
			return false
		}
	}
	return true
}
