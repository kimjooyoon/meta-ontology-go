package artifact

import (
	"encoding/json"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
)

func Build(conceptArtifact []byte, headSHA string) (Receipt, error) {
	snapshot, err := readiness.Evaluate(conceptArtifact)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, "", "")
}

func BuildWithProposalPromotion(
	conceptArtifact, promotionRaw []byte, headSHA string,
) (Receipt, error) {
	promotion := proposalpromotion.Receipt{}
	if err := json.Unmarshal(promotionRaw, &promotion); err != nil {
		return Receipt{}, err
	}
	snapshot, err := readiness.EvaluateWithProposalPromotion(
		conceptArtifact, promotion, headSHA,
	)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, promotion.ReportDigest, "")
}

func build(snapshot readiness.Snapshot, headSHA, promotionDigest, capabilityDigest string) (Receipt, error) {
	input := improvement.FromReadiness(snapshot)
	receipt := seal(Receipt{
		Schema:                  Schema,
		HeadSHA:                 headSHA,
		ProposalPromotionDigest: promotionDigest,
		GuardedCapabilityDigest: capabilityDigest,
		Snapshot:                snapshot,
		TransitionInput:         input,
		FixedPoint:              improvement.Evaluate(input, input),
	})
	if err := Validate(receipt); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}
