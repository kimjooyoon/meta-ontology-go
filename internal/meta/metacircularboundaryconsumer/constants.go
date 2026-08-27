package metacircularboundaryconsumer

const (
	reportSchema              = "gooo/meta-circular-boundary-report/v1"
	receiptSchema             = "gooo/meta-circular-boundary-receipt/v1"
	scope                     = "READ_ONLY_META_CIRCULAR_BOUNDARY"
	metaOperationID           = "meta-circular-boundary.evaluate"
	caseTotal                 = 4
	indicatorTotal            = 10
	decisionPass              = "PASS"
	decisionFailClosed        = "FAIL_CLOSED"
	decisionOpen              = "OPEN"
	decisionRefuted           = "REFUTED"
	resolutionExact           = "EXACT"
	resolutionLower           = "LOWER_RESOLUTION"
	reasonContract            = "META_CIRCULAR_BOUNDARY_PROVEN"
	reasonContractUnsatisfied = "META_CIRCULAR_BOUNDARY_CONTRACT_NOT_SATISFIED"
	reasonMismatch            = "INDEPENDENT_JUDGE_MISMATCH"
	reasonSource              = "SELF_DESCRIPTION_BINDING_UNKNOWN"
	reasonCaseData            = "CASE_COMPUTATION_DATA_UNKNOWN"
	reasonContradictory       = "CONTRADICTORY_CAPABILITY_EVIDENCE"
	descriptionBound          = "BOUND"
	authorizationGrant        = "GRANTED"
	authorizationDeny         = "DENIED"
	executionAllowed          = "ALLOWED"
	executionBlocked          = "BLOCKED"
	scopeReadOnly             = "READ_ONLY"
	scopeWrite                = "WRITE"
	reasonDescription         = "DESCRIPTION_IS_NOT_AUTHORITY"
	reasonExplicit            = "EXPLICIT_READ_ONLY_CAPABILITY_ACCEPTED"
	reasonForged              = "FORGED_AUTHORIZATION_REJECTED"
	reasonOutOfScope          = "CAPABILITY_SCOPE_EXCEEDS_READ_ONLY"
	proofFoundation           = "FOUNDATION"
	proofCoherence            = "COHERENCE"
	proofRegression           = "REGRESSION"
	relationEqual             = "EQUAL"
	relationLessOrEqual       = "LESS_OR_EQUAL"
	relationGreaterOrEqual    = "GREATER_OR_EQUAL"
	metaValue                 = "DESCRIPTION_AUTHORIZATION_EXECUTION_ARE_DISTINCT"
)

var requiredEntities = []string{
	"MetaOperation",
	"SelfDescription",
	"ReadOnlyCapability",
	"ExecutionReceipt",
	"ForgedCapability",
}

var requiredActivities = []string{
	"DescribeMetaOperation",
	"GrantReadOnlyMetaCapability",
	"ExecuteMetaOperation",
	"InspectForgedCapability",
	"DescriptionOnlyAttempt",
	"ExplicitReadOnlyCapabilityAttempt",
	"ForgedCapabilityAttempt",
	"WriteCapabilityAttempt",
}

func notClaimed() []string {
	return []string{
		"a self-hosting evaluator", "cryptographic unforgeability of the fixture handle",
		"general capability confinement", "arbitrary Gooo execution", "repository mutation or promotion",
	}
}
