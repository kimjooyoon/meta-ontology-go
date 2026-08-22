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

	FoundationSubjectSHA     = "027785b635246537354ac7517054b755235ae765"
	FoundationWorkflowRunID  = int64(32599122103)
	FoundationArtifactID     = int64(9482381929)
	FoundationArtifactDigest = "sha256:6fe9223adb6260e644bba5cfb28247291623ba139c4773bb15da139ae3da63f0"
	FoundationReportFileSHA  = "sha256:4aefc3ff9d9ce4af323511fa77eee32bf489a2ca58efde27cdd269d945209408"
	FoundationReportDigest   = "sha256:5a1d897ea91843a6f52df95641720fa7d2a09e0ce810487bcddde9097154b959"
)

const (
	guardPath   = "internal/meta/languagereadiness/guardedpromotion"
	witnessPath = "cmd/guarded-promotion-witness"
)
