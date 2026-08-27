package causalci

const (
	ObservationSchema = "gooo/causal-ci-selection-observation/v2"
	ReceiptSchema     = "gooo/causal-ci-selection-receipt/v2"
	ReportSchema      = "gooo/causal-ci-selection-intervention/v1"
	ReceiptScope      = "READ_ONLY_CAUSAL_SELECTION"

	ConformancePass       = "PASS"
	ConformanceFailClosed = "FAIL_CLOSED"

	ResolutionSelected   = "SELECTED"
	ResolutionUnknown    = "UNKNOWN"
	ResolutionFailClosed = "FAIL_CLOSED"

	PlanSelective = "SELECTIVE_PLAN"
	PlanFull      = "DESCEND_TO_FULL_SUITE"
	PlanNone      = "NO_PLAN"

	ClaimOpen       = "OPEN"
	ClaimDischarged = "DISCHARGED"
	ClaimRefuted    = "REFUTED"

	ProofCausalPath  = "CLAIM_IMPACT_REASON"
	ProofFullDescent = "FULL_SUITE_FALLBACK"
	ProofNone        = "NO_PLAN"

	ReasonMalformedObservation = "MALFORMED_RAW_OBSERVATION"
	ReasonMalformedPolicy      = "MALFORMED_GOOO_POLICY"
	ReasonUnknownSubject       = "SOURCE_NOT_BOUND_TO_POLICY"
	ReasonMissingRoute         = "CAUSAL_ROUTE_NOT_RECONSTRUCTED"
	ReasonContradictoryPolicy  = "CONTRADICTORY_POLICY_PATH"
	ReasonCompletePath         = "COMPLETE_CLAIM_SURFACE_CHECK_PATH"
	ReasonClaimDischarged      = "COMPLETE_PATH_OBSERVED"
	ReasonClaimLowered         = "UNKNOWN_PATH_PRESERVED_OPEN"
	ReasonClaimRefuted         = "EXPLICIT_POLICY_CONTRADICTION"

	FixedCheckDenominator     = 6
	FixedIndicatorDenominator = 6
)

var fixedCheckIDs = [...]string{"gofmt", "go-vet", "go-test", "go-test-race", "semantic-conformance", "ci-policy"}

var indicatorIDs = [...]string{
	"semantic-policy-derived",
	"changed-file-observation-bound",
	"unknown-descends-to-full",
	"claim-transition-append-only",
	"isolation-derived-read-only",
	"plan-only-no-execution-claim",
}

const (
	programChangedFileToClaim = "causal-ci.changed-file-to-claim/v2"
	programClaimToSurface     = "causal-ci.claim-to-surface/v2"
	programSurfaceToCheck     = "causal-ci.surface-to-check/v2"
	programPriorClaimState    = "causal-ci.prior-claim-state/open/v2"
	programDischarge          = "causal-ci.claim-transition/discharged/v2"
	programLowerResolution    = "causal-ci.claim-transition/open-lower-resolution/v2"
	programRefute             = "causal-ci.claim-transition/refuted/v2"
)

const (
	stageConformance    = "CONFORMANCE"
	stageSubject        = "CAUSAL_SELECTION"
	stageClaimLedger    = "CLAIM_LEDGER"
	stepParse           = "parse"
	stepLower           = "lower"
	stepValidatePolicy  = "validate-policy"
	stepObserveSubject  = "observe-subject"
	stepSelectChecks    = "select-checks"
	stepDescendFull     = "descend-full-suite"
	stepClaimTransition = "append-transition"
)

func fixedCheckIDSet() map[string]struct{} {
	result := make(map[string]struct{}, len(fixedCheckIDs))
	for _, id := range fixedCheckIDs {
		result[id] = struct{}{}
	}
	return result
}
