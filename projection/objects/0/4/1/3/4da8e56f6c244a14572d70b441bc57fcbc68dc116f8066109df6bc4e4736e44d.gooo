package query

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeOptionalInferenceID(id ID, field string) (ID, error) {
	if id == "" {
		return "", nil
	}
	parsed, err := ParseID(id.String())
	if err != nil {
		return "", envelopeError(ErrInvalidInferenceQuery, "invalid_"+field, err.Error())
	}
	return parsed, nil
}
func invalidInferenceSelector(field string, value fmt.Stringer) error {
	return envelopeError(ErrInvalidInferenceQuery, "invalid_"+field, value.String())
}
func validInferencePredicate(predicate string) bool {
	switch predicate {
	case InferencePredicateRecordID, InferencePredicateSubjectID, InferencePredicateObjectID,
		InferencePredicateEvidence, InferencePredicateKind, InferencePredicatePhase,
		InferencePredicateLayer, InferencePredicateEffect, InferencePredicateClaimKind:
		return true
	default:
		return false
	}
}
func validateInferencePredicateValue(query InferenceQuery) error {
	var present bool
	switch query.Predicate {
	case InferencePredicateRecordID:
		present = query.RecordID != ""
	case InferencePredicateSubjectID:
		present = query.SubjectID != ""
	case InferencePredicateObjectID:
		present = query.ObjectID != ""
	case InferencePredicateEvidence:
		present = query.EvidenceID != ""
	case InferencePredicateKind:
		present = query.Kind != ""
	case InferencePredicatePhase:
		present = query.Phase != ""
	case InferencePredicateLayer:
		present = query.Layer != ""
	case InferencePredicateEffect:
		present = query.Effect != ""
	case InferencePredicateClaimKind:
		present = query.ClaimKind != ""
	default:
		return nil
	}
	if !present {
		return envelopeError(ErrInvalidInferenceQuery, "missing_predicate_value", query.Predicate)
	}
	return nil
}
func controlsEmpty(controls semantic.InferenceControls) bool {
	return controls.CatalogDigest == "" && controls.PolicyDigest == "" &&
		controls.Profile.ID == "" && controls.Profile.Version == "" && controls.Profile.Digest == ""
}
func controlsEqual(left, right semantic.InferenceControls) bool {
	return left == right
}
func snapshotsMatch(expected, actual semantic.SnapshotDigests) bool {
	return (expected.Source == "" || expected.Source == actual.Source) &&
		(expected.Semantic == "" || expected.Semantic == actual.Semantic)
}
