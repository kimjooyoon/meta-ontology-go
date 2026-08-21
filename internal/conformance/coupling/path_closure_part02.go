package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func sameReceiptEvidence(ids []string, chain []semantic.InferenceEdge, claim semantic.SemanticChangeClaim, evidence map[semantic.ID]semantic.InferenceEvidence) bool {
	if len(ids) != len(sortedUnique(ids)) {
		return false
	}
	want := make(map[string]struct{})
	for _, ref := range claim.Evidence {
		want[ref.ID.String()] = struct{}{}
	}
	for _, edge := range chain {
		for _, ref := range edge.Evidence {
			want[ref.ID.String()] = struct{}{}
		}
	}
	actual := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		parsed, err := semantic.ParseIdentity(id)
		if err != nil {
			return false
		}
		if _, ok := evidence[parsed]; !ok {
			return false
		}
		actual[id] = struct{}{}
	}
	if len(actual) != len(want) {
		return false
	}
	for id := range want {
		if _, ok := actual[id]; !ok {
			return false
		}
	}
	return true
}
