package phaseseparation

const (
	Schema                   = "gooo/meta.phase-separation-witness/v2"
	DecisionPass             = "PASS"
	DecisionUnknown          = "UNKNOWN"
	ResolutionExact          = "EXACT"
	ResolutionLower          = "LOWER_RESOLUTION"
	ReasonExact              = "PHASE_SEPARATION_WITNESS_EXACT"
	ReasonUnknownSource      = "UNKNOWN_SOURCE_CONTRACT"
	ReasonUnknownContract    = ReasonUnknownSource
	ReasonUnknownSyntax      = "UNKNOWN_SOURCE_SYNTAX"
	ReasonUnknownCI          = "UNKNOWN_CI_SNAPSHOT"
	Toolchain                = "go1.27.0"
	ProducerID               = "source-authority"
	ConsumerID               = "independent-adjudicator"
	MetaOperationID          = "preserve-explicit-claim"
	ProofChoiceID            = "boundary-receipt"
	ExpectedSourceCases      = 6
	ExpectedCleanCases       = 1
	ExpectedLeakageCases     = 5
	ExpectedClaimTransitions = 2
	ExpectedIndicators       = 12
	ExpectedViews            = 3
	ExpectedProofs           = 3
	ExpectedSemanticChecks   = 3
	ExpectedProducerImports  = 0
)

const (
	PayloadClaim     = "claim"
	PayloadValue     = "value"
	PayloadAuthority = "authority"
	PayloadEvidence  = "evidence"

	StateOpen       = "OPEN"
	StateDischarged = "DISCHARGED"
	StateRefuted    = "REFUTED"
)

var expectedCaseKeys = []string{
	"clean",
	"value-leak",
	"authority-leak",
	"evidence-leak",
	"phase-skip",
	"phase-reverse",
}

var expectedLeakReasons = map[string]string{
	PayloadValue:     "VALUE_CROSSES_PHASE",
	PayloadAuthority: "AUTHORITY_CROSSES_PHASE",
	PayloadEvidence:  "EVIDENCE_CROSSES_PHASE",
	"phase-skip":     "PHASE_EDGE_SKIPS",
	"phase-reverse":  "PHASE_EDGE_REVERSES",
}

var phases = []string{"source", "expansion", "execution"}

var expectedLiteralClass = map[string]string{
	"source":    "declared",
	"expansion": "expanded",
	"execution": "observed",
}

const zeroDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
