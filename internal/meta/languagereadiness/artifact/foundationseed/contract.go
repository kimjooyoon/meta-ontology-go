package foundationseed

const (
	Schema             = "gooo/language-readiness-foundation-seed/v1"
	DecisionAuthorized = "FOUNDATION_SEED_AUTHORIZED"
	DecisionFailClosed = "FAIL_CLOSED"
	ReasonExact        = "READINESS_FOUNDATION_EXHAUSTION_EXACT"
	ReasonUnknown      = "READINESS_FOUNDATION_EVIDENCE_UNKNOWN"
	ResolutionExact    = "EXACT_EXHAUSTED_ANCESTRY"
	ResolutionLower    = "LOWER_RESOLUTION"
	IndicatorCount     = 12
	canonicalBranch    = "dev"
	canonicalWorkflow  = "Transformation effect ledger"
)

var fixedNonClaims = []string{
	"language readiness improvement",
	"resolved readiness predecessor",
	"repository mutation",
	"promotion or automatic adoption",
	"foundation eligibility after a canonical predecessor exists",
}
