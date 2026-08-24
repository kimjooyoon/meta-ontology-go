package externalcapabilityexecution

const (
	ObservationSchema = "gooo/external-capability-observation/v1"
	ReportSchema      = "gooo/external-capability-report/v1"
	SuiteSchema       = "gooo/external-capability-conformance/v1"

	ExpectedRepository = "https://github.com/cosmos72/gomacro"
	ExpectedCommit     = "cf0d4bf32da393dbda97e3572f216731013ffa55"
	ExpectedTree       = "8cc240a53dd29432ad83620b20fd8a0a05674c6d"
	ExpectedGoVersion  = "go1.27.0"

	MetricDenominator = 10
	SuiteDenominator  = 15

	DecisionExecutable = "CAPABILITY_EXECUTABLE"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
	ResolutionUnknown  = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNoEffect     = "NO_EFFECT"
	EffectBlock        = "BLOCK"
	StatusSatisfied    = "SATISFIED"
	StatusUnsatisfied  = "UNSATISFIED"
	StatusUnknown      = "UNKNOWN"

	ReasonExecutable = "CAPABILITY_EXECUTION_EXACT_PARENT_FAIL_CLOSED"
	ReasonUnknown    = "CAPABILITY_EXECUTION_EVIDENCE_UNKNOWN"
	ReasonInvariant  = "CAPABILITY_EXECUTION_INVARIANT_VIOLATED"
)

var caseIDs = []string{
	"exact", "observation-unavailable", "run-status-unknown",
	"parent-decision-unknown", "repository-mismatch", "commit-mismatch",
	"tree-mismatch", "toolchain-mismatch", "arithmetic-mismatch",
	"function-mismatch", "macro-mismatch", "replay-mismatch",
	"project-write", "external-write", "parent-promoted",
}
