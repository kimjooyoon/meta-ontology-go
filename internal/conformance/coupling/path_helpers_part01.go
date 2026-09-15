package coupling

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func sameEvidenceIDs(ids []string, refs []semantic.EvidenceReference) bool {
	if len(ids) != len(sortedUnique(ids)) {
		return false
	}
	left := sortedUnique(ids)
	right := make([]string, 0, len(refs))
	for _, ref := range refs {
		right = append(right, ref.ID.String())
	}
	right = sortedUnique(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func hasIndependentEvidence(refs []semantic.EvidenceReference, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	for _, ref := range refs {
		if record, ok := evidence[ref.ID]; ok && record.Independent {
			return true
		}
	}
	return false
}
func countIndependentEdges(edges map[semantic.ID]semantic.InferenceEdge) int {
	count := 0
	for _, edge := range edges {
		if edge.Kind == semantic.InferenceIndependentVerification {
			count++
		}
	}
	return count
}
func parseUniqueIDs(values []string) ([]semantic.ID, string) {
	result := make([]semantic.ID, 0, len(values))
	seen := make(map[semantic.ID]struct{}, len(values))
	for _, value := range values {
		id, err := semantic.ParseIdentity(value)
		if err != nil {
			return nil, "invalid-id"
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, "duplicate-id"
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, ""
}
func pathDigest(path semantic.InferencePathV1) string {
	raw := pathToWire(path)
	normalizePathWire(&raw)
	data, _ := json.Marshal(raw)
	return digestBytes(data)
}
