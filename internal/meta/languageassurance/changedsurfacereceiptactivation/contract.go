package changedsurfacereceiptactivation

const (
	Schema                    = "gooo/changed-surface-receipt-activation/v1"
	DenominatorID             = "gooo/changed-surface-receipt-activation-denominator/v1"
	PredecessorSHA            = "1b9acfaac0ff2d1a2353de6c5019d515ab542eb3"
	EvidenceSubjectSHA        = "25ee2f076d67eafceec4eac3ff4d85a90abb597f"
	MetricID                  = "gooo.metric.semantic.changed-surface-receipt-totality.v1"
	MetaOperation             = "totalize-changed-surface-receipts"
	EligibilityDenominatorID  = "gooo/changed-surface-receipt-eligibility-denominator/v1"
	DecisionApplied           = "APPLIED"
	DecisionFailClosed        = "FAIL_CLOSED"
	ResolutionExact           = "EXACT"
	ResolutionUnknown         = "UNKNOWN"
	ResolutionInvariant       = "INVARIANT_ONLY"
	ReasonApplied             = "CHANGED_SURFACE_RECEIPT_ACTIVATED"
	ReasonUnavailable         = "CHANGED_SURFACE_RECEIPT_ACTIVATION_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch      = "CHANGED_SURFACE_RECEIPT_ACTIVATION_CAPSULE_DIGEST_MISMATCH"
	ReasonAssuranceMismatch   = "CHANGED_SURFACE_RECEIPT_ACTIVATION_ASSURANCE_MISMATCH"
	ReasonEligibilityNotExact = "CHANGED_SURFACE_RECEIPT_ACTIVATION_ELIGIBILITY_NOT_EXACT"
	AssuranceCapsuleHash      = "sha256:6cfdc133e9797483b88a368fc0e5eeb73659d0731a5438ae6e4fffdf12730cc6"
	EligibilityCapsuleHash    = "sha256:53ee343de4b5116b6a4810132d5b6b668b56e4ebbd23516b1549a7651f4acb23"
	EligibilityReportHash     = "sha256:96317cd87644ca478e918d5e20690aff57de8a9d52c0f1ee7c5c3463be0ab61e"
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
