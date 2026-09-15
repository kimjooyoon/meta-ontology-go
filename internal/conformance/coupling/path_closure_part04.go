package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func chainHasProjection(chain []semantic.InferenceEdge, binding CodeBinding) bool {
	hasOwner := false
	for _, edge := range chain {
		if edge.SubjectID.String() == binding.SemanticOwnerID || edge.ObjectID.String() == binding.SemanticOwnerID {
			hasOwner = true
		}
	}
	if !hasOwner {
		return false
	}
	for _, edge := range chain {
		if edge.Kind == semantic.InferenceDerivedProjection && edge.ObjectID.String() == binding.CodeSymbolID {
			return true
		}
	}
	return false
}
func claimMatchesReceipt(kind semantic.SemanticChangeKind, claim ChangeClaim) bool {
	switch claim {
	case ClaimDelta:
		return kind == semantic.SemanticDelta
	case ClaimNoDelta:
		return kind == semantic.NoSemanticDelta
	default:
		return false
	}
}
func rootsUsed(roots []semantic.ID, edges map[semantic.ID]semantic.InferenceEdge, used map[semantic.ID]struct{}) bool {
	seen := make(map[semantic.ID]struct{})
	for id := range used {
		edge := edges[id]
		if edge.Kind == semantic.InferenceAuthoritativeDeclaration {
			seen[edge.SubjectID] = struct{}{}
		}
	}
	if len(seen) != len(roots) {
		return false
	}
	for _, root := range roots {
		if _, ok := seen[root]; !ok {
			return false
		}
	}
	return true
}
