package query

import (
	"errors"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	// InferenceQuerySchema is a separate versioned read contract for typed
	// inference evidence. It does not add vocabulary to the semantic kernel.
	InferenceQuerySchema  = "gooo-query/inference/v1"
	DefaultInferenceLimit = 100
	MaxInferenceLimit     = MaxEnvelopeLimit
	DefaultInferenceDepth = MaxEnvelopeDepth
	MaxInferenceDepth     = MaxEnvelopeDepth
	DefaultInferenceWork  = 10000
	MaxInferenceWork      = 100000
)

var (
	ErrInvalidInferenceQuery    = errors.New("invalid inference query")
	ErrInferenceQueryBudget     = errors.New("inference query budget exceeded")
	ErrInferenceBudget          = ErrInferenceQueryBudget
	ErrInferenceStaleSnapshot   = errors.New("stale inference snapshot")
	ErrInferenceUnsupportedPred = errors.New("unsupported inference predicate")
	ErrInferenceChain           = errors.New("invalid inference explanation chain")
)

// InferencePredicate names the stable fields that may be selected by a
// request. The semantic path itself remains typed; predicates are query
// selectors, not a second semantic vocabulary.
const (
	InferencePredicateRecordID  = "record_id"
	InferencePredicateSubjectID = "subject_id"
	InferencePredicateObjectID  = "object_id"
	InferencePredicateEvidence  = "evidence_id"
	InferencePredicateKind      = "kind"
	InferencePredicatePhase     = "phase"
	InferencePredicateLayer     = "authority_layer"
	InferencePredicateEffect    = "authority_effect"
	InferencePredicateClaimKind = "semantic_change_kind"
)

// NewInferenceProjection validates and normalizes exactly once through the
// semantic kernel. The input path is never mutated or retained by reference.
func NewInferenceProjection(path semantic.InferencePathV1) (*InferenceProjection, error) {
	normalized, err := path.Normalized()
	if err != nil {
		return nil, err
	}
	return &InferenceProjection{path: normalized}, nil
}

// FromInferencePath is an adapter-oriented constructor spelling.
func FromInferencePath(path semantic.InferencePathV1) (*InferenceProjection, error) {
	return NewInferenceProjection(path)
}

// ProjectInferencePath is a function-oriented constructor spelling.
func ProjectInferencePath(path semantic.InferencePathV1) (*InferenceProjection, error) {
	return NewInferenceProjection(path)
}

// Path returns a detached normalized snapshot.
func (projection InferenceProjection) Path() semantic.InferencePathV1 {
	return cloneInferencePath(projection.path)
}

// Canonical and StableHash expose the normalized source snapshot receipt, not
// a mutable query result or a new authority graph hash.
func (projection InferenceProjection) Canonical() string { return projection.path.Canonical() }

func (projection InferenceProjection) StableHash() string { return projection.path.StableHash() }

func cloneInferencePath(path semantic.InferencePathV1) semantic.InferencePathV1 {
	out := semantic.InferencePathV1{Version: path.Version}
	out.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	out.Claims = append([]semantic.SemanticChangeClaim(nil), path.Claims...)
	out.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	for i := range out.Edges {
		out.Edges[i].SourceRoots = append([]semantic.ID(nil), path.Edges[i].SourceRoots...)
		out.Edges[i].Evidence = append([]semantic.EvidenceReference(nil), path.Edges[i].Evidence...)
	}
	for i := range out.Claims {
		out.Claims[i].Evidence = append([]semantic.EvidenceReference(nil), path.Claims[i].Evidence...)
	}
	return out
}

func (query InferenceQuery) hasSelectors() bool {
	return query.RecordID != "" || query.SubjectID != "" || query.ObjectID != "" ||
		query.EvidenceID != "" || query.Kind != "" || query.Phase != "" ||
		query.Layer != "" || query.Effect != "" || query.ClaimKind != "" ||
		query.hasSnapshotOrControlSelectors()
}

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

func inferenceRecordMatches(query InferenceQuery, record semantic.InferenceRecord, kind semantic.InferenceKind, claimKind semantic.SemanticChangeKind, isClaim bool) bool {
	if query.RecordID != "" && query.RecordID != ID(record.RecordID.String()) {
		return false
	}
	if query.SubjectID != "" && query.SubjectID != ID(record.SubjectID.String()) {
		return false
	}
	if query.ObjectID != "" && query.ObjectID != ID(record.ObjectID.String()) {
		return false
	}
	if query.Kind != "" && (isClaim || query.Kind != kind) {
		return false
	}
	if query.Phase != "" && query.Phase != record.Phase.Phase {
		return false
	}
	if query.Layer != "" && query.Layer != record.Authority.Layer {
		return false
	}
	if query.Effect != "" && query.Effect != record.Authority.Effect {
		return false
	}
	if query.ClaimKind != "" && (!isClaim || query.ClaimKind != claimKind) {
		return false
	}
	if !snapshotsMatch(query.Before, record.Before) || !snapshotsMatch(query.After, record.After) {
		return false
	}
	if !controlsEmpty(query.Controls) && !controlsEqual(query.Controls, record.Controls) {
		return false
	}
	return true
}

func evidenceReferencesMatch(id ID, refs []semantic.EvidenceReference) bool {
	if id == "" {
		return true
	}
	for _, ref := range refs {
		if ID(ref.ID.String()) == id {
			return true
		}
	}
	return false
}
