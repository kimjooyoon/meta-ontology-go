package metacircularboundary

const (
	ReportSchema       = "gooo/meta-circular-boundary-report/v1"
	DenominatorID      = "gooo/meta-circular-boundary-denominator/v1"
	Scope              = "READ_ONLY_META_CIRCULAR_BOUNDARY"
	MetaOperationID    = "meta-circular-boundary.evaluate"
	ExpectedSourcePath = "examples/meta-circular-boundary/main.gooo"
	CaseTotal          = 4
	IndicatorTotal     = 10

	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
	ResolutionLower    = "LOWER_RESOLUTION"

	DescriptionBound     = "BOUND"
	AuthorizationGranted = "GRANTED"
	AuthorizationDenied  = "DENIED"
	ExecutionAllowed     = "ALLOWED"
	ExecutionBlocked     = "BLOCKED"

	ScopeReadOnly = "READ_ONLY"
	ScopeWrite    = "WRITE"

	ReasonDescriptionOnly      = "DESCRIPTION_IS_NOT_AUTHORITY"
	ReasonExplicitCapability   = "EXPLICIT_READ_ONLY_CAPABILITY_ACCEPTED"
	ReasonForgedCapability     = "FORGED_AUTHORIZATION_REJECTED"
	ReasonOutOfScopeCapability = "CAPABILITY_SCOPE_EXCEEDS_READ_ONLY"
	ReasonSourceBindingUnknown = "SELF_DESCRIPTION_BINDING_UNKNOWN"
	ReasonReportMismatch       = "INDEPENDENT_JUDGE_MISMATCH"
	ReasonContractSatisfied    = "META_CIRCULAR_BOUNDARY_PROVEN"
	ReasonReplayMismatch       = "INDEPENDENT_REPLAY_MISMATCH"

	ProofFoundation = "FOUNDATION"
	ProofCoherence  = "COHERENCE"
	ProofRegression = "REGRESSION"

	RelationEqual          = "EQUAL"
	RelationLessOrEqual    = "LESS_OR_EQUAL"
	RelationGreaterOrEqual = "GREATER_OR_EQUAL"
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
