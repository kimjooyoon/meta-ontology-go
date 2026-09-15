package toolchainformatfix

const (
	RegistrySchema     = "gooo/toolchain-format-fix-cases/v1"
	ReportSchema       = "gooo/toolchain-format-fix-report/v1"
	RegistryVersion    = "2026-08-23"
	FixedTotal         = 12
	FixedPositive      = 6
	FixedGuardrails    = 6
	FixedIndicators    = 18
	ExpectedRuns       = 24
	ExpectedStructured = 3
	ExpectedPlans      = 2
)

type Decision string
type Resolution string

const (
	DecisionPass    Decision   = "PASS"
	DecisionClosed  Decision   = "FAIL_CLOSED"
	ResolutionExact Resolution = "EXACT"
	ResolutionLower Resolution = "LOWER_RESOLUTION"
)
