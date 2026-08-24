package candidateleakageeligibility

const (
	ReportSchema  = "gooo/candidate-leakage-eligibility/v1"
	SuiteSchema   = "gooo/candidate-leakage-eligibility-conformance/v1"
	DenominatorID = "gooo/candidate-leakage-eligibility-denominator/v1"
	MetricID      = "gooo.metric.semantic.candidate-leakage.v1"
	MetaOperation = "detect-candidate-leakage"

	DecisionEligible   = "ELIGIBLE"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
	ResolutionUnknown  = "UNKNOWN"
	ResolutionInvariant = "INVARIANT_ONLY"
	EffectNone          = "NO_EFFECT"
	EffectBlock         = "BLOCK"

	ReasonEligible        = "CANDIDATE_LEAKAGE_PROMOTION_ELIGIBLE"
	ReasonUnavailable     = "CANDIDATE_LEAKAGE_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch  = "CANDIDATE_LEAKAGE_CAPSULE_DIGEST_MISMATCH"
	ReasonSemanticMismatch = "CANDIDATE_LEAKAGE_CAPSULE_SEMANTIC_MISMATCH"
	ReasonSubjectUnknown   = "CANDIDATE_LEAKAGE_SUBJECT_UNKNOWN"

	EvidenceSubjectSHA = "466c12d94f657f2e04f355919788260c17faa678"
	AssuranceArtifactID int64 = 9508504894
	AssuranceArtifactDigest = "sha256:362d11847f9b10e781a78e990feb650d147935c386851d5d4078c617288af760"
	AssuranceCapsuleDigest = "sha256:449e4386e921ae33bff97cca4e562527f2ccf33f53d042d43ebe388911fe5cc4"
	ShadowArtifactID int64 = 9508504114
	ShadowArtifactDigest = "sha256:67d1c38717b4d38c07482ea9fcba45ce0ca9cb3e8a3bf6f4ff4f2e9df5100857"
	ShadowCapsuleDigest = "sha256:dc7dbed9a3b89e301c5ed210afd6de23045e7e73f2177f88a2857f362ebbbf70"
)

type Capsule struct {
	Name           string `json:"name"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	Payload        []byte `json:"-"`
}

type Input struct {
	SubjectSHA string  `json:"subject_sha"`
	Assurance  Capsule `json:"assurance"`
	Shadow     Capsule `json:"shadow"`
}

type MetaOperationBinding struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}

func MetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{ID: "consume-merged-candidate-shadow", ProofChoice: "FOUNDATION"},
		{ID: "bind-candidate-eligibility-baseline", ProofChoice: "FOUNDATION"},
		{ID: "evaluate-candidate-leakage-eligibility", ProofChoice: "COHERENCE"},
		{ID: "preserve-eligibility-unknown", ProofChoice: "REGRESSION"},
		{ID: "deny-eligibility-side-effects", ProofChoice: "COHERENCE"},
	}
}
