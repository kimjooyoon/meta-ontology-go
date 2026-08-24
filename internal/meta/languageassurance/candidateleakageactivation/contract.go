package candidateleakageactivation

const (
	Schema                    = "gooo/candidate-leakage-activation/v1"
	DenominatorID             = "gooo/candidate-leakage-activation-denominator/v1"
	PredecessorSHA            = "308233198b91a47bdfc34016f73e44905e2be582"
	EvidenceSubjectSHA        = "466c12d94f657f2e04f355919788260c17faa678"
	MetricID                  = "gooo.metric.semantic.candidate-leakage.v1"
	MetaOperation             = "detect-candidate-leakage"
	EligibilityDenominatorID  = "gooo/candidate-leakage-eligibility-denominator/v1"
	DecisionApplied           = "APPLIED"
	DecisionFailClosed        = "FAIL_CLOSED"
	ResolutionExact           = "EXACT"
	ResolutionUnknown         = "UNKNOWN"
	ResolutionInvariant       = "INVARIANT_ONLY"
	ReasonApplied             = "CANDIDATE_LEAKAGE_ACTIVATED"
	ReasonUnavailable         = "CANDIDATE_LEAKAGE_ACTIVATION_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch      = "CANDIDATE_LEAKAGE_ACTIVATION_CAPSULE_DIGEST_MISMATCH"
	ReasonAssuranceMismatch   = "CANDIDATE_LEAKAGE_ACTIVATION_ASSURANCE_MISMATCH"
	ReasonEligibilityNotExact = "CANDIDATE_LEAKAGE_ACTIVATION_ELIGIBILITY_NOT_EXACT"
	AssuranceCapsuleHash      = "sha256:22cbebb4ddf03dbf114a8ce7b70deb6c74910a5bf1b4af21c863bd198ea8e741"
	EligibilityCapsuleHash    = "sha256:e816861985039c32ba1e064fbf5db537a3567f6982da496b9a2cd34058ad3a8b"
	EligibilityReportHash     = "sha256:4e9697727a34dae856a77d2b349a0992b1bea888be0cde7db2bcca8b73ddefc9"
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
