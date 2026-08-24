package verticalsliceclosureeligibility

const (
	ReportSchema      = "gooo/vertical-slice-closure-eligibility/v1"
	SuiteSchema       = "gooo/vertical-slice-closure-eligibility-conformance/v1"
	SuiteDenominator  = "gooo/vertical-slice-closure-eligibility-denominator/v1"
	MetricID          = "gooo.metric.capability.vertical-slice-closure.v1"
	MetaOperation     = "qualify-vertical-slice-closure"
	DecisionEligible  = "ELIGIBLE"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact   = "EXACT"
	ResolutionUnknown = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNone        = "NO_EFFECT"
	EffectBlock       = "BLOCK"

	ReasonEligible       = "VERTICAL_SLICE_CLOSURE_PROMOTION_ELIGIBLE"
	ReasonUnavailable    = "VERTICAL_SLICE_ELIGIBILITY_EVIDENCE_UNAVAILABLE"
	ReasonSubjectUnknown = "VERTICAL_SLICE_ELIGIBILITY_SUBJECT_UNKNOWN"
	ReasonDigestMismatch = "VERTICAL_SLICE_ELIGIBILITY_CAPSULE_DIGEST_MISMATCH"
	ReasonDecisionUnknown = "VERTICAL_SLICE_ELIGIBILITY_DECISION_UNKNOWN"
	ReasonLinkMismatch   = "VERTICAL_SLICE_ELIGIBILITY_LINK_MISMATCH"
	ReasonWriteObserved  = "VERTICAL_SLICE_ELIGIBILITY_WRITE_OBSERVED"
	ReasonSemanticMismatch = "VERTICAL_SLICE_ELIGIBILITY_SEMANTIC_MISMATCH"

	AssuranceName = "language-assurance"
	ShadowName    = "vertical-slice-closure-shadow"
	AssuranceEvidenceSubject = "145b81c8bb8e4b1eb46cb10af0ea21a6b6be51b5"
	ShadowEvidenceHead       = "5a6f3535ed6984fa8ed4bd9806638f246d9c0263"
	AssuranceDenominatorID   = "gooo/language-assurance-denominator/v1"
	AssuranceDenominatorDigest = "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02"
	ShadowDenominatorDigest  = "sha256:6b4b3793133313d430d4c53792baf04cb17fc0fd9ac5592aeb33bc01c0ad6962"
	AssuranceReportDigest    = "sha256:9b27f001edc971ad39958ef5ef7f94293c73c8727be66390df0ee8b3a60d703e"
	ShadowReportDigest       = "sha256:fdc3a0c735999a70b7b7fa8663756659b50f89ab63384e9cafe082a166536840"
	AssuranceArchiveDigest   = "sha256:48663ffb5d413de23b6efc90cb08229d7c56cef4798a0545aa89510a3f21781f"
	AssuranceCapsuleDigest   = "sha256:13581ebf64e0e3a512d1e8b3ca05de05e14d4453b64f3c7eff8e3b854a89d969"
	ShadowArchiveDigest      = "sha256:25852404b2b1a8cba987bddaeede9d852ff11e7cff18aba1988bd8563641d6d8"
	ShadowCapsuleDigest      = "sha256:a3aa917fbc47a2485f5252497ce8f83a6dfbf607fc6e0f923e1ac45551d8ffea"
	AssuranceArtifactID int64 = 9518495029
	ShadowArtifactID    int64 = 9520815531
)

type Capsule struct {
	Name string `json:"name"`
	ArtifactID int64 `json:"artifact_id"`
	ArchiveDigest string `json:"archive_digest"`
	CapsuleDigest string `json:"capsule_digest"`
	Payload []byte `json:"-"`
}

type Input struct {
	SubjectSHA string `json:"subject_sha"`
	Assurance Capsule `json:"assurance"`
	Shadow Capsule `json:"shadow"`
}

type MetaOperationBinding struct {
	ID string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}

func MetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{ID: "consume-fixed-assurance", ProofChoice: "FOUNDATION"},
		{ID: "consume-merged-vertical-shadow", ProofChoice: "FOUNDATION"},
		{ID: MetaOperation, ProofChoice: "COHERENCE"},
		{ID: "preserve-unknown-decision", ProofChoice: "REGRESSION"},
		{ID: "preserve-read-only-eligibility", ProofChoice: "FOUNDATION"},
		{ID: "deny-eligibility-side-effects", ProofChoice: "COHERENCE"},
	}
}
