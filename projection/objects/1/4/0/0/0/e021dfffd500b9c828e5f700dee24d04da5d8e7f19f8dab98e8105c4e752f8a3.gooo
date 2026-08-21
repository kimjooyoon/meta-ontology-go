package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func collectEvidence(records []semantic.InferenceEvidence, beforeDigest, afterDigest string, config EvaluationConfig) (map[semantic.ID]semantic.InferenceEvidence, string) {
	result := make(map[semantic.ID]semantic.InferenceEvidence, len(records))
	for _, record := range records {
		if !validID(record.ID.String()) || !validDigest(record.Digest) || record.Before.Semantic != beforeDigest || record.After.Semantic != afterDigest || !validSnapshot(record.Before) || !validSnapshot(record.After) || !validControls(record.Controls, config) {
			return nil, "evidence"
		}
		if _, duplicate := result[record.ID]; duplicate {
			return nil, "duplicate-evidence"
		}
		result[record.ID] = record
	}
	return result, ""
}
func collectEdges(records []semantic.InferenceEdge, evidence map[semantic.ID]semantic.InferenceEvidence, beforeDigest, afterDigest string, config EvaluationConfig) (map[semantic.ID]semantic.InferenceEdge, string) {
	result := make(map[semantic.ID]semantic.InferenceEdge, len(records))
	for _, edge := range records {
		if !validRecord(edge.InferenceRecord, evidence, beforeDigest, afterDigest, config) || !edge.Kind.Valid() {
			return nil, "edge"
		}
		if _, duplicate := result[edge.RecordID]; duplicate {
			return nil, "duplicate-edge"
		}
		if !validKindBinding(edge) {
			return nil, "kind-binding"
		}
		if edge.Kind == semantic.InferenceObservationCandidate && edge.Before.Semantic != edge.After.Semantic {
			return nil, "candidate-authority"
		}
		if edge.Kind == semantic.InferenceAuthoritativeDeclaration && len(edge.SourceRoots) == 0 {
			return nil, "declaration-root"
		}
		if edge.Kind == semantic.InferenceAcceptedLift {
			if edge.AcceptanceReceipt == "" || !sourceBackedReceipt(edge.AcceptanceReceipt, edge.Evidence, evidence) {
				return nil, "accepted-lift"
			}
		} else if edge.AcceptanceReceipt != "" {
			return nil, "unexpected-receipt"
		}
		result[edge.RecordID] = edge
	}
	return result, ""
}
func collectClaims(records []semantic.SemanticChangeClaim, evidence map[semantic.ID]semantic.InferenceEvidence, beforeDigest, afterDigest, deltaText string) (map[semantic.ID]semantic.SemanticChangeClaim, string) {
	result := make(map[semantic.ID]semantic.SemanticChangeClaim, len(records))
	for _, claim := range records {
		if !validRecord(claim.InferenceRecord, evidence, beforeDigest, afterDigest, EvaluationConfig{}) {
			return nil, "claim"
		}
		if claim.Authority.Layer != semantic.AuthoritySemantic || !claim.Kind.Valid() {
			return nil, "claim-kind"
		}
		switch claim.Kind {
		case semantic.SemanticDelta:
			if claim.Before.Semantic == claim.After.Semantic || claim.Authority.Effect != semantic.AuthorityDelta || claim.CanonicalDelta != deltaText || claim.DeltaDigest != digestBytes([]byte(deltaText)) {
				return nil, "delta-claim"
			}
		case semantic.NoSemanticDelta:
			if claim.Before.Semantic != claim.After.Semantic || claim.Authority.Effect != semantic.AuthorityNoChange || claim.CanonicalDelta != "" || claim.DeltaDigest != "" {
				return nil, "no-delta-claim"
			}
		}
		if _, duplicate := result[claim.RecordID]; duplicate {
			return nil, "duplicate-claim"
		}
		result[claim.RecordID] = claim
	}
	return result, ""
}
