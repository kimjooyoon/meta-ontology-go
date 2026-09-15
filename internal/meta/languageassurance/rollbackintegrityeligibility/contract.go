package rollbackintegrityeligibility

const (
	ReportSchema  = "gooo/rollback-integrity-eligibility/v1"
	SuiteSchema   = "gooo/rollback-integrity-eligibility-conformance/v1"
	DenominatorID = "gooo/rollback-integrity-eligibility-denominator/v1"
	MetricID      = "gooo.metric.operation.rollback-integrity.v1"
	MetaOperation = "verify-rollback-integrity"

	DecisionEligible    = "ELIGIBLE"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionUnknown   = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNone          = "NO_EFFECT"
	EffectBlock         = "BLOCK"

	ReasonEligible         = "ROLLBACK_INTEGRITY_PROMOTION_ELIGIBLE"
	ReasonUnavailable      = "ROLLBACK_INTEGRITY_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch   = "ROLLBACK_INTEGRITY_CAPSULE_DIGEST_MISMATCH"
	ReasonSemanticMismatch = "ROLLBACK_INTEGRITY_CAPSULE_SEMANTIC_MISMATCH"
	ReasonSubjectUnknown   = "ROLLBACK_INTEGRITY_SUBJECT_UNKNOWN"

	EvidenceSubjectSHA               = "1f27d342faf7a435ca4c534e2a816f29befe21a4"
	ShadowAssuranceSubjectSHA        = "4b8758428951e3e208b42cb568cc208cbe83dea9"
	AssuranceArtifactID        int64 = 9514168645
	AssuranceArtifactDigest          = "sha256:22295324edb942b0f910754c2f5ec3651613cc5d1e760d59d40009b98e8bc25b"
	AssuranceCapsuleDigest           = "sha256:5a970af60410c89b7949f0604288a8227e0bbdc53a0999bc8549816ead4841f8"
	ShadowArtifactID           int64 = 9514191722
	ShadowArtifactDigest             = "sha256:f2840da5805b7690145be917749f2933af24e050e67059814750e546f5a534ef"
	ShadowReportACapsuleDigest       = "sha256:a2fae67f1d4e842b188cd11f2fe87b1647c557ce8a8a07529ac34aa3cbfba105"
	ShadowReportBCapsuleDigest       = "sha256:a2fae67f1d4e842b188cd11f2fe87b1647c557ce8a8a07529ac34aa3cbfba105"
	ShadowEvidenceDigest             = "sha256:4641750a0d3299524088d4b56678d594ee2873c0a13eaea25332709cfd9471ac"
	ShadowReportDigest               = "sha256:430d36c124bf1d14b07615871b94236d03a264fc90282d3d4a406fc5d33ed40c"
)

type Capsule struct {
	Name           string `json:"name"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	Payload        []byte `json:"-"`
}

type Input struct {
	SubjectSHA    string  `json:"subject_sha"`
	Assurance     Capsule `json:"assurance"`
	ShadowReportA Capsule `json:"shadow_report_a"`
	ShadowReportB Capsule `json:"shadow_report_b"`
}

type MetaOperationBinding struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}
