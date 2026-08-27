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
	DecisionOpen       = "OPEN"
	DecisionRefuted    = "REFUTED"
	ResolutionExact    = "EXACT"
	ResolutionLower    = "LOWER_RESOLUTION"

	DescriptionBound     = "BOUND"
	AuthorizationGranted = "GRANTED"
	AuthorizationDenied  = "DENIED"
	ExecutionAllowed     = "ALLOWED"
	ExecutionBlocked     = "BLOCKED"
	ExecutionUnknown     = "UNKNOWN"
	AuthorizationUnknown = "UNKNOWN"

	ScopeReadOnly = "READ_ONLY"
	ScopeWrite    = "WRITE"

	ReasonDescriptionOnly       = "DESCRIPTION_IS_NOT_AUTHORITY"
	ReasonExplicitCapability    = "EXPLICIT_READ_ONLY_CAPABILITY_ACCEPTED"
	ReasonForgedCapability      = "FORGED_AUTHORIZATION_REJECTED"
	ReasonOutOfScopeCapability  = "CAPABILITY_SCOPE_EXCEEDS_READ_ONLY"
	ReasonSourceBindingUnknown  = "SELF_DESCRIPTION_BINDING_UNKNOWN"
	ReasonCaseDataUnknown       = "CASE_COMPUTATION_DATA_UNKNOWN"
	ReasonContradictory         = "CONTRADICTORY_CAPABILITY_EVIDENCE"
	ReasonReportMismatch        = "INDEPENDENT_JUDGE_MISMATCH"
	ReasonContractSatisfied     = "META_CIRCULAR_BOUNDARY_PROVEN"
	ReasonContractUnsatisfied   = "META_CIRCULAR_BOUNDARY_CONTRACT_NOT_SATISFIED"
	ReasonReplayMismatch        = "INDEPENDENT_REPLAY_MISMATCH"
	ReasonGrantUnknown          = "EXTERNAL_GRANT_UNKNOWN"
	ReasonGrantDenied           = "EXTERNAL_GRANT_DENIED_BY_POLICY"
	ReasonExecutionUnknown      = "EXECUTION_EVIDENCE_UNKNOWN"
	ReasonExecutionInvalid      = "EXECUTION_EVIDENCE_INVALID"
	ReasonGraphUnknown          = "TYPED_META_OPERATION_GRAPH_UNKNOWN"
	ReasonUnauthorizedArtifact  = "UNAUTHORIZED_EXECUTION_ARTIFACT"
	ReasonDescriptionForgery    = "DESCRIPTION_AUTHORITY_CLAIM_BLOCKED"
	ReasonEffectUnknown         = "WORKSPACE_EFFECT_EVIDENCE_UNKNOWN"
	ReasonReplayUnknown         = "REPLAY_EVIDENCE_UNKNOWN"
	ReasonCasePredicateMismatch = "CASE_PREDICATE_MISMATCH"

	ProofFoundation = "FOUNDATION"
	ProofCoherence  = "COHERENCE"
	ProofRegression = "REGRESSION"

	RelationEqual          = "EQUAL"
	RelationLessOrEqual    = "LESS_OR_EQUAL"
	RelationGreaterOrEqual = "GREATER_OR_EQUAL"

	GraphSchema              = "gooo/meta-circular-boundary-graph/v1"
	GrantSchema              = "gooo/meta-circular-boundary-external-grant/v1"
	EffectSchema             = "gooo/meta-circular-boundary-effect/v1"
	ReplaySchema             = "gooo/meta-circular-boundary-replay/v1"
	ExecutionSchema          = "gooo/meta-circular-boundary-execution/v1"
	JudgeReceiptSchema       = "gooo/meta-circular-boundary-judge-receipt/v1"
	GrantProducer            = "external-authority-fixture"
	EffectProducer           = "ci-workspace-observer"
	ReplayProducer           = "ci-replay-observer"
	ExecutionProducer        = "metacircularboundary.ExecuteReadOnlyMetaOperation"
	GrantDecision            = "GRANT"
	GrantDeny                = "DENY"
	RequestNone              = "NONE"
	RequestReadOnly          = "READ_ONLY"
	AuthorityDenied          = "DENIED"
	AuthorityGranted         = "GRANTED"
	AuthorityUnknown         = "UNKNOWN"
	PredicateDescriptionOnly = "DESCRIPTION_REQUEST_WITHOUT_EXTERNAL_GRANT"
	PredicateExplicitGrant   = "EXPLICIT_EXTERNAL_READ_ONLY_GRANT"
	PredicateForgedGrant     = "EXTERNAL_GRANT_HANDLE_OR_ISSUER_FORGED"
	PredicateOutOfScopeGrant = "EXTERNAL_GRANT_SCOPE_OUT_OF_BOUNDS"
)

var requiredEntities = []string{
	"MetaOperation",
	"SelfDescription",
	"ReadOnlyCapability",
	"ExecutionReceipt",
	"ForgedCapability",
}
