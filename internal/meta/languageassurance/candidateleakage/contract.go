package candidateleakage

const (
	InputSchema       = "gooo/candidate-leakage-input/v1"
	ReportSchema      = "gooo/candidate-leakage-report/v1"
	SuiteSchema       = "gooo/candidate-leakage-conformance/v1"
	DenominatorID     = "gooo/candidate-leakage-denominator/v1"
	MetricID          = "gooo.metric.semantic.candidate-leakage.v1"
	BoundaryOperation = "promote-candidate-state"

	DecisionPass        = "PASS"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNone          = "NO_EFFECT"
	EffectBlock         = "BLOCK"

	CandidateAllowLimited  = "ALLOW_LIMITED"
	CandidateBlock         = "BLOCK"
	PromotionAuthorized    = "AUTHORIZED"
	PromotionDenied        = "DENIED"
	PromotionFailClosed    = "FAIL_CLOSED"
	OfficialOperating      = "OPERATING"
	OfficialNotImplemented = "NOT_IMPLEMENTED"
	OfficialAllow          = "ALLOW"
	OfficialPass           = "PASS"
	OfficialFixedPoint     = "FIXED_POINT"
	OfficialAuthorized     = "AUTHORIZED"
	OfficialBlock          = "BLOCK"
	OfficialFailClosed     = "FAIL_CLOSED"
	OfficialResolutionNone = "NONE"

	ReasonCandidateIsolated        = "CANDIDATE_REMAINS_NON_AUTHORITATIVE"
	ReasonExactPromotionBound      = "EXACT_PROMOTION_BOUND"
	ReasonLeakageDetected          = "CANDIDATE_LEAKAGE_DETECTED"
	ReasonSubjectBindingMismatch   = "CANDIDATE_SUBJECT_BINDING_UNKNOWN"
	ReasonDigestBindingMismatch    = "CANDIDATE_DIGEST_BINDING_UNKNOWN"
	ReasonOperationBindingMismatch = "CANDIDATE_OPERATION_BINDING_UNKNOWN"
	ReasonDecisionUnknown          = "CANDIDATE_BOUNDARY_DECISION_UNKNOWN"
	ReasonResolutionUnknown        = "CANDIDATE_BOUNDARY_RESOLUTION_UNKNOWN"
)

func MetaOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "freeze-candidate-leakage-denominator", ProofChoice: "COHERENCE"},
		{ID: "observe-candidate-envelope", ProofChoice: "FOUNDATION"},
		{ID: "bind-promotion-authority", ProofChoice: "FOUNDATION"},
		{ID: "detect-candidate-leakage", ProofChoice: "COHERENCE"},
		{ID: "preserve-candidate-unknown", ProofChoice: "REGRESSION"},
		{ID: "deny-shadow-promotion-credit", ProofChoice: "FOUNDATION"},
	}
}
