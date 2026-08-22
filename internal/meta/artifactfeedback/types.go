package artifactfeedback

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

type MetaOperation struct {
	ID          string      `json:"id"`
	Activity    string      `json:"activity"`
	ProofChoice ProofChoice `json:"proof_choice"`
}

type Indicator struct {
	MetricID      string         `json:"metric_id"`
	Class         IndicatorClass `json:"class"`
	Target        int            `json:"target"`
	Unit          string         `json:"unit"`
	Relation      Relation       `json:"relation"`
	ProofChoice   ProofChoice    `json:"proof_choice"`
	Producer      string         `json:"producer"`
	Consumer      string         `json:"consumer"`
	MetaOperation string         `json:"meta_operation"`
	Activity      string         `json:"activity"`
}

type Program struct {
	Schema         string          `json:"schema"`
	Authority      string          `json:"authority"`
	MetaOperations []MetaOperation `json:"meta_operations"`
	Indicators     []Indicator     `json:"indicators"`
}

