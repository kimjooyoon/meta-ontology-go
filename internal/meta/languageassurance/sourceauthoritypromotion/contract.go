package sourceauthoritypromotion

const (
	Schema                  = "gooo/source-authority-promotion-eligibility/v1"
	EligibilityDenominator  = "gooo/source-authority-promotion-eligibility-denominator/v1"
	AssuranceSchema         = "gooo/language-assurance-report/v1"
	AssuranceDenominator    = "gooo/language-assurance-denominator/v1"
	AssuranceDigest         = "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02"
	UpstreamSchema          = "gooo/upstream-source-conformance/v1"
	UpstreamDenominator     = "gooo/upstream-source-conformance-denominator/v1"
	UpstreamDigest          = "sha256:c5815a32fe85302ccdedbefc716f909d1001c732e2f79c59d794e50a81944783"
	SourceMetric            = "gooo.metric.semantic.source-backed-authority.v1"
	SourceOperation         = "bind-source-backed-authority"
	DecisionEligible        = "ELIGIBLE"
	DecisionBlock           = "BLOCK"
	ResolutionExact         = "EXACT"
	ResolutionInvariantOnly = "INVARIANT_ONLY"
	EnforcementNoEffect     = "NO_EFFECT"
	ReasonEligible          = "SOURCE_AUTHORITY_PROMOTION_ELIGIBLE"
	ReasonMalformed         = "PROMOTION_EVIDENCE_MALFORMED"
	ReasonSubjectMismatch   = "PROMOTION_SUBJECT_MISMATCH"
	ReasonAssuranceBoundary = "ASSURANCE_BASELINE_NOT_EXACT"
	ReasonBaselineState     = "SOURCE_AUTHORITY_BASELINE_NOT_NOT_IMPLEMENTED"
	ReasonUpstreamNotExact  = "UPSTREAM_SOURCE_EVIDENCE_NOT_EXACT"
)

type CaseSpec struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
}

func Denominator() []CaseSpec {
	return []CaseSpec{
		{ID: "eligible", ExpectedDecision: DecisionEligible, ExpectedResolution: ResolutionExact, ExpectedReason: ReasonEligible},
		{ID: "upstream-not-exact", ExpectedDecision: DecisionBlock, ExpectedResolution: ResolutionInvariantOnly, ExpectedReason: ReasonUpstreamNotExact},
		{ID: "baseline-already-operating", ExpectedDecision: DecisionBlock, ExpectedResolution: ResolutionInvariantOnly, ExpectedReason: ReasonBaselineState},
	}
}
