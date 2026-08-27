package causalci

const (
	InputSchema   = "gooo/causal-ci-selection-input/v1"
	PolicySchema  = "gooo/causal-ci-selection-policy/v1"
	ReceiptSchema = "gooo/causal-ci-selection-receipt/v1"
	ReceiptScope  = "READ_ONLY_CAUSAL_SELECTION"

	DecisionPass         = "PASS"
	DecisionSelected     = "SELECTED"
	DecisionFullFallback = "FULL_FALLBACK"
	DecisionRejected     = "REJECTED"

	ResolutionSelective = "SELECTIVE_PLAN"
	ResolutionFull      = "DESCEND_TO_FULL_SUITE"
	ResolutionRejected  = "NO_PLAN"

	ReasonCompletePaths  = "COMPLETE_CLAIM_IMPACT_PATHS"
	ReasonUnknownPath    = "UNKNOWN_IMPACT_PATH"
	ReasonContradictory  = "CONTRADICTORY_IMPACT_PATH"
	ReasonUnregistered   = "UNREGISTERED_NODE"
	ReasonClaimBypass    = "CLAIM_BYPASS"
	ReasonNoRoute        = "NO_CLAUSAL_ROUTE"
	ReasonMalformedInput = "MALFORMED_INPUT"

	FixedCheckDenominator     = 6
	FixedIndicatorDenominator = 6
	ScenarioDenominator       = 3
)

var requiredCheckIDs = [...]string{
	"gofmt",
	"go-vet",
	"go-test",
	"go-test-race",
	"semantic-conformance",
	"ci-policy",
}

var indicatorIDs = [...]string{
	"claim-mediated-paths",
	"impact-path-reasons",
	"selection-choice-bound",
	"unknown-full-descent",
	"rejection-fail-closed",
	"persistent-claim-ledger",
}

const (
	proofCausalPath  = "CAUSAL_PATH"
	proofFullDescent = "FULL_SUITE_FALLBACK"
	proofNone        = "NO_PLAN"
)

func requiredScenarioIDs() []string {
	return []string{"selection", "full-fallback", "rejection"}
}
