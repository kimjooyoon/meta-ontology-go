package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func baselinePathParts(input Input, before, after, delta string) (semantic.ID, map[string]semantic.InferenceEdge, map[string]semantic.SemanticChangeClaim, bool) {
	root, err := semantic.ParseIdentity(input.Roots[0])
	if err != nil {
		return "", nil, nil, false
	}
	edges := map[string]semantic.InferenceEdge{}
	for _, edge := range input.Path.Edges {
		if !edge.Kind.Valid() || edge.RecordID == "" || !baselineID(edge.RecordID.String()) || edges[edge.RecordID.String()].RecordID != "" {
			return "", nil, nil, false
		}
		edges[edge.RecordID.String()] = edge
	}
	claims := map[string]semantic.SemanticChangeClaim{}
	for _, claim := range input.Path.Claims {
		if !baselineID(claim.RecordID.String()) || claims[claim.RecordID.String()].RecordID != "" || claim.Before.Semantic != before || claim.After.Semantic != after || !claim.Kind.Valid() {
			return "", nil, nil, false
		}
		if claim.Kind == semantic.SemanticDelta && (claim.CanonicalDelta != delta || claim.DeltaDigest != baselineHash(delta)) || claim.Kind == semantic.NoSemanticDelta && (before != after || claim.CanonicalDelta != "" || claim.DeltaDigest != "") {
			return "", nil, nil, false
		}
		claims[claim.RecordID.String()] = claim
	}
	evidenceIDs := map[string]bool{}
	for _, evidence := range input.Path.Evidence {
		if !baselineID(evidence.ID.String()) || evidence.Before.Semantic != before || evidence.After.Semantic != after || evidenceIDs[evidence.ID.String()] {
			return "", nil, nil, false
		}
		evidenceIDs[evidence.ID.String()] = true
	}
	return root, edges, claims, true
}
