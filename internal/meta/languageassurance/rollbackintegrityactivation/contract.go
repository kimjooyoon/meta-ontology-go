package rollbackintegrityactivation

const (
	Schema                       = "gooo/rollback-integrity-activation/v1"
	DenominatorID                = "gooo/rollback-integrity-activation-denominator/v1"
	PredecessorSHA               = "0908088ab35447098671263ffa7290c0363f7404"
	EvidenceSubjectSHA           = "1f27d342faf7a435ca4c534e2a816f29befe21a4"
	MetricID                     = "gooo.metric.operation.rollback-integrity.v1"
	MetaOperation                = "verify-rollback-integrity"
	EligibilityDenominatorID     = "gooo/rollback-integrity-eligibility-denominator/v1"
	EligibilityDenominatorDigest = "sha256:de7279775cab0cba55c8331f14969dcd029702d888e3d1f11299eac027871b30"
	DecisionApplied              = "APPLIED"
	DecisionFailClosed           = "FAIL_CLOSED"
	ResolutionExact              = "EXACT"
	ResolutionUnknown            = "UNKNOWN"
	ResolutionInvariant          = "INVARIANT_ONLY"
	EffectApply                  = "APPLY_TRANSITION"
	EffectBlock                  = "BLOCK"
	ReasonApplied                = "ROLLBACK_INTEGRITY_ACTIVATED"
	ReasonUnavailable            = "ROLLBACK_INTEGRITY_ACTIVATION_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch         = "ROLLBACK_INTEGRITY_ACTIVATION_CAPSULE_DIGEST_MISMATCH"
	ReasonAssuranceMismatch      = "ROLLBACK_INTEGRITY_ACTIVATION_ASSURANCE_MISMATCH"
	ReasonEligibilityNotExact    = "ROLLBACK_INTEGRITY_ACTIVATION_ELIGIBILITY_NOT_EXACT"
	ReasonEligibilityUnknown     = "ROLLBACK_INTEGRITY_ACTIVATION_ELIGIBILITY_UNKNOWN"
	AssuranceCapsuleHash         = "sha256:e57279b58a03b90c8dad1f4f86b3f17023cf217c4c63fdfd9fd2f1a8e8d1fae8"
	EligibilityCapsuleHash       = "sha256:29059e16b53b18177e391796bb6d9cffa85f2872f22722f5e7a387b1b16d51bc"
	EligibilityReportHash        = "sha256:0354624c036750dd84dd1fb2f49969f4cf9e94ab5ea609c1dbf29b8443be12a4"
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
