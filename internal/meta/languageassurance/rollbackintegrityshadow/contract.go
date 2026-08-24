package rollbackintegrityshadow

const (
	Schema              = "gooo/rollback-integrity-shadow/v1"
	MetricID            = "gooo.metric.operation.rollback-integrity.v1"
	MetaOperation       = "verify-rollback-integrity"
	PredecessorSHA      = "4b8758428951e3e208b42cb568cc208cbe83dea9"
	AssuranceDigest     = "sha256:4641750a0d3299524088d4b56678d594ee2873c0a13eaea25332709cfd9471ac"
	DecisionShadowPass  = "SHADOW_PASS"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"
	ReasonShadowPass    = "ROLLBACK_INTEGRITY_SHADOW_PROVEN"
	ReasonUnavailable   = "ROLLBACK_INTEGRITY_ASSURANCE_UNAVAILABLE"
	ReasonDigest        = "ROLLBACK_INTEGRITY_ASSURANCE_DIGEST_MISMATCH"
	ReasonBaseline      = "ROLLBACK_INTEGRITY_BASELINE_MISMATCH"
	ReasonSuite         = "ROLLBACK_INTEGRITY_META_SUITE_MISMATCH"
	EnforcementNoEffect = "NO_EFFECT"
	caseTotal           = 7
)

type assuranceSummary struct {
	DenominatorTotal      int `json:"denominator_total"`
	Operating             int `json:"operating"`
	NotImplemented        int `json:"not_implemented"`
	CoverageBPS           int `json:"implementation_coverage_bps"`
	UnknownTopDecisions   int `json:"unknown_top_decisions"`
	UnresolvedIndicators  int `json:"unresolved_indicators"`
	ViolatedGuardrails     int `json:"violated_guardrails"`
	RepositoryWrites      int `json:"repository_writes"`
}

type assuranceObligation struct {
	MetricID, Status, Resolution, MetaOperation string
}

type assuranceReport struct {
	Schema, SubjectSHA, AssuranceDecision, CandidateDecision string
	Summary                                                  assuranceSummary
	Obligations                                              []assuranceObligation
}
