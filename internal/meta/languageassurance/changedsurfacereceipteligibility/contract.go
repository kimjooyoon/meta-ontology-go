package changedsurfacereceipteligibility

const (
	ReportSchema  = "gooo/changed-surface-receipt-eligibility/v1"
	SuiteSchema   = "gooo/changed-surface-receipt-eligibility-conformance/v1"
	DenominatorID = "gooo/changed-surface-receipt-eligibility-denominator/v1"
	MetricID      = "gooo.metric.semantic.changed-surface-receipt-totality.v1"
	MetaOperation = "totalize-changed-surface-receipts"

	DecisionEligible    = "ELIGIBLE"
	DecisionFailClosed  = "FAIL_CLOSED"
	ResolutionExact     = "EXACT"
	ResolutionUnknown   = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNone          = "NO_EFFECT"
	EffectBlock         = "BLOCK"

	ReasonEligible         = "CHANGED_SURFACE_RECEIPT_PROMOTION_ELIGIBLE"
	ReasonUnavailable      = "CHANGED_SURFACE_RECEIPT_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch   = "CHANGED_SURFACE_RECEIPT_CAPSULE_DIGEST_MISMATCH"
	ReasonSemanticMismatch = "CHANGED_SURFACE_RECEIPT_CAPSULE_SEMANTIC_MISMATCH"
	ReasonSubjectUnknown   = "CHANGED_SURFACE_RECEIPT_SUBJECT_UNKNOWN"

	EvidenceSubjectSHA              = "25ee2f076d67eafceec4eac3ff4d85a90abb597f"
	AssuranceArtifactID       int64 = 9511257735
	AssuranceArtifactDigest         = "sha256:c1da86e6c384b903fd1d5ecf3b17489d28cc89670d630a753a99aded96b52653"
	AssuranceCapsuleDigest          = "sha256:a086d68d8fd4ed493d4e8ce199c620b99f58d5c713e34325edfd6f1384476202"
	ShadowArtifactID          int64 = 9511274498
	ShadowArtifactDigest            = "sha256:9cf5e762dcf353810647cee60947dd0fcfe315539776b80aef48f69464389d6e"
	ShadowReportCapsuleDigest       = "sha256:ba25d0ee3ee0943bbb7c345040a1178626c511aa73ff08df25284677033e09d3"
	ShadowSuiteCapsuleDigest        = "sha256:b50e515649ab22a047f7c2d75009c4a43aa98430fa250ab69c999957fcc24f26"
)

type Capsule struct {
	Name           string `json:"name"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	Payload        []byte `json:"-"`
}

type Input struct {
	SubjectSHA   string  `json:"subject_sha"`
	Assurance    Capsule `json:"assurance"`
	ShadowReport Capsule `json:"shadow_report"`
	ShadowSuite  Capsule `json:"shadow_suite"`
}

type MetaOperationBinding struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}
