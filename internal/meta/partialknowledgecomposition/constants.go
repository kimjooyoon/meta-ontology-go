package partialknowledgecomposition

const (
	Schema                 = "gooo/meta/partial-knowledge-composition-receipt/v2"
	SourcePath             = "examples/partial-knowledge-composition/main.gooo"
	ObservationSchema      = "gooo.partial-knowledge.recipe/v3"
	RawEvidenceSchema      = "gooo/partial-knowledge/raw-evidence/v3"
	Producer               = "partial-knowledge-producer"
	Consumer               = "partial-knowledge-composition-consumer"
	MetaOperation          = "compose-partial-knowledge"
	DecisionCalculusProven = "CALCULUS_PROVEN"
	ResolutionCalculus     = "CALCULUS"
	SubjectResolution      = "PARTIAL_KNOWLEDGE"
	FixedDenominator       = 5
)

type State string

const (
	StateExact             State = "EXACT"
	StateDirectUnknown     State = "DIRECT_UNKNOWN"
	StateDependencyBlocked State = "DEPENDENCY_BLOCKED"
	StateInvariantOnly     State = "INVARIANT_ONLY"
	StateMixedUnresolved   State = "MIXED_UNRESOLVED"
)

type ProofChoice string

const (
	ProofFoundation ProofChoice = "FOUNDATION"
	ProofCoherence  ProofChoice = "COHERENCE"
	ProofRegression ProofChoice = "REGRESSION"
)

type InterventionMode string

const (
	InterventionNone        InterventionMode = "none"
	InterventionSemantic    InterventionMode = "semantic"
	InterventionCommentOnly InterventionMode = "comment-only"
)

var fixedCaseIDs = []string{
	"exact-pair", "direct-unknown", "dependency-blocked",
	"invariant-preservation", "mixed-unknown-and-blocked",
}

var fixedActivityNames = []string{
	"ObserveExactPair", "ObserveDirectUnknown", "ObserveDependencyBlock",
	"ObserveInvariant", "ObserveMixedUnresolved",
}

var fixedMetaOperations = []string{
	MetaOperation, MetaOperation, MetaOperation, "preserve-known-invariant", MetaOperation,
}

var fixedProofChoices = []ProofChoice{
	ProofCoherence, ProofFoundation, ProofCoherence, ProofFoundation, ProofRegression,
}

func validState(value State) bool {
	return value == StateExact || value == StateDirectUnknown ||
		value == StateDependencyBlocked || value == StateInvariantOnly ||
		value == StateMixedUnresolved
}

func validProofChoice(value ProofChoice) bool {
	return value == ProofFoundation || value == ProofCoherence || value == ProofRegression
}

func validIntervention(value InterventionMode) bool {
	return value == InterventionNone || value == InterventionSemantic || value == InterventionCommentOnly
}
