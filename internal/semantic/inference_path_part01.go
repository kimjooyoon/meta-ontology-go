package semantic

import (
	"errors"
)

const InferencePathSchemaVersion = "gooo-inference-path/v1"

var ErrInferencePath = errors.New("invalid inference path")

// InferenceKind is the closed set of typed transitions. ObservationCandidate
// is deliberately not an authority edge; SemanticChangeKind below is a
// separate sum and cannot be used as an inference kind.
type InferenceKind string

const (
	InferenceAuthoritativeDeclaration InferenceKind = "AUTHORITATIVE_DECLARATION"
	InferenceDeterministicDerivation  InferenceKind = "DETERMINISTIC_DERIVATION"
	InferenceDerivedProjection        InferenceKind = "DERIVED_PROJECTION"
	InferenceObservationCandidate     InferenceKind = "OBSERVATION_CANDIDATE"
	InferenceAcceptedLift             InferenceKind = "ACCEPTED_LIFT"
	InferenceIndependentVerification  InferenceKind = "INDEPENDENT_VERIFICATION"
)

func (k InferenceKind) Valid() bool {
	switch k {
	case InferenceAuthoritativeDeclaration, InferenceDeterministicDerivation,
		InferenceDerivedProjection, InferenceObservationCandidate,
		InferenceAcceptedLift, InferenceIndependentVerification:
		return true
	default:
		return false
	}
}
func (k InferenceKind) String() string { return string(k) }

// SemanticChangeKind is a closed claim sum, not an inference transition.
type SemanticChangeKind string

const (
	SemanticDelta   SemanticChangeKind = "SEMANTIC_DELTA"
	NoSemanticDelta SemanticChangeKind = "NO_SEMANTIC_DELTA"
)

func (k SemanticChangeKind) Valid() bool    { return k == SemanticDelta || k == NoSemanticDelta }
func (k SemanticChangeKind) String() string { return string(k) }

type InferencePhase string

const (
	PhaseDeclaration  InferencePhase = "DECLARATION"
	PhaseDerivation   InferencePhase = "DERIVATION"
	PhaseProjection   InferencePhase = "PROJECTION"
	PhaseObservation  InferencePhase = "OBSERVATION"
	PhaseLift         InferencePhase = "LIFT"
	PhaseVerification InferencePhase = "VERIFICATION"
)

func (p InferencePhase) Valid() bool {
	switch p {
	case PhaseDeclaration, PhaseDerivation, PhaseProjection, PhaseObservation, PhaseLift, PhaseVerification:
		return true
	default:
		return false
	}
}
func (p InferencePhase) String() string { return string(p) }

type AuthorityLayer string
