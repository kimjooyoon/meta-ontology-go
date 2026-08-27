package phaseseparation

const (
	Schema                  = "gooo/meta.phase-separation-witness/v1"
	DecisionPass             = "PASS"
	DecisionUnknown          = "UNKNOWN"
	ResolutionExact          = "EXACT"
	ResolutionLower          = "LOWER_RESOLUTION"
	ReasonExact              = "PHASE_SEPARATION_WITNESS_EXACT"
	ReasonUnknownSource      = "UNKNOWN_SOURCE_SYNTAX"
	ReasonUnknownContract    = "UNKNOWN_SOURCE_CONTRACT"
	Toolchain                = "go1.27.0"
	ExpectedCleanCases       = 1
	ExpectedLeakageCases     = 5
	ExpectedClaimTransitions = 2
	ExpectedIndicators       = 12
	ExpectedViews            = 3
	ExpectedProofs           = 3
)

var expectedLeakReasons = map[string]string{
	"value-leak":     "VALUE_CROSSES_PHASE",
	"authority-leak": "AUTHORITY_CROSSES_PHASE",
	"evidence-leak":  "EVIDENCE_CROSSES_PHASE",
	"phase-skip":     "PHASE_EDGE_SKIPS",
	"phase-reverse":  "PHASE_EDGE_REVERSES",
}

var phases = []string{"source", "expansion", "execution"}
