package languagepackageruntime

const (
	RegistrySchema     = "gooo/language-package-runtime-cases/v1"
	ReportSchema       = "gooo/language-package-runtime-report/v1"
	RegistryVersion    = "2026-08-23"
	FixedTotal         = 18
	FixedPositive      = 10
	FixedGuardrails    = 8
	FixedIndicators    = 18
	ExpectedPackages   = 40
	ExpectedSources    = 50
	ExpectedImports    = 40
	ExpectedInits      = 40
	ExpectedEntries    = 10
	ExpectedSemantics  = 50
	ExpectedReplays    = 10
	ExpectedInvariants = 3
)

type Decision string
type Resolution string

const (
	DecisionPass    Decision   = "PASS"
	DecisionClosed  Decision   = "FAIL_CLOSED"
	ResolutionExact Resolution = "EXACT"
	ResolutionLower Resolution = "LOWER_RESOLUTION"
)
