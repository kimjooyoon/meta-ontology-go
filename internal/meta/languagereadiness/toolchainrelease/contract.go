package toolchainrelease

const (
	PlatformReceiptSchema = "gooo/toolchain-release-platform-receipt/v1"
	CorpusSchema          = "gooo/toolchain-cross-platform-release-corpus/v1"
	ReportSchema          = "gooo/toolchain-cross-platform-release-report/v1"
	MetaOperation         = "assemble-exact-cross-platform-release"
	ExpectedToolchain     = "go1.27.0"
	DecisionPass          = "PASS"
	DecisionFailClosed    = "FAIL_CLOSED"
	ResolutionExact       = "EXACT"
	ResolutionInvariant   = "INVARIANT"
	CaseSatisfied         = "SATISFIED"
	CaseNotSatisfied      = "NOT_SATISFIED"
	TargetCount           = 3
	CaseCount             = 20
	OutcomeCount          = 3
	DriverCount           = 16
	GuardrailCount        = 20
	IndicatorCount        = OutcomeCount + DriverCount + GuardrailCount
)

const (
	metricProducer = "toolchainrelease.Evaluate"
	metricConsumer = "self-improvement-cycle"
)
