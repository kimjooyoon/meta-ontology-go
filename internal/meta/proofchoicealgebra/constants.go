package proofchoicealgebra

const (
	InputSchema             = "gooo/proof-choice-input/v3"
	ReceiptSchema           = "gooo/proof-choice-algebra-receipt/v3"
	FixedDenom              = 3
	ClaimKind               = "claim"
	MetricKind              = "metric"
	CompositionKind         = "composition"
	Pass                    = "PASS"
	FailClosed              = "FAIL_CLOSED"
	Exact                   = "EXACT"
	Lower                   = "LOWER_RESOLUTION"
	UnknownState            = "UNKNOWN"
	InsufficientState       = "INSUFFICIENT"
	Foundation        Route = "FOUNDATION"
	Coherence         Route = "COHERENCE"
	Regression        Route = "REGRESSION"
)

type Route string

func (r Route) Valid() bool {
	return r == Foundation || r == Coherence || r == Regression
}
