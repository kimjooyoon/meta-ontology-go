package artifact

import (
	"encoding/json"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
)

func BuildWithPromotionEvidence(
	conceptArtifact, promotionRaw, capabilityRaw []byte, headSHA string,
) (Receipt, error) {
	promotion := proposalpromotion.Receipt{}
	if err := json.Unmarshal(promotionRaw, &promotion); err != nil {
		return Receipt{}, err
	}
	capability := guardedcapability.Receipt{}
	if err := json.Unmarshal(capabilityRaw, &capability); err != nil {
		return Receipt{}, err
	}
	snapshot, err := readiness.EvaluateWithPromotionEvidence(
		conceptArtifact, promotion, capability, headSHA,
	)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, promotion.ReportDigest, capability.ReportDigest)
}
