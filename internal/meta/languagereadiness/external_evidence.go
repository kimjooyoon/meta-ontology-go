package languagereadiness

const (
	autonomousProposalConcept = "autonomous-change-proposal"
	guardedPromotionConcept   = "guarded-exact-promotion"
	toolchainUseCasesConcept  = "toolchain-executable-use-cases"
)

type evidenceDigests struct {
	proposal string
	guarded  string
	useCases string
}

type externalEvidence struct {
	Concept conceptEvidence `json:"concept"`
	Digest  string          `json:"digest"`
}

func requiredEvidence(conceptID string, evidence evidenceDigests) (string, string, bool) {
	switch conceptID {
	case autonomousProposalConcept:
		return evidence.proposal, "VERIFIED_PROMOTION_RECEIPT_REQUIRED", true
	case guardedPromotionConcept:
		return evidence.guarded, "GUARDED_CAPABILITY_RECEIPT_REQUIRED", true
	case toolchainUseCasesConcept:
		return evidence.useCases, "EXECUTABLE_USE_CASE_RECEIPT_REQUIRED", true
	default:
		return "", "", false
	}
}
