package eligibilityregistry

type StableID string

type Digest string

type ItemKind uint8

const (
	ItemSemantic   ItemKind = 0
	ItemStructural ItemKind = 1
)

type AuthorityKind uint8

const (
	AuthorityBusinessDSL AuthorityKind = 0
	AuthoritySemanticIR  AuthorityKind = 1
)

type ProjectionKind uint8

const (
	ProjectionSemanticIR  ProjectionKind = 0
	ProjectionGeneratedGo ProjectionKind = 1
)

type RegistryEntry struct {
	ID                 StableID
	Kind               ItemKind
	Authority          AuthorityKind
	RequiredProjection ProjectionKind
}

type ProjectionObservation struct {
	ID             StableID
	Kind           ProjectionKind
	RegistryDigest Digest
}

type TrustedSourceSnapshot struct {
	Digest Digest
}

type Input struct {
	RegistrySourceDigest Digest
	CurrentTrustedSource TrustedSourceSnapshot
	Entries              []RegistryEntry
	Observations         []ProjectionObservation
}

type Decision uint8

const (
	DecisionPass       Decision = 0
	DecisionUnknown    Decision = 1
	DecisionFailClosed Decision = 2
)

type Reason uint8

const (
	ReasonNone                Reason = 0
	ReasonMissingRegistry     Reason = 1
	ReasonZeroDenominator     Reason = 2
	ReasonInvalidItem         Reason = 3
	ReasonDuplicateItem       Reason = 4
	ReasonConflictingItem     Reason = 5
	ReasonInvalidDigest       Reason = 6
	ReasonStaleRegistry       Reason = 7
	ReasonDuplicateProjection Reason = 8
	ReasonUnknownProjection   Reason = 9
	ReasonMissingProjection   Reason = 10
	ReasonStaleSource         Reason = 11
	ReasonMissingSource       Reason = 12
)

type EnforcementEffect uint8

const EnforcementEffectNoEffect EnforcementEffect = 0

type CoverageResult struct {
	RegistryDigest       Digest
	RegistrySourceDigest Digest
	CurrentSourceDigest  Digest
	Decision             Decision
	Reason               Reason
	Faults               []Reason
	FullSuiteRequired    bool
	ExecutionAuthorized  bool
	EnforcementEffect    EnforcementEffect
	RegisteredCount      uint64
	ObservedCount        uint64
	CoveredCount         uint64
	ResultDigest         Digest
	ReplayDigest         Digest
}
