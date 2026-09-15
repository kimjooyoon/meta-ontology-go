package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func sourceBackedReceipt(id semantic.ID, refs []semantic.EvidenceReference, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	for _, ref := range refs {
		if ref.ID == id {
			record, ok := evidence[id]
			return ok && record.SourceBacked && record.Before.Source != "" && record.After.Source != ""
		}
	}
	return false
}
func declarationRootsMatch(edges map[semantic.ID]semantic.InferenceEdge, roots map[semantic.ID]struct{}) bool {
	seen := make(map[semantic.ID]struct{})
	for _, edge := range edges {
		if edge.Kind != semantic.InferenceAuthoritativeDeclaration {
			continue
		}
		for _, root := range edge.SourceRoots {
			if _, ok := roots[root]; !ok {
				return false
			}
			seen[root] = struct{}{}
		}
	}
	return len(seen) == len(roots)
}
