package sourceauthorityactivation

const (
	Schema                = "gooo/source-authority-activation/v1"
	DenominatorID         = "gooo/source-authority-activation-denominator/v1"
	PredecessorSHA        = "240bb8019af2f5488701cd00797a4d3598bda213"
	MetricID              = "gooo.metric.semantic.source-backed-authority.v1"
	MetaOperation         = "bind-source-backed-authority"
	DecisionApplied       = "APPLIED"
	DecisionFailClosed    = "FAIL_CLOSED"
	ResolutionExact       = "EXACT"
	ResolutionUnknown     = "UNKNOWN"
	ResolutionInvariant   = "INVARIANT_ONLY"
	ReasonApplied         = "SOURCE_AUTHORITY_ACTIVATED"
	ReasonUnavailable     = "ACTIVATION_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch  = "ACTIVATION_CAPSULE_DIGEST_MISMATCH"
	ReasonEligibility     = "ACTIVATION_ELIGIBILITY_NOT_EXACT"
	EligibilityReportHash = "sha256:8e2cd3ae80878fdb1ceeb9085ecc9da796a06be7c7612e7e0edff83b602a92bb"
	AssuranceCapsuleHash  = "sha256:eb2196cffb17287c9dcda27ff2593b484265d326f998fbafff388513b66db9a9"
	UpstreamCapsuleHash   = "sha256:a9b4851d3c127d37af99a63ef5be6be649e6d9daa29abd1f29a8c096f1c3ab44"
	EligibilityCapsuleHash = "sha256:fcc0a54917796f94a8c1650704fc66f17f629023182e551f0f6cd253c863954c"
)

type CaseSpec struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
}

func Denominator() []CaseSpec {
	return []CaseSpec{
		{ID: "exact", ExpectedDecision: DecisionApplied, ExpectedResolution: ResolutionExact, ExpectedReason: ReasonApplied},
		{ID: "unavailable", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedReason: ReasonUnavailable},
		{ID: "digest-mismatch", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionInvariant, ExpectedReason: ReasonDigestMismatch},
	}
}
