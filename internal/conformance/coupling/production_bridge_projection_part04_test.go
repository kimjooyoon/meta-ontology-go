package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func productionSelectedEvidence(path semantic.InferencePathV1, ids []semantic.ID) []semantic.ID {
	byID := make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges))
	for _, edge := range path.Edges {
		byID[edge.RecordID] = edge
	}
	selected := make(map[semantic.ID]struct{})
	for _, id := range ids {
		for _, ref := range byID[id].Evidence {
			selected[ref.ID] = struct{}{}
		}
	}
	result := make([]semantic.ID, 0, len(selected))
	for id := range selected {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
func productionEvidenceDigest(path semantic.InferencePathV1, id semantic.ID) string {
	for _, evidence := range path.Evidence {
		if evidence.ID == id {
			return bridgeRawDigest(evidence.Digest)
		}
	}
	return ""
}
