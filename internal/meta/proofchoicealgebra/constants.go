package proofchoicealgebra

const (
	InputSchema           = "gooo/proof-choice-input/v2"
	ReceiptSchema         = "gooo/proof-choice-algebra-receipt/v2"
	FixedDenom            = 3
	ClaimKind             = "claim"
	MetricKind            = "metric"
	ObservationKind       = "observation"
	Pass                  = "PASS"
	FailClosed            = "FAIL_CLOSED"
	Exact                 = "EXACT"
	Lower                 = "LOWER_RESOLUTION"
	Unknown               = "UNKNOWN"
	Foundation      Route = "FOUNDATION"
	Coherence       Route = "COHERENCE"
	Regression      Route = "REGRESSION"
)

type Route string

func (r Route) Valid() bool {
	return r == Foundation || r == Coherence || r == Regression
}
