package toolchaincli

const (
	RegistrySchema      = "gooo/toolchain-cli-cases/v1"
	ReportSchema        = "gooo/toolchain-cli-report/v2"
	RegistryVersion     = "2026-08-23"
	FixedTotal          = 12
	FixedPositive       = 6
	FixedGuardrails     = 6
	FixedIndicators     = 18
	ExpectedRuns        = 24
	ExpectedCommands    = 17
	ExpectedStructured  = 3
	ExpectedLanguageOps = 4
)

type Decision string
type Resolution string

const (
	DecisionPass    Decision   = "PASS"
	DecisionClosed  Decision   = "FAIL_CLOSED"
	ResolutionExact Resolution = "EXACT"
	ResolutionLower Resolution = "LOWER_RESOLUTION"
)
