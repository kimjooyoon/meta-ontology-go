package languagereadiness

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
)

const autonomousProposalConcept = "autonomous-change-proposal"

type promotionEvidence struct {
	Concept         conceptEvidence `json:"concept"`
	PromotionDigest string          `json:"promotion_digest"`
}

func EvaluateWithProposalPromotion(
	raw []byte, promotion proposalpromotion.Receipt, expectedHeadSHA string,
) (Snapshot, error) {
	if err := proposalpromotion.Validate(promotion, expectedHeadSHA); err != nil {
		return Snapshot{}, fmt.Errorf("verify autonomous proposal promotion: %w", err)
	}
	if promotion.Decision != proposalpromotion.DecisionPass {
		return Snapshot{}, fmt.Errorf(
			"FAIL_CLOSED: autonomous proposal promotion decision %q", promotion.Decision,
		)
	}
	return evaluate(raw, promotion.ReportDigest)
}
