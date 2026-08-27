package proofchoicealgebra

const (
	Schema                  = "gooo/proof-choice-algebra-receipt/v1"
	FixedDenominator        = 3
	Claim            Kind   = "CLAIM"
	Metric           Kind   = "METRIC"
	Pass                    = "PASS"
	FailClosed              = "FAIL_CLOSED"
	Exact                   = "EXACT"
	Foundation       Choice = "FOUNDATION"
	Coherence        Choice = "COHERENCE"
	Regression       Choice = "REGRESSION"
)

type Kind string
type Choice string

func (c Choice) Valid() bool {
	return c == Foundation || c == Coherence || c == Regression
}

func (c Choice) String() string { return string(c) }
