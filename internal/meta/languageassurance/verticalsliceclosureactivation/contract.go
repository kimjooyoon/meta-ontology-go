package verticalsliceclosureactivation

const (
	Schema                   = "gooo/vertical-slice-closure-activation/v1"
	DenominatorID            = "gooo/vertical-slice-closure-activation-denominator/v1"
	PredecessorSHA           = "3920551d8db1226810832f6f924783b2fddf4ccd"
	MetricID                 = "gooo.metric.capability.vertical-slice-closure.v1"
	MetaOperation            = "close-vertical-slice"
	EligibilityMetaOperation = "qualify-vertical-slice-closure"
	DecisionApplied          = "APPLIED"
	DecisionFailClosed       = "FAIL_CLOSED"
	ResolutionExact          = "EXACT"
	ResolutionUnknown        = "UNKNOWN"
	ResolutionInvariant      = "INVARIANT_ONLY"
	EffectApply              = "APPLY_TRANSITION"
	EffectBlock              = "BLOCK"
	ReasonApplied            = "VERTICAL_SLICE_CLOSURE_ACTIVATED"
	ReasonUnavailable        = "VERTICAL_SLICE_ACTIVATION_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch     = "VERTICAL_SLICE_ACTIVATION_CAPSULE_DIGEST_MISMATCH"
	ReasonAssuranceMismatch  = "VERTICAL_SLICE_ACTIVATION_ASSURANCE_MISMATCH"
	ReasonEligibilityInvalid = "VERTICAL_SLICE_ACTIVATION_ELIGIBILITY_NOT_EXACT"
	ReasonEligibilityUnknown = "VERTICAL_SLICE_ACTIVATION_ELIGIBILITY_UNKNOWN"
)

const (
	AssuranceCapsuleHash   = "sha256:ba6d4e7bb8db487fbac31e5e8d34e9826f86bb20b04a34094c7cd269aedf17bc"
	EligibilityCapsuleHash = "sha256:5c53ad1ee6c436c7ed788a31cb9d270791ea4958ec1f50db08d5792cf501ed3b"
	AssuranceReportHash    = "sha256:8696b63d14ba5fafa7b8993d04c38f03a8f164bb2e4b4ab0f6e14ff8e9e63f73"
	EligibilityReportHash  = "sha256:ce9a0a7c48ab063a70559272b2a0ed164754968fa26346fb13ad07ff728f3b73"
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
		{ID: "eligibility-unknown", ExpectedDecision: DecisionFailClosed, ExpectedResolution: ResolutionUnknown, ExpectedReason: ReasonEligibilityUnknown},
	}
}
