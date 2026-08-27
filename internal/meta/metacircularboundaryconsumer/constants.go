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
	executionUnknown          = "UNKNOWN"
	authorizationUnknown      = "UNKNOWN"
	scopeReadOnly             = "READ_ONLY"
	scopeWrite                = "WRITE"
	reasonDescription         = "DESCRIPTION_IS_NOT_AUTHORITY"
	reasonExplicit            = "EXPLICIT_READ_ONLY_CAPABILITY_ACCEPTED"
	reasonForged              = "FORGED_AUTHORIZATION_REJECTED"
	reasonOutOfScope          = "CAPABILITY_SCOPE_EXCEEDS_READ_ONLY"
	reasonGrantUnknown        = "EXTERNAL_GRANT_UNKNOWN"
	reasonGrantDenied         = "EXTERNAL_GRANT_DENIED_BY_POLICY"
	reasonExecutionUnknown    = "EXECUTION_EVIDENCE_UNKNOWN"
	reasonExecutionInvalid    = "EXECUTION_EVIDENCE_INVALID"
	reasonGraphUnknown        = "TYPED_META_OPERATION_GRAPH_UNKNOWN"
	reasonDescriptionForgery  = "DESCRIPTION_AUTHORITY_CLAIM_BLOCKED"
	reasonCasePredicate       = "CASE_PREDICATE_MISMATCH"
	reasonEffectUnknown       = "WORKSPACE_EFFECT_EVIDENCE_UNKNOWN"
	reasonReplayUnknown       = "REPLAY_EVIDENCE_UNKNOWN"
	grantDecision             = "GRANT"
	grantDeny                 = "DENY"
	requestNone               = "NONE"
	requestReadOnly           = "READ_ONLY"
	authorityDenied           = "DENIED"
	authorityGranted          = "GRANTED"
	authorityUnknown          = "UNKNOWN"
	graphSchema               = "gooo/meta-circular-boundary-graph/v1"
	grantSchema               = "gooo/meta-circular-boundary-external-grant/v1"
	effectSchema              = "gooo/meta-circular-boundary-effect/v1"
	replaySchema              = "gooo/meta-circular-boundary-replay/v1"
	executionSchema           = "gooo/meta-circular-boundary-execution/v1"
	judgeReceiptSchema        = "gooo/meta-circular-boundary-judge-receipt/v1"
	grantProducer             = "external-authority-fixture"
	effectProducer            = "ci-workspace-observer"
	replayProducer            = "ci-replay-observer"
	executionProducer         = "metacircularboundary.ExecuteReadOnlyMetaOperation"
	predicateDescriptionOnly  = "DESCRIPTION_REQUEST_WITHOUT_EXTERNAL_GRANT"
	predicateExplicitGrant    = "EXPLICIT_EXTERNAL_READ_ONLY_GRANT"
	predicateForgedGrant      = "EXTERNAL_GRANT_HANDLE_OR_ISSUER_FORGED"
	predicateOutOfScopeGrant  = "EXTERNAL_GRANT_SCOPE_OUT_OF_BOUNDS"
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

func notClaimed() []string {
	return []string{
		"a self-hosting evaluator", "cryptographic unforgeability of the fixture handle",
		"general capability confinement", "arbitrary Gooo execution", "repository mutation or promotion",
	}
}
