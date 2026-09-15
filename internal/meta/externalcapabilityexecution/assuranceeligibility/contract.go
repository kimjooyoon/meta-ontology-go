package assuranceeligibility

const (
	ReportSchema     = "gooo/external-conformance-eligibility/v1"
	SuiteSchema      = "gooo/external-conformance-eligibility-conformance/v1"
	SuiteDenominator = "gooo/external-conformance-eligibility-denominator/v1"
	MetricID         = "gooo.metric.ecosystem.external-conformance.v1"
	MetaOperation    = "qualify-external-conformance"

	DecisionEligible    = "ELIGIBLE_SHADOW"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionUnknown   = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNone          = "NO_EFFECT"
	EffectBlock         = "BLOCK"

	ReasonEligible           = "EXTERNAL_CONFORMANCE_ASSURANCE_ELIGIBLE_SHADOW"
	ReasonUnavailable        = "EXTERNAL_CONFORMANCE_ELIGIBILITY_EVIDENCE_UNAVAILABLE"
	ReasonSubjectUnknown     = "EXTERNAL_CONFORMANCE_ELIGIBILITY_SUBJECT_UNKNOWN"
	ReasonMalformed          = "EXTERNAL_CONFORMANCE_ELIGIBILITY_EVIDENCE_MALFORMED"
	ReasonDecisionUnknown    = "EXTERNAL_CONFORMANCE_ELIGIBILITY_DECISION_UNKNOWN"
	ReasonDigestMismatch     = "EXTERNAL_CONFORMANCE_ELIGIBILITY_DIGEST_MISMATCH"
	ReasonAssuranceMismatch  = "EXTERNAL_CONFORMANCE_ASSURANCE_BASELINE_MISMATCH"
	ReasonSubjectMismatch    = "EXTERNAL_CONFORMANCE_ELIGIBILITY_SUBJECT_MISMATCH"
	ReasonReferenceMismatch  = "EXTERNAL_CONFORMANCE_REFERENCE_MISMATCH"
	ReasonParentMismatch     = "EXTERNAL_CONFORMANCE_PARENT_BOUNDARY_MISMATCH"
	ReasonCapabilityMismatch = "EXTERNAL_CONFORMANCE_CAPABILITY_MISMATCH"
	ReasonWriteObserved      = "EXTERNAL_CONFORMANCE_ELIGIBILITY_WRITE_OBSERVED"
	ReasonMutationObserved   = "EXTERNAL_CONFORMANCE_ELIGIBILITY_MUTATION_OBSERVED"
	ReasonPromotionObserved  = "EXTERNAL_CONFORMANCE_ELIGIBILITY_PROMOTION_OBSERVED"

	AssuranceName             = "language-assurance"
	ParentReportName          = "external-parent-report"
	ParentObservationName     = "external-parent-observation"
	ParentSuiteName           = "external-parent-suite"
	CapabilityReportName      = "external-capability-report"
	CapabilityObservationName = "external-capability-observation"
	CapabilitySuiteName       = "external-capability-suite"

	AssuranceSubject     = "6e9a65b48bfa734968813014532203c15a5d0e3f"
	AssuranceDigest      = "sha256:13669bdc1e871f659cf00d0659fa179f94bdcbe47531a86db793dc0b2edc809b"
	AssuranceReport      = "sha256:31d1b838c9fe5504569aa7e8c81782317ff4ba9ad6231a44e48e42d5ba1827e4"
	AssuranceDenominator = "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02"
	ParentDenominator    = "sha256:c90c65ee1d4867a394a776784dac3f9c44d63a4e2918e6dad1999f1a2c410d9d"

	ReferenceURL    = "https://github.com/cosmos72/gomacro"
	ReferenceCommit = "cf0d4bf32da393dbda97e3572f216731013ffa55"
	ReferenceTree   = "8cc240a53dd29432ad83620b20fd8a0a05674c6d"
	GoVersion       = "go1.27.0"
)

var artifactNames = []string{AssuranceName, ParentReportName, ParentObservationName,
	ParentSuiteName, CapabilityReportName, CapabilityObservationName, CapabilitySuiteName}
