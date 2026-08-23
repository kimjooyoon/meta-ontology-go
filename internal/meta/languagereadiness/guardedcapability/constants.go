package guardedcapability

const (
	Schema = "gooo/language-guarded-promotion-capability/v1"

	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
	ResolutionLower    = "LOWER_RESOLUTION"
	ReasonExact        = "GUARDED_PROMOTION_CAPABILITY_EXACT"
	ReasonRejected     = "GUARDED_PROMOTION_CAPABILITY_REJECTED"
	ReasonUnknown      = "GUARDED_PROMOTION_CAPABILITY_EVIDENCE_UNKNOWN"

	FoundationSubjectSHA     = "d9960ae95ffdc66179de0a1be13364aefeab76ea"
	FoundationWorkflowRunID  = int64(32670602811)
	FoundationArtifactID     = int64(9501263129)
	FoundationArtifactDigest = "sha256:178679f9ed4db0c844edce5ed2103a2cf8a2c59f0ebb3d5c06b6cd051a66d92f"
	FoundationReportFileSHA  = "sha256:7f379d7eea9875aff3657b3bc77a039da98676cb87e504f25b3957339be2803b"
	FoundationReportDigest   = "sha256:67cdb6610b00f9533c3db462804e67e84d1c037a7625e3bef562a82b515d96bd"
)

const (
	guardPath   = "internal/meta/languagereadiness/guardedpromotion"
	witnessPath = "cmd/guarded-promotion-witness"
)
