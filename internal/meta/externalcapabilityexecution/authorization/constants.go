package authorization

const (
	EnvelopeSchema   = "gooo/external-capability-authorization-envelope/v1"
	FoundationSchema = "gooo/external-capability-policy-foundation/v1"
	ReceiptSchema    = "gooo/external-capability-authorization-receipt/v1"
	SuiteSchema      = "gooo/external-capability-authorization-suite/v1"

	MetricDenominator = 10
	SuiteDenominator  = 20

	DecisionAuthorized = "AUTHORIZED_SHADOW"
	DecisionDenied     = "DENIED"
	DecisionFailClosed = "FAIL_CLOSED"
	DecisionPass       = "PASS"
	ResolutionExact    = "EXACT"
	ResolutionUnknown  = "UNKNOWN"
	EffectNoEffect     = "NO_EFFECT"
	EffectBlock        = "BLOCK"
	StatusSatisfied    = "SATISFIED"
	StatusUnsatisfied  = "UNSATISFIED"
	StatusUnknown      = "UNKNOWN"

	ExpectedIssuer          = "github-actions:transformation-effect/external-conformance-go127"
	ExpectedOperation       = "gomacro.evaluate-and-generate"
	ExpectedScope           = "embedded-eval,interpreted-function,ast-macro"
	ExpectedDefaultDecision = "DENY"
)
