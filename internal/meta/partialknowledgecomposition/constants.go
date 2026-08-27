package partialknowledgecomposition

const (
	Schema        = "gooo/meta/partial-knowledge-composition-receipt/v1"
	FixtureSchema = "gooo/meta/partial-knowledge-composition-fixture/v1"
	Producer      = "partial-knowledge-producer"
	Consumer      = "partial-knowledge-composition-consumer"
	MetaOperation = "compose-partial-knowledge"
	SourcePath    = "examples/partial-knowledge-composition/main.gooo"

	FixedDenominator = 5
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

func validState(value State) bool {
	return value == StateExact || value == StateDirectUnknown ||
		value == StateDependencyBlocked || value == StateInvariantOnly
}

func validResultState(value State) bool {
	return validState(value) || value == StateMixedUnresolved
}

func validProofChoice(value ProofChoice) bool {
	return value == ProofFoundation || value == ProofCoherence || value == ProofRegression
}
