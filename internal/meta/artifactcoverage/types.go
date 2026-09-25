package artifactcoverage

const (
	Schema        = "gooo/meta-operation-artifact-program/v1"
	AuthorityPath = "examples/meta-operation-artifact-coverage/main.gooo"
)

type ProofChoice string

const (
	ProofFoundation ProofChoice = "foundation"
	ProofCoherence  ProofChoice = "coherence"
	ProofRegression ProofChoice = "regression"
)

type IndicatorClass string

const (
	ClassOutcome   IndicatorClass = "outcome"
	ClassDriver    IndicatorClass = "driver"
	ClassGuardrail IndicatorClass = "guardrail"
)

type Relation string

const (
	RelationGreaterOrEqual Relation = "greater_or_equal"
	RelationLessOrEqual    Relation = "less_or_equal"
)
