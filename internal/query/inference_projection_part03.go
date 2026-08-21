package query

import (
	"fmt"
)

func (query InferenceQuery) normalized() (InferenceQuery, error) {
	if query.Schema != InferenceQuerySchema {
		return InferenceQuery{}, envelopeError(ErrInvalidInferenceQuery, "invalid_inference_schema", "schema must be gooo-query/inference/v1")
	}
	if query.Limit < 1 || query.Limit > MaxInferenceLimit {
		return InferenceQuery{}, envelopeError(ErrInvalidInferenceQuery, "invalid_inference_limit", fmt.Sprintf("must be 1..%d", MaxInferenceLimit))
	}
	if query.MaxDepth < 1 || query.MaxDepth > MaxInferenceDepth {
		return InferenceQuery{}, envelopeError(ErrInvalidInferenceQuery, "invalid_inference_depth", fmt.Sprintf("must be 1..%d", MaxInferenceDepth))
	}
	if query.MaxWork < 1 || query.MaxWork > MaxInferenceWork {
		return InferenceQuery{}, envelopeError(ErrInvalidInferenceQuery, "invalid_inference_work", fmt.Sprintf("must be 1..%d", MaxInferenceWork))
	}
	if query.Predicate != "" && !validInferencePredicate(query.Predicate) {
		return InferenceQuery{}, envelopeError(ErrInferenceUnsupportedPred, "unsupported_predicate", query.Predicate)
	}
	var err error
	query.RecordID, err = normalizeOptionalInferenceID(query.RecordID, "record_id")
	if err != nil {
		return InferenceQuery{}, err
	}
	query.SubjectID, err = normalizeOptionalInferenceID(query.SubjectID, "subject_id")
	if err != nil {
		return InferenceQuery{}, err
	}
	query.ObjectID, err = normalizeOptionalInferenceID(query.ObjectID, "object_id")
	if err != nil {
		return InferenceQuery{}, err
	}
	query.EvidenceID, err = normalizeOptionalInferenceID(query.EvidenceID, "evidence_id")
	if err != nil {
		return InferenceQuery{}, err
	}
	query.ChainStartID, err = normalizeOptionalInferenceID(query.ChainStartID, "chain_start_id")
	if err != nil {
		return InferenceQuery{}, err
	}
	query.ChainEndID, err = normalizeOptionalInferenceID(query.ChainEndID, "chain_end_id")
	if err != nil {
		return InferenceQuery{}, err
	}
	if query.Kind != "" && !query.Kind.Valid() {
		return InferenceQuery{}, invalidInferenceSelector("kind", query.Kind)
	}
	if query.Phase != "" && !query.Phase.Valid() {
		return InferenceQuery{}, invalidInferenceSelector("phase", query.Phase)
	}
	if query.Layer != "" && !query.Layer.Valid() {
		return InferenceQuery{}, invalidInferenceSelector("authority_layer", query.Layer)
	}
	if query.Effect != "" && !query.Effect.Valid() {
		return InferenceQuery{}, invalidInferenceSelector("authority_effect", query.Effect)
	}
	if query.ClaimKind != "" && !query.ClaimKind.Valid() {
		return InferenceQuery{}, invalidInferenceSelector("semantic_change_kind", query.ClaimKind)
	}
	if query.ClaimKind != "" {
		query.IncludeClaims = true
	}
	if err := validateInferencePredicateValue(query); err != nil {
		return InferenceQuery{}, err
	}
	if !query.Explain && (query.ChainStartID != "" || query.ChainEndID != "") {
		return InferenceQuery{}, envelopeError(ErrInvalidInferenceQuery, "chain_selector_without_explanation", "chain selectors require explain=true")
	}
	return query, nil
}
