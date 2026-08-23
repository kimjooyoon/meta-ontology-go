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

	FoundationSubjectSHA     = "547f13fbcc533ea3ff5b90340bcb0c320f61a475"
	FoundationWorkflowRunID  = int64(32667794738)
	FoundationArtifactID     = int64(9500514949)
	FoundationArtifactDigest = "sha256:2c4f65b146af7dddf223f75b53d09f2f505d20a86a7b29fcf04b3db20bcd938d"
	FoundationReportFileSHA  = "sha256:2a1859d7766178d94085488aa1a5490922407378e3a03405f2f34a970622027a"
	FoundationReportDigest   = "sha256:9ebed6830b873a35074e37ebab3cc92ed67eef672680565b4627605e06db9435c"
)

const (
	guardPath   = "internal/meta/languagereadiness/guardedpromotion"
	witnessPath = "cmd/guarded-promotion-witness"
)
