package query

import (
	"errors"
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
func (projection InferenceProjection) Canonical() string  { return projection.path.Canonical() }
func (projection InferenceProjection) StableHash() string { return projection.path.StableHash() }
