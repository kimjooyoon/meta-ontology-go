package externalconformanceactivation

const (
	Schema                    = "gooo/external-conformance-activation/v1"
	DenominatorID             = "gooo/external-conformance-activation-denominator/v1"
	PredecessorSHA            = "2104e3fcede21572951c266cd7b61fd38f386aef"
	EligibilitySubjectSHA     = "0fa760bdcb9c90e386d85978bff1c6552e8b503a"
	EligibilityAssuranceSHA   = "6e9a65b48bfa734968813014532203c15a5d0e3f"
	MetricID                  = "gooo.metric.ecosystem.external-conformance.v1"
	MetaOperation             = "verify-external-conformance"
	EligibilityMetaOperation  = "qualify-external-conformance"
	DecisionApplied           = "APPLIED"
	DecisionFailClosed        = "FAIL_CLOSED"
	ResolutionExact           = "EXACT"
	ResolutionUnknown         = "UNKNOWN"
	ResolutionInvariant       = "INVARIANT_ONLY"
	EffectApply               = "APPLY_TRANSITION"
	EffectBlock               = "BLOCK"
	ReasonApplied             = "EXTERNAL_CONFORMANCE_ASSURANCE_ACTIVATED"
	ReasonUnavailable         = "EXTERNAL_CONFORMANCE_ACTIVATION_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch      = "EXTERNAL_CONFORMANCE_ACTIVATION_CAPSULE_DIGEST_MISMATCH"
	ReasonAssuranceMismatch   = "EXTERNAL_CONFORMANCE_ACTIVATION_ASSURANCE_MISMATCH"
	ReasonEligibilityMismatch = "EXTERNAL_CONFORMANCE_ACTIVATION_ELIGIBILITY_MISMATCH"
	ReasonMergeMismatch       = "EXTERNAL_CONFORMANCE_ACTIVATION_MERGE_MISMATCH"
	ReasonAssuranceUnknown    = "EXTERNAL_CONFORMANCE_ACTIVATION_ASSURANCE_DECISION_UNKNOWN"
	ReasonEligibilityUnknown  = "EXTERNAL_CONFORMANCE_ACTIVATION_ELIGIBILITY_DECISION_UNKNOWN"
	ReasonMergeUnknown        = "EXTERNAL_CONFORMANCE_ACTIVATION_MERGE_STATE_UNKNOWN"
)

const (
	AssuranceCapsuleHash   = "sha256:b6d03b8905bc9d0ac2d1773d0b3dd273743244fa9366ed33bcb4e5d4c5e363ec"
	EligibilityCapsuleHash = "sha256:f5e2d8f35c3f052adad9edc8e05612a3e0aa13f1c336e8e9bc86dafb3b675c10"
	MergeCapsuleHash       = "sha256:2972619d7bc2d42fc24e5b6557dfb781e3c84aa836e0dbe9b6c9f4e26be13295"
	AssuranceReportHash    = "sha256:9cf5725a2071f83fede528d2fc5bb45d3eaa0dd35bb43bc4622da64a7fe9ca42"
	EligibilityReportHash  = "sha256:2fc3ccd5b9022d947dcb92281b13ca8363ea98106ca3a488319e5d02d091d888"
)

type CaseSpec struct {
	ID, ExpectedDecision, ExpectedResolution, ExpectedReason string
}

func Denominator() []CaseSpec {
	return []CaseSpec{
		{"exact", DecisionApplied, ResolutionExact, ReasonApplied},
		{"unavailable", DecisionFailClosed, ResolutionUnknown, ReasonUnavailable},
		{"digest-mismatch", DecisionFailClosed, ResolutionInvariant, ReasonDigestMismatch},
		{"eligibility-unknown", DecisionFailClosed, ResolutionUnknown, ReasonEligibilityUnknown},
		{"eligibility-fixed-point", DecisionFailClosed, ResolutionUnknown, ReasonEligibilityUnknown},
		{"eligibility-unrecognized", DecisionFailClosed, ResolutionUnknown, ReasonEligibilityUnknown},
		{"assurance-unknown", DecisionFailClosed, ResolutionUnknown, ReasonAssuranceUnknown},
		{"merge-unknown", DecisionFailClosed, ResolutionUnknown, ReasonMergeUnknown},
	}
}
