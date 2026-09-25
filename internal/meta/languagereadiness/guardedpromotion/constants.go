package guardedpromotion

const (
	Schema = "gooo/autonomy-guarded-promotion/v1"

	DecisionAuthorized = "AUTHORIZED"
	DecisionDenied     = "DENIED"
	DecisionFailClosed = "FAIL_CLOSED"

	ReasonAuthorized         = "MERGED_PUSH_PROMOTION_AUTHORIZED"
	ReasonMergedPushRequired = "MERGED_PUSH_REQUIRED"
	ReasonGuardrailRejected  = "GUARDED_PROMOTION_REJECTED"
	ReasonEvidenceUnknown    = "GUARDED_PROMOTION_EVIDENCE_UNKNOWN"
	ReasonRepositoryMismatch = "REPOSITORY_IDENTITY_MISMATCH"

	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER_RESOLUTION"

	PromotionSchema       = "gooo/autonomous-change-proposal-promotion/v1"
	PromotionArtifactBase = "language-readiness-proposal-promotion-"
	TransformationPath    = ".github/workflows/metric-counterfactual.yml"
	CIPath                = ".github/workflows/ci.yml"
)
