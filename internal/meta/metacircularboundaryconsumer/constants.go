package metacircularboundaryconsumer

const (
	reportSchema           = "gooo/meta-circular-boundary-report/v1"
	receiptSchema          = "gooo/meta-circular-boundary-receipt/v1"
	scope                  = "READ_ONLY_META_CIRCULAR_BOUNDARY"
	metaOperationID        = "meta-circular-boundary.evaluate"
	caseTotal              = 4
	indicatorTotal         = 10
	decisionPass           = "PASS"
	resolutionExact        = "EXACT"
	reasonContract         = "META_CIRCULAR_BOUNDARY_PROVEN"
	reasonMismatch         = "INDEPENDENT_JUDGE_MISMATCH"
	descriptionBound       = "BOUND"
	authorizationGrant     = "GRANTED"
	authorizationDeny      = "DENIED"
	executionAllowed       = "ALLOWED"
	executionBlocked       = "BLOCKED"
	scopeReadOnly          = "READ_ONLY"
	scopeWrite             = "WRITE"
	reasonDescription      = "DESCRIPTION_IS_NOT_AUTHORITY"
	reasonExplicit         = "EXPLICIT_READ_ONLY_CAPABILITY_ACCEPTED"
	reasonForged           = "FORGED_AUTHORIZATION_REJECTED"
	reasonOutOfScope       = "CAPABILITY_SCOPE_EXCEEDS_READ_ONLY"
	reasonSource           = "SELF_DESCRIPTION_BINDING_UNKNOWN"
	proofFoundation        = "FOUNDATION"
	proofCoherence         = "COHERENCE"
	proofRegression        = "REGRESSION"
	relationEqual          = "EQUAL"
	relationLessOrEqual    = "LESS_OR_EQUAL"
	relationGreaterOrEqual = "GREATER_OR_EQUAL"
	metaValue              = "DESCRIPTION_AUTHORIZATION_EXECUTION_ARE_DISTINCT"
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
}

func notClaimed() []string {
	return []string{
		"a self-hosting evaluator", "cryptographic unforgeability of the fixture handle",
		"general capability confinement", "arbitrary Gooo execution", "repository mutation or promotion",
	}
}
