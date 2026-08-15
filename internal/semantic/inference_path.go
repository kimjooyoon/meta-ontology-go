package semantic

import (
	"errors"
	"strings"
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

func (k SemanticChangeKind) Valid() bool { return k == SemanticDelta || k == NoSemanticDelta }

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

const (
	AuthoritySource       AuthorityLayer = "SOURCE"
	AuthoritySemantic     AuthorityLayer = "SEMANTIC"
	AuthorityDerived      AuthorityLayer = "DERIVED"
	AuthorityCandidate    AuthorityLayer = "CANDIDATE"
	AuthorityVerification AuthorityLayer = "VERIFICATION"
)

func (l AuthorityLayer) Valid() bool {
	switch l {
	case AuthoritySource, AuthoritySemantic, AuthorityDerived, AuthorityCandidate, AuthorityVerification:
		return true
	default:
		return false
	}
}

func (l AuthorityLayer) String() string { return string(l) }

type AuthorityEffect string

const (
	AuthorityDeclare  AuthorityEffect = "DECLARE"
	AuthorityDerive   AuthorityEffect = "DERIVE"
	AuthorityProject  AuthorityEffect = "PROJECT"
	AuthorityObserve  AuthorityEffect = "OBSERVE"
	AuthorityLift     AuthorityEffect = "LIFT"
	AuthorityVerify   AuthorityEffect = "VERIFY"
	AuthorityDelta    AuthorityEffect = "SEMANTIC_DELTA"
	AuthorityNoChange AuthorityEffect = "NO_SEMANTIC_CHANGE"
)

func (e AuthorityEffect) Valid() bool {
	switch e {
	case AuthorityDeclare, AuthorityDerive, AuthorityProject, AuthorityObserve,
		AuthorityLift, AuthorityVerify, AuthorityDelta, AuthorityNoChange:
		return true
	default:
		return false
	}
}

func (e AuthorityEffect) String() string { return string(e) }

type PhasePlacement struct {
	Phase   InferencePhase
	Ordinal uint64
}

type AuthorityBinding struct {
	Layer  AuthorityLayer
	Effect AuthorityEffect
}

type RuleBinding struct {
	ID      ID
	Version string
	Digest  string
}

type SnapshotDigests struct {
	Source   string
	Semantic string
}

type ProfileBinding struct {
	ID      string
	Version string
	Digest  string
}

type InferenceControls struct {
	CatalogDigest string
	PolicyDigest  string
	Profile       ProfileBinding
}

// EvidenceReference has no producer or actor field. Its digest binds the
// reference to an append-only, producer-independent evidence record.
type EvidenceReference struct {
	ID     ID
	Digest string
}

type InferenceEvidence struct {
	ID           ID
	Digest       string
	Before       SnapshotDigests
	After        SnapshotDigests
	SourceBacked bool
	Independent  bool
	Controls     InferenceControls
}

// InferenceRecord contains the identity and proof tuple common to every edge
// and semantic-change claim. No display name, path, timestamp, or actor is in
// this tuple.
type InferenceRecord struct {
	RecordID  ID
	SubjectID ID
	ObjectID  ID
	Rule      RuleBinding
	Phase     PhasePlacement
	Before    SnapshotDigests
	After     SnapshotDigests
	Authority AuthorityBinding
	Evidence  []EvidenceReference
	Controls  InferenceControls
}

type InferenceEdge struct {
	InferenceRecord
	Kind              InferenceKind
	SourceRoots       []ID
	AcceptanceReceipt ID
}

type SemanticChangeClaim struct {
	InferenceRecord
	Kind           SemanticChangeKind
	CanonicalDelta string
	DeltaDigest    string
}

// InferencePathV1 is the finite, append-only semantic record set. It is a
// typed evidence carrier, not a graph database, scheduler, or PROV replacement.
type InferencePathV1 struct {
	Version  string
	Edges    []InferenceEdge
	Claims   []SemanticChangeClaim
	Evidence []InferenceEvidence
}

type InferencePathIssue struct {
	Code   string
	Record ID
	Detail string
}

type InferencePathErrors struct{ Issues []InferencePathIssue }

func (e *InferencePathErrors) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrInferencePath.Error()
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, issue.Code+": "+issue.Detail)
	}
	return ErrInferencePath.Error() + ": " + strings.Join(parts, "; ")
}

func (e *InferencePathErrors) Unwrap() error { return ErrInferencePath }

func (e *InferencePathErrors) add(code string, record ID, detail string) {
	e.Issues = append(e.Issues, InferencePathIssue{Code: code, Record: record, Detail: detail})
}
