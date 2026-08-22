package languagereadiness

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
)

func EvaluateWithProposalPromotion(
	raw []byte, promotion proposalpromotion.Receipt, expectedHeadSHA string,
) (Snapshot, error) {
	digest, err := validateProposalPromotion(promotion, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	return evaluate(raw, evidenceDigests{proposal: digest})
}

func EvaluateWithPromotionEvidence(raw []byte, promotion proposalpromotion.Receipt,
	capability guardedcapability.Receipt, expectedHeadSHA string) (Snapshot, error) {
	promotionDigest, err := validateProposalPromotion(promotion, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	if err := guardedcapability.ValidateForHead(capability, expectedHeadSHA); err != nil {
		return Snapshot{}, fmt.Errorf("verify guarded promotion capability: %w", err)
	}
	if capability.Decision != guardedcapability.DecisionPass {
		return Snapshot{}, fmt.Errorf("FAIL_CLOSED: guarded capability decision %q", capability.Decision)
	}
	return evaluate(raw, evidenceDigests{
		proposal: promotionDigest, guarded: capability.ReportDigest,
	})
}

func validateProposalPromotion(
	promotion proposalpromotion.Receipt, expectedHeadSHA string,
) (string, error) {
	if err := proposalpromotion.Validate(promotion, expectedHeadSHA); err != nil {
		return "", fmt.Errorf("verify autonomous proposal promotion: %w", err)
	}
	if promotion.Decision != proposalpromotion.DecisionPass {
		return "", fmt.Errorf(
			"FAIL_CLOSED: autonomous proposal promotion decision %q", promotion.Decision,
		)
	}
	return promotion.ReportDigest, nil
}
